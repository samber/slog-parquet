package slogparquet

type Logger struct {
}

// NewLogger returns a logger that doesn't do anything.
func NewLogger() Logger {
	return Logger{}
}

func (Logger) Log(...interface{}) error {
	return nil
}
