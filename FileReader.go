package exploder

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

// FileReader provides the ability to read the underlying bytes from a file in an archive.
type FileReader struct {
	r       io.Reader
	tmpfile string
	n, size int64
	err     error
}

// Read implement the io.Reader interface
func (r *FileReader) Read(p []byte) (n int, err error) {
	if r.err != nil {
		return 0, r.err
	}
	if r.size == -1 || int64(len(p)) <= r.size-r.n {
		n, r.err = r.r.Read(p)
		r.n += int64(n)
	} else if r.n < r.size {
		n, r.err = r.r.Read(p[:r.size-r.n])
		r.n += int64(n)
	}
	if r.n == r.size && r.err == nil {
		err = io.EOF
	}
	if err != nil && r.tmpfile != "" {
		if ct, ok := r.r.(*os.File); ok {
			ct.Close()
		}
		os.Remove(r.tmpfile) // clean up
		r.tmpfile = ""
	}
	return
}

// Return the size if known, other wise determine the size by reading off the buffer and then returning the size.
func (r *FileReader) Size() int64 {
	if r.size < 0 {
		fmt.Println("create temp")
		f, err := os.CreateTemp("", "exploder")
		if err != nil {
			return -1
		}
		r.tmpfile = f.Name()
		r.size, _ = io.Copy(f, r.r)
		r.r = bufio.NewReader(f)
	}
	return r.size
}

func (r *FileReader) finalize(name string) error {
	if r.err != nil {
		return fmt.Errorf("FileReader: %s, %s", name, r.err)
	}
	if _, err := io.Copy(io.Discard, r); err != nil {
		return fmt.Errorf("FileReader: %s, %s", name, err)
	} else if r.size >= 0 && r.n != r.size {
		return fmt.Errorf("FileReader: copied file size %d, and expected %d, do not match in %s", r.n, r.size, name)
	}
	return nil
}
