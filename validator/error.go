// Package validator
// Copyright 2026 wantnotshould. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
package validator

import "fmt"

type Language string

const (
	EN Language = "en"
	ZH Language = "zh"
)

var messages = map[Language]map[string]string{
	EN: {
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
	},
	ZH: {
		"required": "此字段是必填的",
		"min":      "此字段必须至少为 %v",
		"max":      "此字段不能超过 %v",
		"len":      "此字段必须恰好是 %v 个字符",
		"gt":       "此字段必须大于 %v",
		"gte":      "此字段必须大于或等于 %v",
		"lt":       "此字段必须小于 %v",
		"lte":      "此字段必须小于或等于 %v",
		"email":    "此字段必须是有效的电子邮件地址",
		"regex":    "此字段格式无效",
	},
}

var currentLanguage = EN

// SetLanguage sets the current language for validation messages
func SetLanguage(lang Language) {
	currentLanguage = lang
}

// ValidationErrors represents validation errors
type ValidationErrors map[string][]string

// Error implements the error interface
func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return ""
	}
	return "validation failed"
}

// Add adds a validation error for a given field
func (ve ValidationErrors) Add(field, message string) {
	if ve[field] == nil {
		ve[field] = make([]string, 0, 1)
	}
	ve[field] = append(ve[field], message)
}

// GetMessage returns the localized validation message for a given rule
func GetMessage(rule string, param any) string {
	// First try the current language
	if msg, ok := messages[currentLanguage][rule]; ok {
		if param != nil {
			return fmt.Sprintf(msg, param)
		}
		return msg
	}
	// If the rule is not found in the current language, fallback to the default language (EN)
	if msg, ok := messages[EN][rule]; ok {
		if param != nil {
			return fmt.Sprintf(msg, param)
		}
		return msg
	}
	// If still not found, return a generic message
	return "Invalid validation rule"
}
