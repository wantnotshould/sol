// Package sol
// Copyright 2025 wantnotshould. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
package sol

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
)

type HandlerFunc func(*Context)

type Context struct {
	Request *http.Request
	Writer  http.ResponseWriter

	params map[string]string
	// data stores custom data for the request
	data map[string]any

	index    int8
	handlers []HandlerFunc
	aborted  bool

	// mu protects data map
	mu sync.RWMutex
}

// Context returns the request's context
func (c *Context) Context() context.Context {
	return c.Request.Context()
}

// Header returns the value of a request header.
func (c *Context) Header(key string) string {
	return c.Request.Header.Get(key)
}

// SetHeader sets a response header.
func (c *Context) SetHeader(key, value string) {
	c.Writer.Header().Set(key, value)
}

// Status sets the HTTP status code (does not write headers yet).
func (c *Context) Status(code int) {
	c.Writer.WriteHeader(code)
}

// SetCookie sets a cookie in the response.
func (c *Context) SetCookie(cookie *http.Cookie) {
	c.Writer.Header().Add("Set-Cookie", cookie.String())
}

// Cookie gets the value of a named cookie from the request.
// Returns empty string and error if not found.
func (c *Context) Cookie(name string) (string, error) {
	cookie, err := c.Request.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// DeleteCookie removes a cookie by setting it to expired.
func (c *Context) DeleteCookie(name string) {
	c.SetCookie(&http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// Host to get the host of the request
func (c *Context) Host() string {
	return c.Request.Host
}

// URL to get the full URL (scheme://host/path?query)
func (c *Context) URL() string {
	return c.Scheme() + "://" + c.Host() + c.Request.URL.Path
}

// Scheme to get the scheme of the request
func (c *Context) Scheme() string {
	if c.Request.TLS != nil {
		return "https"
	}
	return "http"
}

// Path to get the full normalized path of the request
func (c *Context) Path() string {
	return c.Request.URL.Path
}

// Method to get the HTTP method of the request
func (c *Context) Method() string {
	return c.Request.Method
}

// Param returns the value of a named route parameter.
func (c *Context) Param(key string) string {
	if c.params == nil {
		return ""
	}
	return c.params[key]
}

// Params returns the Context params.
func (c *Context) Params() map[string]string {
	return c.params
}

// QueryParam returns the first value for the named query parameter.
func (c *Context) QueryParam(key string) string {
	return c.Request.URL.Query().Get(key)
}

// QueryAll returns the full parsed query values.
func (c *Context) QueryAll() url.Values {
	return c.Request.URL.Query()
}

// Set stores a value in the request context.
func (c *Context) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.data == nil {
		c.data = make(map[string]any)
	}
	c.data[key] = value
}

// Get retrieves a value from the request context.
func (c *Context) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.data == nil {
		return nil, false
	}
	v, ok := c.data[key]
	return v, ok
}

// GetString is a convenience wrapper to retrieve and assert a string value.
func (c *Context) GetString(key string) (string, bool) {
	if v, ok := c.Get(key); ok {
		s, ok := v.(string)
		return s, ok
	}
	return "", false
}

// Delete removes a value from the context by its key.
func (c *Context) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.data != nil {
		delete(c.data, key)
	}
}

// Next invokes the next handler in the chain.
func (c *Context) Next() {
	// If already aborted or request context is done, stop processing
	if c.aborted {
		return
	}

	c.index++

	for c.index < int8(len(c.handlers)) {
		if c.aborted {
			return
		}

		if c.Request.Context().Err() != nil {
			return
		}

		c.handlers[c.index](c)
		c.index++
	}
}

// Abort stops execution of remaining handlers.
func (c *Context) Abort() {
	c.aborted = true
}

// IsAborted reports whether the handler chain has been aborted.
func (c *Context) IsAborted() bool {
	return c.aborted
}

func (c *Context) String(status int, format string, values ...any) {
	c.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.Writer.WriteHeader(status)
	msg := fmt.Sprintf(format, values...)
	c.Writer.Write([]byte(msg))
}

func (c *Context) JSON(status int, obj any) {
	c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	c.Writer.WriteHeader(status)

	data, err := json.Marshal(obj)
	if err != nil {
		http.Error(c.Writer, `{"error":"json marshal failed"}`, 500)
		return
	}
	c.Writer.Write(data)
}

func (c *Context) HTML(status int, html string) {
	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Writer.WriteHeader(status)
	c.Writer.Write([]byte(html))
}
