// Package binding
// Copyright 2026 wantnotshould. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
package binding

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/wantnotshould/sol"
)

// Constants for max memory and supported content types
const maxMemory = 32 << 20 // 32 MB

// Form binds URL-encoded form data to the given Go struct.
func Form(c *sol.Context, obj any) error {
	if err := c.Request.ParseForm(); err != nil {
		return fmt.Errorf("parse form error: %w", err)
	}
	return bindFromValues(c.Request.Form, obj)
}

// MultipartForm binds multipart form data (including files) to the given Go struct.
func MultipartForm(c *sol.Context, obj any) error {
	if err := c.Request.ParseMultipartForm(maxMemory); err != nil {
		return fmt.Errorf("parse multipart form error: %w", err)
	}
	return bindMultipartFormData(c, obj)
}

// JSON binds JSON request body data to the given Go struct.
func JSON(c *sol.Context, obj any) error {
	contentType := c.Request.Header.Get("Content-Type")
	if !strings.Contains(strings.ToLower(contentType), "application/json") {
		return fmt.Errorf("json binding: Content-Type is not application/json, got %s", contentType)
	}

	if c.Request.Body == nil {
		return fmt.Errorf("json binding: request body is nil")
	}

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return fmt.Errorf("read request body error: %w", err)
	}
	if len(bodyBytes) == 0 {
		return fmt.Errorf("json binding: empty request body")
	}

	if err := json.Unmarshal(bodyBytes, obj); err != nil {
		return fmt.Errorf("json unmarshal error: %w", err)
	}

	return nil
}

// XML binds XML request body data to the given Go struct.
func XML(c *sol.Context, obj any) error {
	contentType := c.Request.Header.Get("Content-Type")
	lowerCT := strings.ToLower(contentType)
	if !strings.Contains(lowerCT, "application/xml") && !strings.Contains(lowerCT, "text/xml") {
		return fmt.Errorf("xml binding: Content-Type is not xml, got %s", contentType)
	}

	if c.Request.Body == nil {
		return fmt.Errorf("xml binding: request body is nil")
	}

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return fmt.Errorf("read request body error: %w", err)
	}
	if len(bodyBytes) == 0 {
		return fmt.Errorf("xml binding: empty request body")
	}

	if err := xml.Unmarshal(bodyBytes, obj); err != nil {
		return fmt.Errorf("xml unmarshal error: %w", err)
	}

	return nil
}

// bindFromValues binds form values to the struct based on the form tags.
func bindFromValues(values url.Values, obj any) error {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return fmt.Errorf("binding: obj must be a non-nil pointer")
	}
	if v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("binding: obj must be pointer to struct")
	}

	elem := v.Elem()

	for i := 0; i < elem.NumField(); i++ {
		field := elem.Type().Field(i)
		tag := field.Tag.Get("form")
		if tag == "" || tag == "-" {
			continue
		}

		if strs, ok := values[tag]; ok && len(strs) > 0 {
			value := strs[0]
			fieldValue := elem.Field(i)
			if !fieldValue.CanSet() {
				continue
			}
			if err := setField(fieldValue, value); err != nil {
				return fmt.Errorf("bind %s=%s: %w", tag, value, err)
			}
		}
	}
	return nil
}

// bindMultipartFormData binds multipart form data, including files, to the struct.
func bindMultipartFormData(c *sol.Context, obj any) error {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Pointer || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("binding: obj must be pointer to struct")
	}
	v = v.Elem()
	t := v.Type()

	if c.Request.MultipartForm != nil && c.Request.MultipartForm.Value != nil {
		if err := bindFromValues(c.Request.MultipartForm.Value, obj); err != nil {
			return err
		}
	}

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		if fieldValue.Kind() != reflect.Pointer && fieldValue.Kind() != reflect.Slice {
			continue
		}

		tag := field.Tag.Get("form")
		if tag == "" || tag == "-" {
			continue
		}

		files := c.Request.MultipartForm.File[tag]
		if len(files) == 0 {
			continue
		}

		if !fieldValue.CanSet() {
			continue
		}

		switch field.Type {
		case reflect.TypeFor[*multipart.FileHeader]():
			if len(files) > 0 {
				fieldValue.Set(reflect.ValueOf(files[0]))
			}

		case reflect.TypeFor[[]*multipart.FileHeader]():
			fileSlice := make([]*multipart.FileHeader, len(files))
			copy(fileSlice, files)
			fieldValue.Set(reflect.ValueOf(fileSlice))

		default:
			return fmt.Errorf("unsupported file field type: %s, only support *multipart.FileHeader or []*multipart.FileHeader", field.Type)
		}
	}

	return nil
}

// setField sets the value of a struct field based on its type.
func setField(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid int value: %w", err)
		}
		field.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid uint value: %w", err)
		}
		field.SetUint(u)
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid bool value: %w", err)
		}
		field.SetBool(b)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float value: %w", err)
		}
		field.SetFloat(f)
	default:
		return fmt.Errorf("unsupported field type: %v", field.Kind())
	}
	return nil
}
