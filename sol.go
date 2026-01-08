// Package sol
// Copyright 2025 wantnotshould. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
package sol

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Sol struct {
	router
	server   *http.Server
	stop     chan struct{}
	stopOnce sync.Once
}

func New() *Sol {
	router := newRouter()
	sl := &Sol{
		router: router,
		stop:   make(chan struct{}),
		server: &http.Server{
			ReadHeaderTimeout: 10 * time.Second,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       90 * time.Second,
		},
	}

	sl.server.Handler = sl
	sl.Use(Recover())

	return sl
}

func (sl *Sol) WithLogger() *Sol {
	sl.Use(Logger())
	return sl
}

func (sl *Sol) WithServer(server *http.Server) *Sol {
	if server != nil {
		if server.Handler == nil {
			server.Handler = sl
		}
		sl.server = server
	}
	return sl
}

func formatListenURL(addr string, isTLS bool) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}

	scheme := "http"
	if isTLS {
		scheme = "https"
	}

	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "localhost"
	}

	if (isTLS && port == "443") || (!isTLS && port == "80") {
		return fmt.Sprintf("%s://%s", scheme, host)
	}

	return fmt.Sprintf("%s://%s:%s", scheme, host, port)
}

func (sl *Sol) Run(addr ...string) {
	runAddr := ":23719"

	if len(addr) > 0 && addr[0] != "" {
		runAddr = addr[0]
	} else if env := os.Getenv("SOL_ADDR"); env != "" {
		runAddr = env
	}

	sl.server.Addr = runAddr
	log.Printf("🌌 Sol starting on %s", formatListenURL(runAddr, false))

	go func() {
		if err := sl.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	sl.waitStopSignal()
}

func (sl *Sol) RunTLS(addr, certFile, keyFile string) {
	if addr == "" {
		addr = ":443"
	}

	sl.server.Addr = addr
	sl.server.TLSConfig = &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	log.Printf("🌌 Sol starting on %s", formatListenURL(addr, true))

	go func() {
		if err := sl.server.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
			log.Fatalf("TLS Server error: %v", err)
		}
	}()

	sl.waitStopSignal()
}

func (sl *Sol) waitStopSignal() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sl.stop:
		log.Println("Received Stop() call")
	case s := <-sig:
		log.Printf("Received signal: %v, shutting down gracefully...", s)
	}

	log.Println("Shutting down server, will timeout after 30 seconds...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := sl.server.Shutdown(ctx); err != nil {
		log.Printf("Forced shutdown: %v", err)
	} else {
		log.Println("Server stopped gracefully.")
	}
}

func (sl *Sol) Stop() {
	sl.stopOnce.Do(func() {
		close(sl.stop)
	})
}
