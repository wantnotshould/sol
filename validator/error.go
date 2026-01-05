// Package validator
// Copyright 2026 wantnotshould. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
package validator

import "fmt"

type ValidationErrors map[string][]string

func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return ""
	}
	return "validation failed"
}

func (ve ValidationErrors) Add(field, message string) {
	if ve[field] == nil {
		ve[field] = make([]string, 0, 1)
	}
	ve[field] = append(ve[field], message)
}

func GetMessage(rule string, param any) string {
	messages := map[string]string{
		"required": "This field is required",
		"min":      "This field must be at least %v",
		"max":      "This field must be at most %v",
		"len":      "This field must be exactly %v characters",
		"gt":       "This field must be greater than %v",
		"gte":      "This field must be greater than or equal to %v",
		"lt":       "This field must be less than %v",
		"lte":      "This field must be less than or equal to %v",
		"email":    "This field must be a valid email address",
		"regex":    "This field format is invalid",
	}

	if msg, ok := messages[rule]; ok {
		if param != nil {
			return fmt.Sprintf(msg, param)
		}
		return msg
	}
	return "Invalid validation rule"
}
