package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"testing"
)

type TestLogger struct {
	logs       bytes.Buffer
	std        io.Writer
	logsChan   chan []byte
	signalChan chan struct{}
	wg         sync.WaitGroup
}

func (tl *TestLogger) Write(p []byte) (n int, err error) {
	// Always write to buffer in tests, only print if needed
	b := make([]byte, len(p))
	copy(b, p)
	tl.logsChan <- b
	return len(p), nil // pretend we wrote everything
}

func (tl *TestLogger) process() {
	defer tl.wg.Done()
	for {
		select {
		case <-tl.signalChan:
			return
		case data := <-tl.logsChan:
			tl.logs.Write(data)
		}
	}
}

func NewTestLogger(t testing.TB) AppLogger {
	tl := &TestLogger{
		std:        os.Stdout,
		logsChan:   make(chan []byte, 1_000),
		signalChan: make(chan struct{}),
	}
	tl.wg.Add(1)
	go tl.process()
	t.Cleanup(func() {
		tl.signalChan <- struct{}{}
		tl.wg.Wait()
		if t.Failed() {
			fmt.Fprint(os.Stderr, tl.logs.String()) // only print logs if test failed
		}
	})
	return InitLogger([]io.Writer{tl}, "debug", "tl")
}
