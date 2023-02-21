# go-exploder

This is a generalized archive exploder which can take an array of archive formats and expand them for inspection.

Archive formats supported:
  - 7zip
  - bzip2
  - cab
  - debian
  - gzip / bgzf / apk
  - iso9660
  - lzma
  - rar
  - rpm
  - tar
  - xz
  - zip
  - zstd

# Example

```golang
fh, err := os.Open("data.zip")
stat := fh.Stat()
err = Explode(filePath string, in io.Reader, size int64, -1)
```

# Documentation

Documentation and usage can be found at:

https://pkg.go.dev/github.com/pschou/go-exploder
