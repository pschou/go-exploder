package exploder

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"path"

	"github.com/pschou/go-tease"
)

type ZipFile struct {
	z_reader *zip.Reader
	eof      bool
	count    int
}

func init() {
	formatTests = append(formatTests, formatTest{
		Test:     testZip,
		Read:     readZip,
		Type:     "zip",
		NeedSize: true,
	})
}

func testZip(tr *tease.Reader, fn string) bool {
	tr.Seek(0, io.SeekStart)
	buf := make([]byte, 4)
	tr.Read(buf)
	tr.Seek(0, io.SeekStart)
	return bytes.Compare(buf, []byte{0x50, 0x4b, 0x03, 0x04}) == 0
}

func readZip(tr *tease.Reader, size int64) (Archive, error) {
	tr.Seek(0, io.SeekStart)
	zr, err := zip.NewReader(tr, size)
	if err != nil {
		if Debug {
			fmt.Println("Error reading zip", err)
		}
		return nil, err
	}

	ret := ZipFile{
		z_reader: zr,
		eof:      false,
	}

	tr.Seek(0, io.SeekStart)
	tr.Pipe()
	return &ret, nil
}

func (i *ZipFile) Type() string { return "zip" }
func (i *ZipFile) IsEOF() bool  { return i.eof }

func (c *ZipFile) Close() {
	//if c.z_reader != nil {
	//	c.z_reader.Close()
	//}
}

func (i *ZipFile) Next() (dir, name string, r io.Reader, size int64, err error) {
	var f *zip.File
	for {
		if i.count >= len(i.z_reader.File) {
			err = io.EOF
			return
		}
		f = i.z_reader.File[i.count]
		i.count++
		if !f.FileInfo().IsDir() {
			break
		}
	}

	r, err = f.Open()
	if err != nil {
		return
	}
	dir, name = path.Split(f.Name)
	//fmt.Println("path", dir, name, "f.Name=", f.Name)
	size = f.FileInfo().Size()
	return
}
