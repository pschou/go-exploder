package exploder

import (
	"bytes"
	"fmt"
	"io"
	"path"

	"github.com/cavaliergopher/cpio"
	"github.com/pschou/go-tease"
)

type cpioFile struct {
	cpio_reader *cpio.Reader
	hdr         *cpio.Header
	eof         bool
	count       int
}

func init() {
	formatTests = append(formatTests, formatTest{
		Test: testCPIO,
		Read: readCPIO,
		Type: "cpio",
	})
}

func testCPIO(tr *tease.Reader, _ string) bool {
	tr.Seek(0, io.SeekStart)
	buf := make([]byte, 6)
	tr.Read(buf)
	return bytes.Compare(buf, []byte{0x30, 0x37, 0x30, 0x37, 0x30, 0x37}) == 0
}

func readCPIO(tr *tease.Reader, size int64) (archive, error) {
	r := cpio.NewReader(tr)

	hdr, err := r.Next()
	if err != nil {
		if Debug {
			fmt.Println("Error reading cpio", err)
		}
		return nil, err
	}

	ret := cpioFile{
		cpio_reader: r,
		hdr:         hdr,
		eof:         false,
	}

	tr.Pipe()
	return &ret, nil
}

func (i *cpioFile) Type() string { return "cpio" }
func (i *cpioFile) IsEOF() bool  { return i.eof }
func (i *cpioFile) Close()       {}

func (i *cpioFile) Next() (dir, name string, r io.Reader, size int64, err error) {
	var hdr *cpio.Header
	for {
		if i.hdr != nil {
			hdr = i.hdr
			i.hdr = nil
		} else {
			hdr, err = i.cpio_reader.Next()
			if err != nil {
				return "", "", nil, 0, io.EOF
			}
		}
		if hdr.Mode&^cpio.ModePerm == cpio.TypeReg {
			break
		}
	}
	r = i.cpio_reader
	size = hdr.Size
	dir, name = path.Split(hdr.Name)
	return
}
