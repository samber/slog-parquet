package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"log/slog"

	slogparquet "github.com/samber/slog-parquet/v2"
	"github.com/thanos-io/objstore/providers/s3"
)

func main() {
	// export AWS_ACCESS_KEY=helloworld
	// export AWS_SECRET_KEY=helloworld
	// export AWS_S3_REGION=fr-par
	// export AWS_S3_ENDPOINT=localhost:9000
	// export AWS_S3_BUCKET=slog-test
	// go run *.go

	bucket, err := s3.NewBucketWithConfig(
		slogparquet.NewLogger(),
		s3.Config{
			Endpoint:  os.Getenv("AWS_S3_ENDPOINT"),
			Insecure:  true,
			Region:    os.Getenv("AWS_S3_REGION"),
			AccessKey: os.Getenv("AWS_ACCESS_KEY"),
			SecretKey: os.Getenv("AWS_SECRET_KEY"),
			Bucket:    os.Getenv("AWS_S3_BUCKET"),
			PartSize:  16 * 1024 * 1024, // 16MB
		},
		"samber/slog-parquet",
	)
	if err != nil {
		log.Fatal(err)
	}

	buffer := slogparquet.NewParquetBuffer(bucket, "logs", 10*1024*1024, 5*time.Second)

	logger := slog.New(slogparquet.Option{Level: slog.LevelDebug, Buffer: buffer}.NewParquetHandler())
	logger = logger.
		With("environment", "dev").
		With("release", "v1.0.0")

	// log error
	logger.
		With("category", "sql").
		With("query.statement", "SELECT COUNT(*) FROM users;").
		With("query.duration", 1*time.Second).
		With("error", fmt.Errorf("could not count users")).
		Error("caramba!")

	// log user signup
	logger.
		With(
			slog.Group("user",
				slog.String("id", "user-123"),
				slog.Time("created_at", time.Now()),
			),
		).
		Info("user registration")

	for i := 0; i < 10_000_000; i++ {
		logger.
			With(
				slog.Group("user",
					slog.String("id", "user-123"),
					slog.Time("created_at", time.Now().AddDate(0, 0, -1)),
				),
			).
			With("a", i).
			With("environment", "dev").
			With("error", fmt.Errorf("an error")).
			Error("A message")
	}

	buffer.Flush(true)
	bucket.Close()
}
