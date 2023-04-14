package exploder

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/biogo/hts/bgzf"
	"github.com/pschou/go-tease"
)

type gZipFile struct {
	buf_reader *bufio.Reader
	gz_reader  *gzip.Reader
	bgz_reader *bgzf.Reader
	tr_reader  *tease.Reader
	eof        bool
	count      int
	gz_type    string
}

func init() {
	formatTests = append(formatTests, formatTest{
		Test:     testGzip,
		Read:     readGzip,
		Type:     "gzip / bgzf / apk",
		NeedSize: false,
	})
}

func testGzip(tr *tease.Reader, _ string) bool {
	tr.Seek(0, io.SeekStart)
	buf := make([]byte, 2)
	tr.Read(buf)
	tr.Seek(0, io.SeekStart)
	return bytes.Compare(buf, []byte{0x1f, 0x8b}) == 0
}

func readGzip(tr *tease.Reader, size int64) (archive, error) {
	a, err := readBlockGzip(tr, size)
	if err != nil {
		//fmt.Println("err:", err)
		a, err = readStandardGzip(tr, size)
	}
	if err == nil {
		tr.Pipe()
	}
	return a, err
}

func readStandardGzip(tr *tease.Reader, size int64) (archive, error) {
	tr.Seek(0, io.SeekStart)
	br := bufio.NewReader(tr)
	gzr, err := gzip.NewReader(br)
	if err != nil {
		if Debug {
			fmt.Println("Error reading gzip", err)
		}
		return nil, err
	}

	// Read off one byte for a test
	b := make([]byte, 1)
	n, err := gzr.Read(b)
	if err != nil && err != io.EOF {
		return nil, err
	}
	if n == 0 {
		return nil, errors.New("Gzip test failed")
	}

	// Return to the beginning and reset
	tr.Seek(0, io.SeekStart)
	err = gzr.Reset(tr)
	if err != nil {
		return nil, err
	}
	gzr.Multistream(false)

	ret := gZipFile{
		buf_reader: br,
		gz_reader:  gzr,
		tr_reader:  tr,
		eof:        false,
		gz_type:    "gzip",
	}

	return &ret, nil
}

func readBlockGzip(tr *tease.Reader, size int64) (archive, error) {
	tr.Seek(0, io.SeekStart)
	br := bufio.NewReader(tr)
	gzr, err := bgzf.NewReader(br, 1)
	if err != nil {
		if Debug {
			fmt.Println("Error reading gzip", err)
		}
		return nil, err
	}

	// Read off one byte for a test
	b := make([]byte, 2048)
	n, err := gzr.Read(b)
	if err != nil && err != io.EOF {
		return nil, err
	}
	if n == 0 {
		return nil, errors.New("Gzip test failed")
	}
	gzr.Close()

	// Return to the beginning and reset
	tr.Seek(0, io.SeekStart)
	gzr, err = bgzf.NewReader(br, 1)
	if err != nil {
		return nil, err
	}

	ret := gZipFile{
		buf_reader: br,
		bgz_reader: gzr,
		tr_reader:  tr,
		eof:        false,
		gz_type:    "bgzf",
	}

	return &ret, nil
}

func (i *gZipFile) Type() string { return i.gz_type }
func (i *gZipFile) IsEOF() bool  { return i.eof }
func (c *gZipFile) Close() {
	if c.buf_reader != nil {
		c.buf_reader.Reset(nil)
	}
	if c.gz_reader != nil {
		c.gz_reader.Close()
	}
	if c.bgz_reader != nil {
		c.bgz_reader.Close()
	}
}

func (i *gZipFile) Next() (path, name string, r io.Reader, size int64, err error) {
	if Debug {
		fmt.Println("next() called")
	}
	if i.count == 0 {
		i.count = 1
		if i.gz_reader != nil {
			return ".", "pt_1", i.gz_reader, -1, nil
		}
		return ".", "pt_1", i.bgz_reader, -1, nil
	}
	if i.gz_reader == nil {
		return "", "", nil, -1, io.EOF
	}

	if Debug {
		fmt.Println("dumping out rest of file")
	}

	io.Copy(ioutil.Discard, i.gz_reader)

	if Debug {
		fmt.Println("gzip reset")
	}
	err = i.gz_reader.Reset(i.buf_reader)
	if err != nil {
		i.eof = true
		return "", "", nil, -1, err
	}
	i.gz_reader.Multistream(false)
	i.count++
	return ".", fmt.Sprintf("pt_%d", i.count), i.gz_reader, -1, nil
}
