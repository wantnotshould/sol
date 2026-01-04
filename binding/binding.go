// Package binding
// Copyright 2026 wantnotshould. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
package binding

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/wantnotshould/sol"
)

func Form(c *sol.Context, obj any) error {
	if err := c.Request.ParseForm(); err != nil {
		return fmt.Errorf("parse form error: %w", err)
	}
	values := c.Request.Form
	return bindFromValues(values, obj)
}

func MultipartForm(c *sol.Context, obj any) error {
	const maxMemory = 32 << 20 // 32 MB
	if err := c.Request.ParseMultipartForm(maxMemory); err != nil {
		return fmt.Errorf("parse multipart form error: %w", err)
	}

	if c.Request.MultipartForm != nil && c.Request.MultipartForm.Value != nil {
		if err := bindFromValues(c.Request.MultipartForm.Value, obj); err != nil {
			return err
		}
	}

	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Pointer || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("binding: obj must be pointer to struct")
	}
	v = v.Elem()
	t := v.Type()
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

		fileName := tag

		files := c.Request.MultipartForm.File[fileName]
		if len(files) == 0 {
			continue
		}

		if !fieldValue.CanSet() {
			continue
		}

		switch field.Type {
		case reflect.TypeFor[*multipart.FileHeader]():
			fieldValue.Set(reflect.ValueOf(files[0]))

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

func bindFromValues(values url.Values, obj any) error {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Pointer || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("binding: obj must be pointer to struct")
	}
	v = v.Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("form")
		if tag == "" || tag == "-" {
			continue
		}

		paramName := tag

		if strs, ok := values[paramName]; ok && len(strs) > 0 {
			value := strs[0]

			fieldValue := v.Field(i)
			if !fieldValue.CanSet() {
				continue
			}

			if err := setField(fieldValue, value); err != nil {
				return fmt.Errorf("bind %s=%s: %w", paramName, value, err)
			}
		}
	}
	return nil
}

func setField(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(u)
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(b)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(f)
	default:
		return fmt.Errorf("unsupported field type: %v", field.Kind())
	}
	return nil
}
