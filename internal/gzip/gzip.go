package gzip

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

func GzipMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		origRes := res

		contentType := req.Header.Get("Content-Type")
		if strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/html") {
			acceptEncoding := req.Header.Get("Accept-Encoding")
			if strings.Contains(acceptEncoding, "gzip") {
				newRes := newCompressWriter(res)
				origRes = newRes
				defer newRes.Close()
			}
		}

		contentEncoding := req.Header.Get("Content-Encoding")
		if strings.Contains(contentEncoding, "gzip") {
			newReader, err := newCompressReader(req.Body)
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
			req.Body = newReader
			defer newReader.Close()
		}

		h.ServeHTTP(origRes, req)
	}
}

type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	return c.zw.Close()
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
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
