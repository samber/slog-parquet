package slogparquet

import "time"

type log struct {
	Time       time.Time `parquet:"time"`
	LogLevel   string    `parquet:"log_level"`
	Message    string    `parquet:"message"`
	Attributes []byte    `parquet:"attributes"`
	Source     string    `parquet:"source"`
}
