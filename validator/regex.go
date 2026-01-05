// Package validator
// Copyright 2026 wantnotshould. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
package validator

import (
	"fmt"
	"regexp"
	"sync"
)

var (
	regexCache   = make(map[string]*regexp.Regexp)
	regexCacheMu sync.RWMutex

	emailRegex     *regexp.Regexp
	emailRegexOnce sync.Once
)

func getCachedRegex(pattern string) (*regexp.Regexp, error) {
	regexCacheMu.RLock()
	if re, ok := regexCache[pattern]; ok {
		regexCacheMu.RUnlock()
		return re, nil
	}
	regexCacheMu.RUnlock()

	regexCacheMu.Lock()
	defer regexCacheMu.Unlock()

	if re, ok := regexCache[pattern]; ok {
		return re, nil
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %s", pattern)
	}

	regexCache[pattern] = re
	return re, nil
}

func getEmailRegex() *regexp.Regexp {
	emailRegexOnce.Do(func() {
		emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	})
	return emailRegex
}

func isValidEmail(email string) bool {
	if email == "" {
		return false
	}
	return getEmailRegex().MatchString(email)
}
