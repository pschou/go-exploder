package exploder

import (
	"bytes"
	"compress/bzip2"
	"fmt"
	"io"

	"github.com/pschou/go-tease"
)

type bZip2File struct {
	z_reader io.Reader
	eof      bool
	count    int
}

func init() {
	formatTests = append(formatTests, formatTest{
		Test:     testBzip2,
		Read:     readBzip2,
		Type:     "bzip2",
		NeedSize: false,
	})
}

func testBzip2(tr *tease.Reader, _ string) bool {
	tr.Seek(0, io.SeekStart)
	buf := make([]byte, 3)
	tr.Read(buf)
	tr.Seek(0, io.SeekStart)
	return bytes.Compare(buf, []byte{0x42, 0x5A, 0x68}) == 0
}

func readBzip2(tr *tease.Reader, size int64) (archive, error) {
	tr.Seek(0, io.SeekStart)
	r := bzip2.NewReader(tr)

	// do a test read to try to trigger a read error
	buf := []byte{0}
	n, err := r.Read(buf)

	// special case if we compressed an empty file
	if !(n == 0 && err == io.EOF) && err != nil {
		if Debug {
			fmt.Println("Error reading bzip2", err)
		}
		return nil, err
	}

	tr.Seek(0, io.SeekStart)
	ret := bZip2File{
		z_reader: bzip2.NewReader(tr),
		eof:      false,
	}

	tr.Pipe()
	return &ret, nil
}

func (i *bZip2File) Type() string { return "bzip2" }
func (i *bZip2File) IsEOF() bool  { return i.eof }

func (c *bZip2File) Close() {
}

func (i *bZip2File) Next() (path, name string, r io.Reader, size int64, err error) {
	size = -1
	if i.count == 0 {
		i.count = 1
		i.eof = true
		return ".", "pt_1", i.z_reader, -1, nil
	}
	err = io.EOF
	return
}
