package exploder

import (
	"bytes"
	"fmt"
	"io"
	"path"

	"github.com/nwaples/rardecode"
	"github.com/pschou/go-tease"
)

type rarFile struct {
	z_reader *rardecode.Reader
	hdr      *rardecode.FileHeader
	eof      bool
}

func init() {
	formatTests = append(formatTests, formatTest{
		Test: testRAR,
		Read: readRAR,
		Type: "rar",
	})
}

func testRAR(tr *tease.Reader, _ string) bool {
	tr.Seek(0, io.SeekStart)
	buf := make([]byte, 6)
	tr.Read(buf)
	tr.Seek(0, io.SeekStart)
	return bytes.Compare(buf, []byte{0x52, 0x61, 0x72, 0x21, 0x1A, 0x07}) == 0
}

func readRAR(tr *tease.Reader, size int64) (archive, error) {
	tr.Seek(0, io.SeekStart)
	if size < 10 {
		size = 2048
	}
	zr, err := rardecode.NewReader(tr, "")
	if err != nil {
		if Debug {
			fmt.Println("Error reading rar", err)
		}
		return nil, err
	}

	hdr, err := zr.Next()
	if err != nil {
		return nil, err
	}

	ret := rarFile{
		z_reader: zr,
		eof:      false,
		hdr:      hdr,
	}

	tr.Pipe()
	return &ret, nil
}

func (i *rarFile) Type() string { return "rar" }
func (i *rarFile) IsEOF() bool  { return i.eof }
func (c *rarFile) Close() {
	//if c.z_reader != nil {
	//	c.z_reader.Close()
	//}
}

func (i *rarFile) Next() (dir, name string, r io.Reader, size int64, err error) {
	var hdr *rardecode.FileHeader
	for {
		if i.hdr != nil {
			hdr = i.hdr
			i.hdr = nil
		} else {
			hdr, err = i.z_reader.Next()
			if err != nil {
				return
			}
		}
		if !hdr.IsDir {
			break
		}
	}

	r = i.z_reader
	size = hdr.UnPackedSize
	dir, name = path.Split(hdr.Name)
	return
}
