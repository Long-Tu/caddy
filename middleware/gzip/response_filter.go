package gzip

import (
	"net/http"
	"strconv"
)

// ResponseFilter determines if the response should be gzipped.
type ResponseFilter interface {
	ShouldCompress(http.ResponseWriter) bool
}

// LengthFilter is ResponseFilter for minimum content length.
type LengthFilter int64

// ShouldCompress returns if content length is greater than or
// equals to minimum length.
func (l LengthFilter) ShouldCompress(w http.ResponseWriter) bool {
	contentLength := (w.Header().Get("Content-Length"))
	length, err := strconv.ParseInt(contentLength, 10, 64)
	if err != nil || length == 0 {
		return false
	}
	return l != 0 && int64(l) <= length
}

// ResponseFilterWriter validates ResponseFilters. It writes
// gzip compressed data if ResponseFilters are satisfied or
// uncompressed data otherwise.
type ResponseFilterWriter struct {
	filters        []ResponseFilter
	validated      bool
	shouldCompress bool
	gzipResponseWriter
}

// NewResponseFilterWriter creates and initializes a new ResponseFilterWriter.
func NewResponseFilterWriter(filters []ResponseFilter, gz gzipResponseWriter) *ResponseFilterWriter {
	return &ResponseFilterWriter{filters: filters, gzipResponseWriter: gz}
}

// Write wraps underlying Write method and compresses if filters
// are satisfied
func (r *ResponseFilterWriter) Write(b []byte) (int, error) {
	// One time validation to determine if compression should
	// be used or not.
	if !r.validated {
		r.shouldCompress = true
		for _, filter := range r.filters {
			if !filter.ShouldCompress(r) {
				r.shouldCompress = false
				break
			}
		}
		r.validated = true
	}

	if r.shouldCompress {
		return r.gzipResponseWriter.Write(b)
	}
	return r.ResponseWriter.Write(b)
}
