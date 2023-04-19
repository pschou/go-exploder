package exploder

import (
	"bytes"
	"fmt"
	"io"

	"github.com/pschou/go-tease"
	"github.com/ulikunitz/xz"
)

type xzFile struct {
	z_reader *xz.Reader
	eof      bool
	count    int
}

func init() {
	formatTests = append(formatTests, formatTest{
		Test: testXZ,
		Read: readXZ,
		Type: "xz",
	})
}

func testXZ(tr *tease.Reader, _ string) bool {
	tr.Seek(0, io.SeekStart)
	buf := make([]byte, 6)
	tr.Read(buf)
	return bytes.Compare(buf, []byte{0xFD, 0x37, 0x7A, 0x58, 0x5A, 0x00}) == 0
}

func readXZ(tr *tease.Reader, size int64) (archive, error) {
	tr.Seek(0, io.SeekStart)
	r, err := xz.NewReader(tr)
	if err != nil {
		if Debug {
			fmt.Println("Error reading xz", err)
		}
		return nil, err
	}

	ret := xzFile{
		z_reader: r,
		eof:      false,
	}

	tr.Pipe()
	return &ret, nil
}

func (i *xzFile) Type() string { return "xz" }
func (i *xzFile) IsEOF() bool  { return i.eof }
func (c *xzFile) Close()       {}

func (i *xzFile) Next() (path, name string, r io.Reader, size int64, err error) {
	if i.count == 0 {
		i.count = 1
		i.eof = true
		return ".", "pt_1", i.z_reader, -1, nil
	}
	err = io.EOF
	return
}
