package tail

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
)

const (
	// blockSize is the block size used in tail.
	blockSize = 1024
)

var (
	// eol is the end-of-line sign in the log.
	eol = []byte{'\n'}
)

var (
	ErrInvalidN  = errors.New("cannot tail history with n <= 0")
	ErrEmptyFile = errors.New("cannot tail an empty file")
)

// Tail returns a Reader that starts reading from the nth line from the end of the file.
// * If n < 0, Tail returns an error
// * If n >= 0, Tail returns a Reader that starts reading from the beginning of the last nth line.
// NOTE: It is the caller's responsibility to close the Reader when finished.
func Tail(filename string, n int64) (io.ReadCloser, error) {
	if n <= 0 {
		return nil, ErrInvalidN
	}
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	start, err := findTailLineStartIndex(f, n)
	if err != nil {
		return nil, err
	}
	if _, err := f.Seek(start, io.SeekStart); err != nil {
		return nil, err
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if fi.Size() == 0 {
		return nil, fmt.Errorf("%w: %s", ErrEmptyFile, filename)
	}
	return f, nil
}

// findTailLineStartIndex returns the start of last nth line.
// * If n < 0, return the beginning of the file.
// * If n >= 0, return the beginning of last nth line.
// Notice that if the last line is incomplete (no end-of-line), it will not be counted
// as one line.
// ATTRIBUTION: https://github.com/kubernetes/kubernetes/blob/4b8e819355d791d96b7e9d9efe4cbafae2311c88/pkg/util/tail/tail.go#L63
func findTailLineStartIndex(f io.ReadSeeker, n int64) (int64, error) {
	if n < 0 {
		return 0, nil
	}
	size, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err
	}
	var left, cnt int64
	buf := make([]byte, blockSize)
	for right := size; right > 0 && cnt <= n; right -= blockSize {
		left = right - blockSize
		if left < 0 {
			left = 0
			buf = make([]byte, right)
		}
		if _, err := f.Seek(left, io.SeekStart); err != nil {
			return 0, err
		}
		if _, err := f.Read(buf); err != nil {
			return 0, err
		}
		cnt += int64(bytes.Count(buf, eol))
	}
	for ; cnt > n; cnt-- {
		idx := bytes.Index(buf, eol) + 1
		buf = buf[idx:]
		left += int64(idx)
	}
	return left, nil
}
