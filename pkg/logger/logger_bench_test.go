package logger_test

import (
	"io"
	"testing"

	"github.com/getoptimum/optimum-common/pkg/logger"
)

type writer struct{}

func (w writer) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func BenchmarkLogger(b *testing.B) {
	l0 := logger.InitLogger([]io.Writer{writer{}}, logger.Debug)
	l0 = l0.With(logger.WithService("service1"))
	for b.Loop() {
		l0.Info("hello world")
	}
}
