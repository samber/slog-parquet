package slogparquet

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"github.com/parquet-go/parquet-go"
	"github.com/samber/lo"
	"github.com/thanos-io/objstore"
)

type ParquetBuffer interface {
	Append(time.Time, slog.Level, string, map[string]any) error
	Flush(bool) error
}

type parquetBuffer struct {
	prefix string
	id     string

	maxRecords   int
	maxInterval  time.Duration
	rowGroupSize int

	mutex  sync.Mutex // ☢️
	bucket objstore.Bucket

	start  time.Time
	buffer *parquet.GenericBuffer[log]
}

func NewParquetBuffer(bucket objstore.Bucket, prefix string, maxRecords int, maxInterval time.Duration) ParquetBuffer {
	return &parquetBuffer{
		prefix: prefix,
		id:     lo.Must(uuid.NewV4()).String()[0:5],

		maxRecords:   maxRecords,
		maxInterval:  maxInterval,
		rowGroupSize: int(math.Ceil(float64(maxRecords) / 10)),

		mutex:  sync.Mutex{},
		bucket: bucket,

		start:  time.Now(),
		buffer: resetBuffer(),
	}
}

func (b *parquetBuffer) Append(tIme time.Time, logLevel slog.Level, message string, attributes map[string]any) error {
	// bearer:disable go_lang_deserialization_of_user_input
	serializedAttrs, err := json.Marshal(attributes)
	if err != nil {
		return err
	}

	b.mutex.Lock()

	_, err = b.buffer.Write([]log{
		{
			Time:       tIme,
			LogLevel:   logLevel.String(),
			Message:    message,
			Attributes: serializedAttrs,
			Source:     library,
		},
	})
	if err != nil {
		return err
	}

	if b.buffer.Len() >= b.maxRecords || b.start.Add(b.maxInterval).Before(time.Now()) {
		b.mutex.Unlock()
		return b.Flush(false)
	}

	b.mutex.Unlock()
	return nil
}

func (b *parquetBuffer) Flush(sync bool) error {
	if sync {
		return b.flush()
	}

	go b.flush()
	return nil
}

func (b *parquetBuffer) flush() error {
	b.mutex.Lock()
	buffer := b.buffer
	b.start = time.Now()
	b.buffer = resetBuffer()
	b.mutex.Unlock()

	if buffer.Len() == 0 {
		return nil
	}

	sort.Sort(buffer)

	buff := bytes.NewBuffer(nil)
	writer := parquet.NewGenericWriter[log](
		buff,
		parquet.CreatedBy(goModulePath, semver, buildSHA1),
		parquet.Compression(&parquet.Lz4Raw),
		parquet.MaxRowsPerRowGroup(int64(b.rowGroupSize)),
		parquet.BloomFilters(
			parquet.SplitBlockFilter(4, "log_level"),
		),
	)

	_, err := parquet.CopyRows(writer, buffer.Rows())
	if err != nil {
		return err
	}

	err = writer.Close()
	if err != nil {
		return err
	}

	err = writer.Flush()
	if err != nil {
		return err
	}

	key := fmt.Sprintf(
		"%s/%s/%s.%s.parquet",
		b.prefix,
		b.start.UTC().Format("2006-01-02"),
		b.start.UTC().Format("15:04:05"),
		b.id,
	)

	return b.bucket.Upload(
		context.Background(),
		key,
		buff,
	)
}

func resetBuffer() *parquet.GenericBuffer[log] {
	return parquet.NewGenericBuffer[log](
		parquet.SchemaOf(new(log)),
		parquet.ColumnBufferCapacity(16384),
		parquet.SortingRowGroupConfig(
			parquet.SortingColumns(
				parquet.Ascending("time"),
			),
		),
	)
}
