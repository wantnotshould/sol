// Package sol
// Copyright 2025 wantnotshould. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
package sol

import (
	"log"
	"net/http"
	"runtime/debug"
)

func Recover() HandlerFunc {
	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := string(debug.Stack())
				log.Printf("[PANIC] %v\n%s", err, stack)

				http.Error(c.Writer, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}
