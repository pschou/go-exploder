package exploder

import (
	"bytes"
	"fmt"
	"io"
	"path"

	"github.com/cavaliergopher/cpio"
	"github.com/cavaliergopher/rpm"
	"github.com/klauspost/compress/zstd"
	"github.com/pschou/go-tease"
	"github.com/ulikunitz/xz"
)

type rpmFile struct {
	reader      io.Reader
	cpio_reader *cpio.Reader
	pkg         *rpm.Package
	eof         bool
	size        int64
	count       int
}

func init() {
	formatTests = append(formatTests, formatTest{
		Test: testRPM,
		Read: readRPM,
		Type: "rpm",
	})
}

func testRPM(tr *tease.Reader, _ string) bool {
	tr.Seek(0, io.SeekStart)
	buf := make([]byte, 4)
	tr.Read(buf)
	return bytes.Compare(buf, []byte{0xED, 0xAB, 0xEE, 0xDB}) == 0
}

func readRPM(tr *tease.Reader, size int64) (archive, error) {

	// Read the package headers
	pkg, err := rpm.Read(tr)
	if err != nil {
		if Debug {
			fmt.Println("Error reading rpm", err)
		}
		return nil, err
	}

	var reader io.Reader

	// Check the compression algorithm of the payload
	switch compression := pkg.PayloadCompression(); compression {
	case "xz":
		// Attach a reader to decompress the payload
		reader, err = xz.NewReader(tr)
		if err != nil {
			return nil, err
		}
	case "zstd":
		// Attach a reader to decompress the payload
		reader, err = zstd.NewReader(tr)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("Unsupported compression: %s", compression)
	}

	// Check the archive format of the payload
	if format := pkg.PayloadFormat(); format != "cpio" {
		return nil, fmt.Errorf("Unsupported payload format: %s", format)
	}

	// Attach a reader to unarchive each file in the payload
	cpioReader := cpio.NewReader(reader)

	ret := rpmFile{
		reader:      reader,
		cpio_reader: cpioReader,
		eof:         false,
		size:        size,
		pkg:         pkg,
	}

	tr.Pipe()
	return &ret, nil
}

func (i *rpmFile) Type() string { return "rpm" }
func (i *rpmFile) IsEOF() bool  { return i.eof }
func (c *rpmFile) Close() {
	//if c.z_reader != nil {
	//	c.z_reader.Close()
	//}
}

func (i *rpmFile) Next() (dir, name string, r io.Reader, size int64, err error) {
	var hdr *cpio.Header
	for {
		// Move to the next file in the archive
		hdr, err = i.cpio_reader.Next()
		if err != nil {
			return
		}

		// Skip directories and other irregular file types in this example
		if hdr.Mode.IsRegular() {
			break
		}
	}
	r = i.cpio_reader
	dir, name = path.Split(hdr.Name)
	size = hdr.Size
	//fmt.Println("returning", dir, name, r, err)
	return
}
