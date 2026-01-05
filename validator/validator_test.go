// Package validator
// Copyright 2026 wantnotshould. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
package validator

import (
	"maps"
	"testing"
)

type User struct {
	Name  string `json:"name" validate:"required,len=5"`
	Age   int    `json:"age" validate:"required,min=18,max=100"`
	Email string `json:"email" validate:"required,email"`
}

type Address struct {
	Street string `json:"street" validate:"required"`
	City   string `json:"city" validate:"required"`
}

type UserAddress struct {
	Name    string  `json:"name" validate:"required,len=5"`
	Age     int     `json:"age" validate:"required,min=18,max=100"`
	Email   string  `json:"email" validate:"required,email"`
	Address Address `json:"address" validate:"required"`
}

func TestValidateStruct(t *testing.T) {
	validator := New()

	tests := []struct {
		name     string
		input    any
		expected map[string][]string
	}{
		{
			name: "valid user",
			input: &User{
				Name:  "Perry",
				Age:   25,
				Email: "perry@example.com",
			},
			expected: map[string][]string{},
		},
		{
			name: "missing required name",
			input: &User{
				Name:  "",
				Age:   25,
				Email: "perry@example.com",
			},
			expected: map[string][]string{
				"name": {"This field is required"},
			},
		},
		{
			name: "invalid name length",
			input: &User{
				Name:  "Al",
				Age:   25,
				Email: "perry@example.com",
			},
			expected: map[string][]string{
				"name": {"This field must be exactly 5 characters"},
			},
		},
		{
			name: "invalid age",
			input: &User{
				Name:  "Perry",
				Age:   17,
				Email: "perry@example.com",
			},
			expected: map[string][]string{
				"age": {"This field must be at least 18"},
			},
		},
		{
			name: "invalid email",
			input: &User{
				Name:  "Perry",
				Age:   25,
				Email: "invalid-email",
			},
			expected: map[string][]string{
				"email": {"This field must be a valid email address"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validator.ValidateStruct(tt.input)
			actual := map[string][]string{}
			maps.Copy(actual, errs)

			if !equalErrors(actual, tt.expected) {
				t.Errorf("expected %v, but got %v", tt.expected, actual)
			}
		})
	}
}

func TestRegexCache(t *testing.T) {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

	re1, err := getCachedRegex(pattern)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	re2, err := getCachedRegex(pattern)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if re1 != re2 {
		t.Error("expected regex to be cached, but it was not")
	}
}

func TestRegexMatch(t *testing.T) {
	tests := []struct {
		email    string
		expected bool
	}{
		{"valid@example.com", true},
		{"invalid-email", false},
		{"another.valid@example.co", true},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			result := isValidEmail(tt.email)
			if result != tt.expected {
				t.Errorf("expected %v for %s, but got %v", tt.expected, tt.email, result)
			}
		})
	}
}

func TestValidateStructWithNested(t *testing.T) {
	validator := New()

	tests := []struct {
		name     string
		input    any
		expected map[string][]string
	}{
		{
			name: "valid user with address",
			input: &UserAddress{
				Name:  "Perry",
				Age:   25,
				Email: "perry@example.com",
				Address: Address{
					Street: "123 Main St",
					City:   "Wonderland",
				},
			},
			expected: map[string][]string{}, // 无错误
		},
		{
			name: "missing required address",
			input: &UserAddress{
				Name:  "Perry",
				Age:   25,
				Email: "perry@example.com",
			},
			expected: map[string][]string{
				"address.city":   {"This field is required"},
				"address.street": {"This field is required"},
			},
		},
		{
			name: "invalid email and address",
			input: &UserAddress{
				Name:  "Perry",
				Age:   25,
				Email: "invalid-email",
				Address: Address{
					Street: "",
					City:   "",
				},
			},
			expected: map[string][]string{
				"email":          {"This field must be a valid email address"},
				"address.street": {"This field is required"},
				"address.city":   {"This field is required"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validator.ValidateStruct(tt.input)
			actual := map[string][]string{}
			maps.Copy(actual, errs)

			if !equalErrors(actual, tt.expected) {
				t.Errorf("expected %v, but got %v", tt.expected, actual)
			}
		})
	}
}
