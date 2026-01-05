// Package binding
// Copyright 2026 wantnotshould. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
package binding

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"testing"

	"github.com/wantnotshould/sol"
)

type User struct {
	Name    string `form:"name" validate:"required,len=5"`
	Age     int    `form:"age" validate:"required,min=18,max=100"`
	Email   string `form:"email" validate:"required,email"`
	Address string `form:"address" validate:"required"`
}

type Address struct {
	Street string `form:"street" validate:"required"`
	City   string `form:"city" validate:"required"`
}

type UserAddress struct {
	User    User    `form:"user"`
	Address Address `form:"address"`
}

func TestFormBinding(t *testing.T) {
	body := "name=Perry&age=25&email=perry@example.com&address=Wonderland"

	c := &sol.Context{
		Request: &http.Request{
			Method: http.MethodPost,
			Header: http.Header{
				"Content-Type": []string{"application/x-www-form-urlencoded"},
			},
			Body:          io.NopCloser(bytes.NewReader([]byte(body))),
			ContentLength: int64(len(body)), // must
		},
	}

	user := &User{}
	err := Form(c, user)

	if err != nil {
		t.Fatalf("Form binding failed: %v", err)
	}

	// 断言...
	if user.Name != "Perry" || user.Age != 25 || user.Email != "perry@example.com" || user.Address != "Wonderland" {
		t.Errorf("User binding failed: %+v", user)
	}
}

func TestMultipartFormBinding(t *testing.T) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add fields
	writer.WriteField("name", "Perry")
	writer.WriteField("age", "25")
	writer.WriteField("email", "perry@example.com")
	writer.WriteField("address", "Wonderland")

	// Add file
	file, _ := writer.CreateFormFile("avatar", "avatar.png")
	file.Write([]byte("dummy file content"))

	writer.Close()

	// Create the request with the multipart data
	c := &sol.Context{
		Request: &http.Request{
			Method:        http.MethodPost,
			Header:        http.Header{"Content-Type": []string{writer.FormDataContentType()}},
			Body:          io.NopCloser(&buf),
			ContentLength: int64(buf.Len()),
		},
	}

	// Bind form and file data
	user := &User{}
	err := MultipartForm(c, user)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if user.Name != "Perry" {
		t.Errorf("Expected Perry, got %v", user.Name)
	}
	if user.Age != 25 {
		t.Errorf("Expected 25, got %v", user.Age)
	}
	if user.Email != "perry@example.com" {
		t.Errorf("Expected perry@example.com, got %v", user.Email)
	}
	if user.Address != "Wonderland" {
		t.Errorf("Expected Wonderland, got %v", user.Address)
	}
}

func TestJSONBinding(t *testing.T) {
	jsonBody := `{"name": "Perry", "age": 25, "email": "perry@example.com", "address": "Wonderland"}`

	c := &sol.Context{
		Request: &http.Request{
			Method: http.MethodPost,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Body:          io.NopCloser(bytes.NewReader([]byte(jsonBody))),
			ContentLength: int64(len(jsonBody)),
		},
	}

	user := &User{}
	err := JSON(c, user)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if user.Name != "Perry" {
		t.Errorf("Expected Perry, got %v", user.Name)
	}
	if user.Age != 25 {
		t.Errorf("Expected 25, got %v", user.Age)
	}
	if user.Email != "perry@example.com" {
		t.Errorf("Expected perry@example.com, got %v", user.Email)
	}
	if user.Address != "Wonderland" {
		t.Errorf("Expected Wonderland, got %v", user.Address)
	}
}

func TestXMLBinding(t *testing.T) {
	xmlData := `<User><Name>Perry</Name><Age>25</Age><Email>perry@example.com</Email><Address>Wonderland</Address></User>`
	c := &sol.Context{
		Request: &http.Request{
			Method: http.MethodPost,
			Body:   io.NopCloser(bytes.NewReader([]byte(xmlData))),
			Header: map[string][]string{
				"Content-Type": {"application/xml"},
			},
		},
	}

	user := &User{}
	err := XML(c, user)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if user.Name != "Perry" {
		t.Errorf("Expected Perry, got %v", user.Name)
	}
	if user.Age != 25 {
		t.Errorf("Expected 25, got %v", user.Age)
	}
	if user.Email != "perry@example.com" {
		t.Errorf("Expected perry@example.com, got %v", user.Email)
	}
	if user.Address != "Wonderland" {
		t.Errorf("Expected Wonderland, got %v", user.Address)
	}
}

func TestFormBindingWithInvalidData(t *testing.T) {
	c := &sol.Context{
		Request: &http.Request{
			Method: http.MethodPost,
			Body:   io.NopCloser(bytes.NewReader([]byte("name=Al&age=15&email=invalid-email"))),
		},
	}

	user := &User{}
	err := Form(c, user)

	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestJSONBindingWithNestedStruct(t *testing.T) {
	jsonBody := `{
        "user": {
            "name": "Perry",
            "age": 25,
            "email": "perry@example.com",
            "address": "Wonderland"
        },
        "address": {
            "street": "Main St",
            "city": "Wonderland"
        }
    }`

	c := &sol.Context{
		Request: &http.Request{
			Method: http.MethodPost,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Body:          io.NopCloser(bytes.NewReader([]byte(jsonBody))),
			ContentLength: int64(len(jsonBody)),
		},
	}

	userAddress := &UserAddress{}
	err := JSON(c, userAddress)

	if err != nil {
		t.Fatalf("JSON binding failed: %v", err)
	}

	if userAddress.User.Name != "Perry" {
		t.Errorf("Expected Perry, got %q", userAddress.User.Name)
	}
	if userAddress.User.Age != 25 {
		t.Errorf("Expected 25, got %d", userAddress.User.Age)
	}
	if userAddress.User.Email != "perry@example.com" {
		t.Errorf("Expected perry@example.com, got %q", userAddress.User.Email)
	}
	if userAddress.Address.Street != "Main St" {
		t.Errorf("Expected Main St, got %q", userAddress.Address.Street)
	}
	if userAddress.Address.City != "Wonderland" {
		t.Errorf("Expected Wonderland, got %q", userAddress.Address.City)
	}
}

func TestUnsupportedFileType(t *testing.T) {
	c := &sol.Context{
		Request: &http.Request{
			Method: http.MethodPost,
			Body:   io.NopCloser(bytes.NewReader([]byte("name=Perry"))),
		},
	}

	user := &User{}
	err := MultipartForm(c, user)

	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestMissingRequiredField(t *testing.T) {
	body := "name=Perry&age=25"

	c := &sol.Context{
		Request: &http.Request{
			Method:        http.MethodPost,
			Header:        http.Header{"Content-Type": []string{"application/x-www-form-urlencoded"}},
			Body:          io.NopCloser(bytes.NewReader([]byte(body))),
			ContentLength: int64(len(body)),
		},
	}

	user := &User{}
	err := Form(c, user)

	if err != nil {
		t.Fatalf("Binding should succeed even with missing fields: %v", err)
	}

	if user.Name != "Perry" {
		t.Errorf("Expected Name = Perry, got %q", user.Name)
	}
	if user.Age != 25 {
		t.Errorf("Expected Age = 25, got %d", user.Age)
	}

	if user.Email != "" {
		t.Errorf("Expected Email empty, got %q", user.Email)
	}
	if user.Address != "" {
		t.Errorf("Expected Address empty, got %q", user.Address)
	}
}
