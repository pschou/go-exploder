package exploder

import (
	"bytes"
	"fmt"
	"io"
	"path"

	"github.com/bodgit/sevenzip"
	"github.com/pschou/go-tease"
)

type sevenZipFile struct {
	z_reader *sevenzip.Reader
	eof      bool
	count    int
}

func init() {
	formatTests = append(formatTests, formatTest{
		Test:     test7Zip,
		Read:     read7Zip,
		Type:     "7zip",
		NeedSize: true,
	})
}

func test7Zip(tr *tease.Reader, _ string) bool {
	tr.Seek(0, io.SeekStart)
	buf := make([]byte, 6)
	tr.Read(buf)
	tr.Seek(0, io.SeekStart)
	return bytes.Compare(buf, []byte{0x37, 0x7A, 0xBC, 0xAF, 0x27, 0x1C}) == 0
}

func (i *sevenZipFile) Type() string { return "7zip" }
func (i *sevenZipFile) IsEOF() bool  { return i.eof }

func read7Zip(tr *tease.Reader, size int64) (archive, error) {
	tr.Seek(0, io.SeekStart)
	if size < 10 {
		size = 2048
	}
	zr, err := sevenzip.NewReader(tr, size)
	if err != nil {
		if Debug {
			fmt.Println("Error reading 7zip", err)
		}
		return nil, err
	}

	ret := sevenZipFile{
		z_reader: zr,
		eof:      false,
	}

	tr.Seek(0, io.SeekStart)
	tr.Pipe()
	return &ret, nil
}

func (c *sevenZipFile) Close() {
	//if c.z_reader != nil {
	//	c.z_reader.Close()
	//}
}

func (i *sevenZipFile) Next() (dir, name string, r io.Reader, size int64, err error) {
	var f *sevenzip.File
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
	size = int64(f.UncompressedSize)
	dir, name = path.Split(f.Name)
	//fmt.Println("path", dir, name, "f.Name=", f.Name)
	return
}
