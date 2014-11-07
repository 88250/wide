//部分代码参考(Ctrl+c)自golang bufio
package util

import (
	"errors"
	"io"
	"unicode/utf8"
)

const (
	defaultBufSize = 4096
)

type Reader struct {
	buf  []byte
	rd   io.Reader
	r, w int
	err  error
}

const minReadBufferSize = 16
const maxConsecutiveEmptyReads = 100

func NewReaderSize(rd io.Reader, size int) *Reader {
	b, ok := rd.(*Reader)
	if ok && len(b.buf) >= size {
		return b
	}
	if size < minReadBufferSize {
		size = minReadBufferSize
	}
	r := new(Reader)
	r.reset(make([]byte, size), rd)
	return r
}

// NewReader returns a new Reader whose buffer has the default size.
func NewReader(rd io.Reader) *Reader {
	return NewReaderSize(rd, defaultBufSize)
}

func (b *Reader) reset(buf []byte, r io.Reader) {
	*b = Reader{
		buf: buf,
		rd:  r,
	}
}

func (b *Reader) Read(buf []byte) (count int, err error) {
	return
}

var errNegativeRead = errors.New("bufio: reader returned negative count from Read")

// fill reads a new chunk into the buffer.
func (b *Reader) fill() {
	// Slide existing data to beginning.
	if b.r > 0 {
		copy(b.buf, b.buf[b.r:b.w])
		b.w -= b.r
		b.r = 0
	}

	if b.w >= len(b.buf) {
		panic("bufio: tried to fill full buffer")
	}

	// Read new data: try a limited number of times.
	for i := maxConsecutiveEmptyReads; i > 0; i-- {
		n, err := b.rd.Read(b.buf[b.w:])
		if n < 0 {
			panic(errNegativeRead)
		}
		b.w += n
		if err != nil {
			b.err = err
			return
		}
		if n > 0 {
			return
		}
	}
	b.err = io.ErrNoProgress
}

func (b *Reader) readErr() error {
	err := b.err
	b.err = nil
	return err
}

func (b *Reader) Buffered() int { return b.w - b.r }

func (b *Reader) ReadData() (line string, err error) {
	if n := b.Buffered(); n < len(b.buf) {
		b.fill()
	}
	if b.err != nil {
		line = string(b.buf[b.r:b.w])
		b.r = b.w
		err = b.readErr()
		return
	}
	index := b.w
	for {
		r, _ := utf8.DecodeLastRune(b.buf[b.r:index])
		if r == utf8.RuneError {
			index = index - 1
			if index < b.r {
				break
			}
		} else {
			break
		}
	}
	if index > b.r {
		line = string(b.buf[b.r : b.r+index])
		b.r += index
	} else {
		panic("no utf8 char found")
	}
	return
}
