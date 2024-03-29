package exploder

import (
	"bytes"
	"fmt"
	"io"

	"github.com/kjk/lzma"
	"github.com/pschou/go-tease"
)

type lzmaFile struct {
	z_reader io.Reader
	eof      bool
	count    int
}

func init() {
	formatTests = append(formatTests, formatTest{
		Test: testLzma,
		Read: readLzma,
		Type: "lzma",
	})
}

func testLzma(tr *tease.Reader, _ string) bool {
	tr.Seek(0, io.SeekStart)
	buf := make([]byte, 5)
	tr.Read(buf)
	tr.Seek(0, io.SeekStart)
	switch {
	case bytes.Compare(buf, []byte{0x5d, 0x00, 0x00, 0x01, 0x00}) == 0 ||
		bytes.Compare(buf, []byte{0x5d, 0x00, 0x00, 0x10, 0x00}) == 0 ||
		bytes.Compare(buf, []byte{0x5d, 0x00, 0x00, 0x08, 0x00}) == 0 ||
		//bytes.Compare(buf, []byte{0x5d, 0x00, 0x00, 0x10, 0x00}) == 0 ||
		bytes.Compare(buf, []byte{0x5d, 0x00, 0x00, 0x20, 0x00}) == 0 ||
		bytes.Compare(buf, []byte{0x5d, 0x00, 0x00, 0x40, 0x00}) == 0 ||
		bytes.Compare(buf, []byte{0x5d, 0x00, 0x00, 0x80, 0x00}) == 0 ||
		bytes.Compare(buf, []byte{0x5d, 0x00, 0x00, 0x00, 0x01}) == 0 ||
		bytes.Compare(buf, []byte{0x5d, 0x00, 0x00, 0x00, 0x02}) == 0:
		return true
	}
	return false
}

func readLzma(tr *tease.Reader, size int64) (archive, error) {
	tr.Seek(0, io.SeekStart)
	r := lzma.NewReader(tr)

	// do a test read to try to trigger a read error
	buf := []byte{0}
	n, err := r.Read(buf)

	// special case if we compressed an empty file
	if !(n == 0 && err == io.EOF) && err != nil {
		if Debug {
			fmt.Println("Error reading lzma", err)
		}
		return nil, err
	}

	tr.Seek(0, io.SeekStart)
	r = lzma.NewReader(tr)
	ret := lzmaFile{
		z_reader: r,
		eof:      false,
	}

	tr.Pipe()
	return &ret, nil
}

func (i *lzmaFile) Type() string { return "lzma" }
func (i *lzmaFile) IsEOF() bool  { return i.eof }
func (c *lzmaFile) Close()       {}

func (i *lzmaFile) Next() (path, name string, r io.Reader, size int64, err error) {
	if i.count == 0 {
		i.count = 1
		i.eof = true
		return ".", "pt_1", i.z_reader, -1, nil
	}
	err = io.EOF
	return
}
