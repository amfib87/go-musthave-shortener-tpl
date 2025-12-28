package gzip

import (
	"compress/gzip"
	"io"
	"net/http"
)

type CompressWriter struct {
	W  http.ResponseWriter
	Zw *gzip.Writer
}

func NewCompressWriter(w http.ResponseWriter) *CompressWriter {
	return &CompressWriter{
		W:  w,
		Zw: gzip.NewWriter(w),
	}
}

func (c *CompressWriter) Header() http.Header {
	return c.W.Header()
}

func (c *CompressWriter) Write(p []byte) (int, error) {
	return c.Zw.Write(p)
}

func (c *CompressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.W.Header().Set("Content-Encoding", "gzip")
	}
	c.W.WriteHeader(statusCode)
}

func (c *CompressWriter) Close() error {
	return c.Zw.Close()
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func NewCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}
