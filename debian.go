package exploder

import (
	"bytes"
	"fmt"
	"io"

	"github.com/pschou/go-tease"
	deb "pault.ag/go/debian/deb"
)

type dEBFile struct {
	ar     *deb.Ar
	ar_ent *deb.ArEntry
	eof    bool
	size   int64
}

func init() {
	formatTests = append(formatTests, formatTest{
		Test: testDEB,
		Read: readDEB,
		Type: "debian",
	})
}

func testDEB(tr *tease.Reader, fn string) bool {
	tr.Seek(0, io.SeekStart)
	buf := make([]byte, 8)
	tr.Read(buf)
	tr.Seek(0, io.SeekStart)
	return bytes.Compare(buf, []byte{
		0x21, 0x3C, 0x61, 0x72, 0x63, 0x68, 0x3E, 0x0A}) == 0
}

func readDEB(tr *tease.Reader, size int64) (archive, error) {

	ar, err := deb.LoadAr(tr)
	if err != nil {
		if Debug {
			fmt.Println("Error reading debian", err)
		}
		return nil, err
	}
	ar_ent, err := ar.Next()
	if err != nil {
		return nil, err
	}
	tr.Pipe()
	ret := dEBFile{
		ar:     ar,
		ar_ent: ar_ent,
		eof:    false,
		size:   size,
	}

	return &ret, nil
}

func (i *dEBFile) Type() string { return "debian" }
func (i *dEBFile) IsEOF() bool  { return i.eof }

func (c *dEBFile) Close() {
	//if c.z_reader != nil {
	//	c.z_reader.Close()
	//}
}

func (i *dEBFile) Next() (dir, name string, r io.Reader, size int64, err error) {
	var ar_ent *deb.ArEntry
	for {
		if i.ar_ent != nil {
			ar_ent = i.ar_ent
			i.ar_ent = nil
		} else {
			ar_ent, err = i.ar.Next()
			if err == io.EOF {
				return
			}
		}
		if ar_ent == nil {
			err = fmt.Errorf("Invalid DEB file")
			return
		}
		if ar_ent.IsTarfile() {
			break
		}
	}
	var c interface{}
	_, c, err = ar_ent.Tarfile()
	if err != nil {
		return
	}
	if ir, ok := (c).(io.Reader); ok {
		r = ir
	} else {
		err = io.EOF
		return
	}
	dir = "."
	name = ar_ent.Name
	size = ar_ent.Size
	return
}
