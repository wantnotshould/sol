// Package sol
// Copyright 2025 wantnotshould. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
package sol

import (
	"fmt"
	"maps"
	"net/http"
	"strings"
	"sync"
)

type router interface {
	http.Handler
	GET(path string, handlers ...HandlerFunc)
	POST(path string, handlers ...HandlerFunc)
	PUT(path string, handlers ...HandlerFunc)
	DELETE(path string, handlers ...HandlerFunc)
	PATCH(path string, handlers ...HandlerFunc)
	OPTIONS(path string, handlers ...HandlerFunc)
	HEAD(path string, handlers ...HandlerFunc)

	Group(prefix string, middlewares ...HandlerFunc) *group
	Use(middlewares ...HandlerFunc)
	NotFound(handler HandlerFunc)
}

// node represents a radix tree node.
// https://en.wikipedia.org/wiki/Radix_tree
type node struct {
	children   map[string]*node
	paramChild *node
	handlers   []HandlerFunc
	isEnd      bool
	paramName  string
}

// routerImpl router implementation
type routerImpl struct {
	// trees method -> root node
	trees       map[string]*node
	middlewares []HandlerFunc
	notFound    HandlerFunc
	pool        sync.Pool
}

type group struct {
	prefix      string
	middlewares []HandlerFunc
	parent      *group
	router      *routerImpl
}

func newRouter() router {
	r := &routerImpl{
		trees: make(map[string]*node),
		notFound: func(c *Context) {
			c.Writer.WriteHeader(http.StatusNotFound)
			c.Writer.Write([]byte("404 page not found\n"))
		},
	}
	r.pool.New = func() any {
		return &Context{
			params: make(map[string]string, 4),
			data:   make(map[string]any, 10),
		}
	}
	return r
}

func normalizePath(path string) string {
	if path == "" {
		return "/"
	}

	path = strings.TrimSpace(path)

	// Run the loop first.
	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	if path != "/" {
		path = strings.TrimSuffix(path, "/")
	}

	return path
}

func (r *routerImpl) getTree(method string) *node {
	if r.trees[method] == nil {
		r.trees[method] = &node{
			children: make(map[string]*node),
		}
	}
	return r.trees[method]
}

func (r *routerImpl) insert(method, path string, combined []HandlerFunc) {
	path = normalizePath(path)
	root := r.getTree(method)

	if path == "/" {
		root.isEnd = true
		root.handlers = combined
		return
	}

	segments := strings.Split(path[1:], "/")
	cur := root

	for _, segment := range segments {
		isParam := segment[0] == ':'
		var child *node

		if isParam {
			paramName := segment[1:]
			if cur.paramChild != nil {
				if cur.paramChild.paramName != paramName {
					panic(fmt.Sprintf(
						"cannot register '%s': parameter name ':%s' conflicts with existing ':%s' in previously registered path",
						path, paramName, cur.paramChild.paramName,
					))
				}
			} else {
				cur.paramChild = &node{
					paramName: paramName,
				}
			}
			child = cur.paramChild
		} else {
			if cur.children == nil {
				cur.children = make(map[string]*node)
			}

			if _, ok := cur.children[segment]; !ok {
				cur.children[segment] = &node{
					children: make(map[string]*node),
				}
			}
			child = cur.children[segment]
		}

		cur = child
	}

	// At this point, len(segments) must be greater than 0
	cur.isEnd = true
	cur.handlers = combined
}

func (r *routerImpl) search(method, path string) ([]HandlerFunc, map[string]string) {
	path = normalizePath(path)
	root := r.trees[method]
	if root == nil {
		return nil, nil
	}

	if path == "/" {
		if root.isEnd {
			return root.handlers, nil
		}
		return nil, nil
	}

	segments := strings.Split(path[1:], "/")
	params := make(map[string]string)
	cur := root

	for _, segment := range segments {
		if cur.children != nil {
			if child, ok := cur.children[segment]; ok {
				cur = child
				continue
			}
		}

		if cur.paramChild != nil {
			cur = cur.paramChild
			params[cur.paramName] = segment
			continue
		}

		return nil, nil
	}

	if cur.isEnd {
		return cur.handlers, params
	}

	return nil, nil
}

func (r *routerImpl) addRoute(method, path string, middlewares, handlers []HandlerFunc) {
	// If middlewares is nil, use an empty slice instead.
	if middlewares == nil {
		middlewares = []HandlerFunc{}
	}

	combined := make([]HandlerFunc, 0, len(middlewares)+len(handlers))
	combined = append(combined, middlewares...)
	combined = append(combined, handlers...)

	r.insert(method, path, combined)
}

func (r *routerImpl) GET(path string, h ...HandlerFunc) {
	r.addRoute(http.MethodGet, path, r.middlewares, h)
}
func (r *routerImpl) POST(path string, h ...HandlerFunc) {
	r.addRoute(http.MethodPost, path, r.middlewares, h)
}
func (r *routerImpl) PUT(path string, h ...HandlerFunc) {
	r.addRoute(http.MethodPut, path, r.middlewares, h)
}
func (r *routerImpl) DELETE(path string, h ...HandlerFunc) {
	r.addRoute(http.MethodDelete, path, r.middlewares, h)
}
func (r *routerImpl) PATCH(path string, h ...HandlerFunc) {
	r.addRoute(http.MethodPatch, path, r.middlewares, h)
}
func (r *routerImpl) OPTIONS(path string, h ...HandlerFunc) {
	r.addRoute(http.MethodOptions, path, r.middlewares, h)
}
func (r *routerImpl) HEAD(path string, h ...HandlerFunc) {
	r.addRoute(http.MethodHead, path, r.middlewares, h)
}

func (r *routerImpl) Use(m ...HandlerFunc) {
	r.middlewares = append(r.middlewares, m...)
}

func (r *routerImpl) Group(prefix string, m ...HandlerFunc) *group {
	return &group{
		prefix:      normalizePath(prefix),
		middlewares: m,
		router:      r,
	}
}

func (r *routerImpl) acquireCtx(w http.ResponseWriter, req *http.Request, h []HandlerFunc) *Context {
	ctx := r.pool.Get().(*Context)
	ctx.Writer = w
	ctx.Request = req
	ctx.handlers = h
	ctx.index = -1
	ctx.aborted = false
	clear(ctx.params)
	clear(ctx.data)

	return ctx
}

func (r *routerImpl) releaseCtx(ctx *Context) {
	ctx.handlers = nil
	ctx.Writer = nil
	ctx.Request = nil
	r.pool.Put(ctx)
}

func (r *routerImpl) NotFound(handler HandlerFunc) {
	if handler == nil {
		handler = func(c *Context) {
			c.Writer.WriteHeader(http.StatusNotFound)
			c.Writer.Write([]byte("404 page not found\n"))
		}
	}
	r.notFound = handler
}

func (r *routerImpl) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	handlers, params := r.search(req.Method, req.URL.Path)
	if handlers == nil {
		ctx := r.acquireCtx(w, req, []HandlerFunc{r.notFound})
		ctx.Next()
		r.releaseCtx(ctx)
		return
	}

	ctx := r.acquireCtx(w, req, handlers)
	maps.Copy(ctx.params, params)

	ctx.Next()
	r.releaseCtx(ctx)
}

func (g *group) collectMiddlewares() []HandlerFunc {
	var mids []HandlerFunc
	current := g
	for current != nil {
		mids = append(mids, current.middlewares...)
		current = current.parent
	}

	mids = append(mids, g.router.middlewares...)
	return mids
}

func (g *group) add(method, path string, h ...HandlerFunc) {
	fullPath := g.prefix
	if path = normalizePath(path); path != "/" {
		if !strings.HasSuffix(fullPath, "/") {
			fullPath += "/"
		}
		fullPath += strings.TrimPrefix(path, "/")
	}

	middlewares := g.collectMiddlewares()
	g.router.addRoute(method, fullPath, middlewares, h)
}

func (g *group) GET(path string, h ...HandlerFunc)     { g.add(http.MethodGet, path, h...) }
func (g *group) POST(path string, h ...HandlerFunc)    { g.add(http.MethodPost, path, h...) }
func (g *group) PUT(path string, h ...HandlerFunc)     { g.add(http.MethodPut, path, h...) }
func (g *group) DELETE(path string, h ...HandlerFunc)  { g.add(http.MethodDelete, path, h...) }
func (g *group) PATCH(path string, h ...HandlerFunc)   { g.add(http.MethodPatch, path, h...) }
func (g *group) OPTIONS(path string, h ...HandlerFunc) { g.add(http.MethodOptions, path, h...) }
func (g *group) HEAD(path string, h ...HandlerFunc)    { g.add(http.MethodHead, path, h...) }

func (g *group) Group(sub string, m ...HandlerFunc) *group {
	newPrefix := g.prefix
	if !strings.HasSuffix(newPrefix, "/") {
		newPrefix += "/"
	}
	newPrefix += strings.TrimPrefix(normalizePath(sub), "/")

	return &group{
		prefix:      newPrefix,
		middlewares: m,
		parent:      g,
		router:      g.router,
	}
}
