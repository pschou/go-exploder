package exploder

import (
	"archive/tar"
	"fmt"
	"io"
	"path"

	"github.com/pschou/go-tease"
)

type TarFile struct {
	a_reader *tar.Reader
	eof      bool
	hdr      *tar.Header
	size     int64
	count    int
}

func init() {
	formatTests = append(formatTests, formatTest{
		Test:     testTar,
		Read:     readTar,
		Type:     "tar",
		NeedSize: false,
	})
}

func testTar(tr *tease.Reader, _ string) bool {
	tr.Seek(0, io.SeekStart)
	ar := tar.NewReader(tr)
	_, err := ar.Next()
	tr.Seek(0, io.SeekStart)
	return err == nil
}

func readTar(tr *tease.Reader, size int64) (Archive, error) {
	tr.Seek(0, io.SeekStart)
	ar := tar.NewReader(tr)
	hdr, err := ar.Next()
	if err != nil {
		if Debug {
			fmt.Println("Error reading tar", err)
		}
		return nil, err
	}

	ret := TarFile{
		a_reader: ar,
		eof:      false,
		size:     size,
		hdr:      hdr,
	}

	tr.Pipe()
	return &ret, nil
}

func (i *TarFile) Type() string { return "tar" }
func (i *TarFile) IsEOF() bool  { return i.eof }

func (c *TarFile) Close() {
	//if c.z_reader != nil {
	//	c.z_reader.Close()
	//}
}

func (i *TarFile) Next() (dir, name string, r io.Reader, size int64, err error) {
	var hdr *tar.Header
	for {
		if i.hdr != nil {
			hdr = i.hdr
			i.hdr = nil
		} else {
			hdr, err = i.a_reader.Next()
			if err != nil {
				return "", "", nil, 0, io.EOF
			}
		}
		if hdr.Typeflag == tar.TypeReg {
			break
		}
	}
	r = i.a_reader
	size = hdr.Size
	dir, name = path.Split(hdr.Name)
	//fmt.Println("returning", dir, name, r, err)
	return
}
