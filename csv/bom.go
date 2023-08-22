package csv

import (
	"io"
	"os"
)

const (
	bom0 = 0xef
	bom1 = 0xbb
	bom2 = 0xbf
)

// SkipBom will remove bom from the file if it has bom
//
//	 will skip any error occur during the process, and reset to start position
//	return true means bom is skipped , or no bom is skipped
func SkipBom(file *os.File) bool {
	// get current offset of this file
	offset, err := file.Seek(0, io.SeekCurrent)
	if err != nil {
		// skip all error
		return false
	}

	// read first 3 bytes from current offset
	b := make([]byte, 3)
	n, err := file.Read(b)

	if err != nil || n != 3 {
		// skip all error
		// but try seek to offset
		_, _ = file.Seek(offset, io.SeekStart)
		return false
	}

	// check if it is the bom
	if b[0] == bom0 && b[1] == bom1 && b[2] == bom2 {
		// we're done , bom is skipped !
		return true
	}

	// or we rollback
	_, _ = file.Seek(offset, io.SeekStart)

	return false
}

type SkipBomReader struct {
	r   io.Reader
	buf []byte
	err error
}

func (r *SkipBomReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	if len(r.buf) == 0 {
		if r.err != nil {
			return 0, r.readErr()
		}
		return r.r.Read(p)
	}

	n := copy(p, r.buf)
	r.buf = r.buf[n:]
	return n, nil
}

func (r *SkipBomReader) readErr() error {
	err := r.err
	r.err = nil
	return err
}

// NewSkipBomReader returns a reader which can remove bom from the given io.Reader if it has bom
func NewSkipBomReader(r io.Reader) *SkipBomReader {
	if r == nil {
		return nil
	}

	b := [3]byte{}

	n, err := r.Read(b[:])
	if err != nil || n < 3 {
		return &SkipBomReader{r, b[:n], err}
	}

	// check if it is the bom
	if b[0] == bom0 && b[1] == bom1 && b[2] == bom2 {
		// we're done , bom is skipped !
		return &SkipBomReader{r, nil, nil}
	}

	return &SkipBomReader{r, b[:], nil}
}
