package exploder

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/pschou/go-cabfile/cabfile"
	"github.com/pschou/go-tease"
)

type cABFile struct {
	z_reader    *io.Reader
	cab         *cabfile.Cabinet
	eof         bool
	first_finfo os.FileInfo
	first_r     io.Reader
}

func init() {
	formatTests = append(formatTests, formatTest{
		Test:     testCAB,
		Read:     readCAB,
		Type:     "cab",
		NeedSize: false,
	})
}

func testCAB(tr *tease.Reader, _ string) bool {
	tr.Seek(0, io.SeekStart)
	buf := make([]byte, 4)
	tr.Read(buf)
	tr.Seek(0, io.SeekStart)
	return bytes.Compare(buf, []byte{0x4D, 0x53, 0x43, 0x46}) == 0
}

func readCAB(tr *tease.Reader, size int64) (archive, error) {
	tr.Seek(0, io.SeekStart)
	cab, err := cabfile.New(tr)
	if err != nil {
		if Debug {
			fmt.Println("Error reading cab", err)
		}
		return nil, err
	}

	ret := cABFile{
		eof: false,
		cab: cab,
	}

	ret.first_r, ret.first_finfo, err = cab.Next()
	if err != nil {
		return nil, err
	}
	tr.Pipe()
	//fmt.Println("piped reader", cab)
	return &ret, nil
}

func (i *cABFile) Type() string { return "cab" }
func (i *cABFile) IsEOF() bool  { return i.eof }
func (c *cABFile) Close() {
	//if c.z_reader != nil {
	//	c.z_reader.Close()
	//}
}

func (i *cABFile) Next() (dir, name string, r io.Reader, size int64, err error) {
	if i.first_r != nil {
		r = i.first_r
		i.first_r = nil
		dir, name = path.Split(i.first_finfo.Name())
		return
	}

	var finfo os.FileInfo
	r, finfo, err = i.cab.Next()
	//fmt.Println("next called", r, finfo, err)
	if err == io.EOF {
		i.eof = true
	}
	if err != nil {
		return
	}
	size = finfo.Size()
	dir, name = path.Split(finfo.Name())
	return
}
