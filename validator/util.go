// Package validator
// Copyright 2026 wantnotshould. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
package validator

func equalErrors(a, b map[string][]string) bool {
	if len(a) != len(b) {
		return false
	}
	for key, aMsgs := range a {
		if bMsgs, exists := b[key]; !exists || len(aMsgs) != len(bMsgs) {
			return false
		} else {
			for i, msg := range aMsgs {
				if msg != bMsgs[i] {
					return false
				}
			}
		}
	}
	return true
}
