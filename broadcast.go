// broadcast package
// some code from dotcloud/docker
package main

import (
	"bytes"
	"errors"
	"io"
	"sync"
	"time"
)

var ErrTimeout = errors.New("timeout")

func GoTimeout(f func() error, timeout time.Duration) (err error) {
	done := make(chan bool)
	go func() {
		err = f()
		done <- true
	}()
	select {
	case <-time.After(timeout):
		return ErrTimeout
	case <-done:
		return
	}
}

type StreamWriter struct {
	wc     io.WriteCloser
	stream string
}

type WriteBroadcaster struct {
	sync.Mutex
	buf     *bytes.Buffer
	writers map[StreamWriter]bool
	closed  bool
}

func NewWriteBroadcaster() *WriteBroadcaster {
	bc := &WriteBroadcaster{
		writers: make(map[StreamWriter]bool),
		buf:     bytes.NewBuffer(nil),
		closed:  false,
	}
	return bc
}

func (w *WriteBroadcaster) AddWriter(writer io.WriteCloser, stream string) {
	w.Lock()
	defer w.Unlock()
	if w.closed {
		writer.Close()
		return
	}
	sw := StreamWriter{wc: writer, stream: stream}
	w.writers[sw] = true
}

func (wb *WriteBroadcaster) Closed() bool {
	return wb.closed
}

func (wb *WriteBroadcaster) NewReader(name string) ([]byte, *io.PipeReader) {
	r, w := io.Pipe()
	wb.AddWriter(w, name)
	return wb.buf.Bytes(), r
}

func (wb *WriteBroadcaster) NewBufReader(name string) ([]byte, io.ReadCloser) {
	data, rd := wb.NewReader(name)
	return data, NewBufReader(rd)
}

func (wb *WriteBroadcaster) Bytes() []byte {
	return wb.buf.Bytes()
}

func (w *WriteBroadcaster) Write(p []byte) (n int, err error) {
	w.Lock()
	defer w.Unlock()
	w.buf.Write(p)
	for sw := range w.writers {
		// set write timeout
		err = GoTimeout(func() error {
			if n, err := sw.wc.Write(p); err != nil || n != len(p) {
				return errors.New("broadcast to " + sw.stream + " error")
			}
			return nil
		}, time.Second*1)
		if err != nil {
			// On error, evict the writer
			// Debugf("broadcase write error: %s, %s", sw.stream, err)
			delete(w.writers, sw)
		}
	}
	return len(p), nil
}

func (w *WriteBroadcaster) CloseWriters() error {
	w.Lock()
	defer w.Unlock()
	for sw := range w.writers {
		sw.wc.Close()
	}
	w.writers = make(map[StreamWriter]bool)
	w.closed = true
	return nil
}

// nop writer
type NopWriter struct{}

func (*NopWriter) Write(buf []byte) (int, error) {
	return len(buf), nil
}

type nopWriteCloser struct {
	io.Writer
}

func (w *nopWriteCloser) Close() error { return nil }

func NopWriteCloser(w io.Writer) io.WriteCloser {
	return &nopWriteCloser{w}
}

type bufReader struct {
	sync.Mutex
	buf    *bytes.Buffer
	reader io.Reader
	err    error
	wait   sync.Cond
}

func NewBufReader(r io.Reader) *bufReader {
	reader := &bufReader{
		buf:    &bytes.Buffer{},
		reader: r,
	}
	reader.wait.L = &reader.Mutex
	go reader.drain()
	return reader
}

func (r *bufReader) drain() {
	buf := make([]byte, 1024)
	for {
		n, err := r.reader.Read(buf)
		r.Lock()
		if err != nil {
			r.err = err
		} else {
			r.buf.Write(buf[0:n])
		}
		r.wait.Signal()
		r.Unlock()
		if err != nil {
			break
		}
	}
}

func (r *bufReader) Read(p []byte) (n int, err error) {
	r.Lock()
	defer r.Unlock()
	for {
		n, err = r.buf.Read(p)
		if n > 0 {
			return n, err
		}
		if r.err != nil {
			return 0, r.err
		}
		r.wait.Wait()
	}
}

func (r *bufReader) Close() error {
	closer, ok := r.reader.(io.ReadCloser)
	if !ok {
		return nil
	}
	return closer.Close()
}
