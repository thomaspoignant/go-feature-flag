// This file is a temporary in-repo copy of github.com/pablor21/echo-etag.
//
// Upstream has no Echo v5 release yet, which blocks the relay proxy migration
// tracked in thomaspoignant/go-feature-flag#5294. The port is already open
// upstream:
//
//	https://github.com/pablor21/echo-etag/issues/13
//	https://github.com/pablor21/echo-etag/pull/14
//
// Once upstream tags a v5 release, DELETE this file and etag_test.go and go
// back to depending on github.com/pablor21/echo-etag/v5. The code below is kept
// deliberately faithful to upstream (only the exported names are prefixed, to
// avoid over-generic identifiers in this shared package) so that swap stays a
// clean removal rather than a re-port.
//
// Known upstream sharp edge, intentionally NOT fixed here so the copy stays
// faithful: writeRaw calls WriteHeader(hw.status), which panics if a handler
// returns nil having written neither a status nor a body (hw.status == 0). All
// three endpoints this middleware is mounted on always write a body, so it is
// not reachable in this repo.
//
// MIT License
//
// Copyright (c) 2023 Pablo Ramirez <pablo@pramirez.dev>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package middleware

import (
	"bytes"
	"crypto/sha1" //nolint:gosec // SHA-1 is used for ETag hashing, not security.
	"encoding/hex"
	"fmt"
	"hash"
	"hash/crc32"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// EtagConfig defines the config for the Etag middleware.
// Upstream name: etag.Config
type EtagConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper middleware.Skipper
	// Weak defines if the Etag is weak or strong.
	Weak bool
	// HashFn defines the hash function to use. Default is crc32q.
	HashFn func(config EtagConfig) hash.Hash
}

var (
	// DefaultEtagConfig is the default Etag middleware config.
	DefaultEtagConfig = EtagConfig{
		Skipper: middleware.DefaultSkipper,
		Weak:    true,
		HashFn: func(config EtagConfig) hash.Hash {
			if config.Weak {
				const crcPol = 0xD5828281
				crc32qTable := crc32.MakeTable(crcPol)
				return crc32.New(crc32qTable)
			}
			return sha1.New() //nolint:gosec // SHA-1 is used for ETag hashing, not security.
		},
	}
	normalizedETagName        = http.CanonicalHeaderKey("Etag")
	normalizedIfNoneMatchName = http.CanonicalHeaderKey("If-None-Match")
	weakPrefix                = "W/"
)

// Etag returns an Etag middleware using the default config.
func Etag() echo.MiddlewareFunc {
	c := DefaultEtagConfig
	return EtagWithConfig(c)
}

// EtagWithConfig returns an Etag middleware with config.
// Upstream name: etag.WithConfig
func EtagWithConfig(config EtagConfig) echo.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = DefaultEtagConfig.Skipper
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			skipper := config.Skipper
			if skipper == nil {
				skipper = DefaultEtagConfig.Skipper
			}

			if skipper(c) {
				return next(c)
			}

			// get the hash function
			hashFn := config.HashFn
			if hashFn == nil {
				hashFn = DefaultEtagConfig.HashFn
			}

			originalWriter := c.Response().Writer
			res := c.Response()
			req := c.Request()
			// ResponseWriter
			hw := bufferedWriter{rw: res.Writer, hash: hashFn(config), buf: bytes.NewBuffer(nil)}
			res.Writer = &hw
			err = next(c)
			// restore the original writer
			res.Writer = originalWriter
			if err != nil {
				return err
			}

			resHeader := res.Header()

			if hw.hash == nil ||
				resHeader.Get(normalizedETagName) != "" ||
				strconv.Itoa(hw.status)[0] != '2' ||
				hw.status == http.StatusNoContent ||
				hw.buf.Len() == 0 {
				writeRaw(originalWriter, hw)
				return //nolint:nakedret // kept faithful to upstream
			}

			etag := fmt.Sprintf("\"%v-%v\"", strconv.Itoa(hw.len),
				hex.EncodeToString(hw.hash.Sum(nil)))

			if config.Weak {
				etag = weakPrefix + etag
			}

			resHeader.Set(normalizedETagName, etag)

			ifNoneMatch := req.Header.Get(normalizedIfNoneMatchName) // get the If-None-Match header
			headerFresh := ifNoneMatch == etag || ifNoneMatch == weakPrefix+etag

			if headerFresh {
				originalWriter.WriteHeader(http.StatusNotModified)
				_, _ = originalWriter.Write(nil)
			} else {
				writeRaw(originalWriter, hw)
			}
			return //nolint:nakedret // kept faithful to upstream
		}
	}
}

// bufferedWriter is a wrapper around http.ResponseWriter that
// buffers the response and calculates the hash of the response.
type bufferedWriter struct { //nolint:recvcheck // mixed receivers kept faithful to upstream
	rw     http.ResponseWriter
	hash   hash.Hash
	buf    *bytes.Buffer
	len    int
	status int
}

// Header returns the header map that will be sent by the underlying writer.
func (hw bufferedWriter) Header() http.Header {
	return hw.rw.Header()
}

// WriteHeader sends an HTTP response header with the provided status code.
func (hw *bufferedWriter) WriteHeader(status int) {
	hw.status = status
}

// Write writes the data to the buffer to be sent as part of an HTTP reply.
func (hw *bufferedWriter) Write(b []byte) (int, error) {
	if hw.status == 0 {
		hw.status = http.StatusOK
	}
	// write to the buffer
	l, err := hw.buf.Write(b)
	if err != nil {
		return l, err
	}
	// write to the hash
	l, err = hw.hash.Write(b)
	hw.len += l
	return l, err
}

// writeRaw writes the buffered data to the underlying http.ResponseWriter.
func writeRaw(res http.ResponseWriter, hw bufferedWriter) {
	res.WriteHeader(hw.status)
	_, _ = res.Write(hw.buf.Bytes())
}
