// Package validator
// Copyright 2026 wantnotshould. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
package validator

import "strings"

// Rule {name: "min", param: "18"}
type Rule struct {
	Name  string
	Param string
}

// ParseTag validate:"required,min=18,email" → []Rule
func ParseTag(tag string) []Rule {
	if tag == "" || tag == "-" {
		return nil
	}

	parts := strings.Split(tag, ",")
	rules := make([]Rule, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)

		if strings.Contains(part, "=") {
			kv := strings.SplitN(part, "=", 2)
			rules = append(rules, Rule{Name: kv[0], Param: kv[1]})
		} else {
			rules = append(rules, Rule{Name: part})
		}
	}
	return rules
}
