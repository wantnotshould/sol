// Package validator
// Copyright 2026 wantnotshould. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
package validator

import (
	"reflect"
	"strconv"
	"strings"
)

type Validator struct{}

func New() *Validator {
	return &Validator{}
}

func (v *Validator) ValidateStruct(obj any) ValidationErrors {
	errs := make(ValidationErrors)

	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		errs.Add("", "must be a struct or struct pointer")
		return errs
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		if !fieldVal.CanInterface() {
			continue
		}

		tag := field.Tag.Get("validate")
		if tag == "" {
			continue
		}

		fieldName := field.Tag.Get("json")
		if fieldName == "" || fieldName == "-" {
			fieldName = strings.ToLower(field.Name)
		}

		rules := ParseTag(tag)
		for _, rule := range rules {
			if rule.Name == "required" && isEmpty(fieldVal.Interface()) {
				errs.Add(fieldName, GetMessage("required", nil))
				break
			}

			if errMsg := v.checkRule(fieldVal.Interface(), rule); errMsg != "" {
				errs.Add(fieldName, errMsg)
			}
		}
	}

	return errs
}

func (v *Validator) checkRule(value any, rule Rule) string {
	switch rule.Name {
	case "required":
		if isEmpty(value) {
			return GetMessage("required", nil)
		}
	case "min":
		return checkMin(value, rule.Param)
	case "max":
		return checkMax(value, rule.Param)
	case "len":
		return checkLen(value, rule.Param)
	case "gt":
		return checkGt(value, rule.Param)
	case "gte":
		return checkGte(value, rule.Param)
	case "lt":
		return checkLt(value, rule.Param)
	case "lte":
		return checkLte(value, rule.Param)
	case "email":
		if str, ok := value.(string); ok && str != "" {
			if !isValidEmail(str) {
				return GetMessage("email", nil)
			}
		}
	case "regex":
		if str, ok := value.(string); ok && str != "" {
			if rule.Param == "" {
				return "regex rule parameter is empty"
			}

			re, err := getCachedRegex(rule.Param)
			if err != nil {
				return err.Error()
			}

			if !re.MatchString(str) {
				return GetMessage("regex", nil)
			}
		}
	}
	return ""
}

func isEmpty(value any) bool {
	if value == nil {
		return true
	}
	switch v := reflect.ValueOf(value); v.Kind() {
	case reflect.String, reflect.Array, reflect.Map, reflect.Slice:
		return v.Len() == 0
	case reflect.Ptr:
		return v.IsNil()
	}
	return false
}

func toFloat(value any) (float64, bool) {
	switch v := value.(type) {
	case int, int8, int16, int32, int64:
		return float64(reflect.ValueOf(v).Int()), true
	case uint, uint8, uint16, uint32, uint64:
		return float64(reflect.ValueOf(v).Uint()), true
	case float32, float64:
		return reflect.ValueOf(v).Float(), true
	case string:
		f, err := strconv.ParseFloat(v, 64)
		return f, err == nil
	}
	return 0, false
}

func toInt(value any) (int, bool) {
	switch v := value.(type) {
	case int, int8, int16, int32, int64:
		return int(reflect.ValueOf(v).Int()), true
	case uint, uint8, uint16, uint32, uint64:
		return int(reflect.ValueOf(v).Uint()), true
	case string:
		i, err := strconv.Atoi(v)
		return i, err == nil
	}
	return 0, false
}

func checkMin(value any, param string) string {
	p, err := strconv.ParseFloat(param, 64)
	if err != nil {
		return ""
	}
	if i, ok := toInt(value); ok && float64(i) < p {
		return GetMessage("min", int(p))
	}
	if f, ok := toFloat(value); ok && f < p {
		return GetMessage("min", int(p))
	}
	if s, ok := value.(string); ok && len(s) < int(p) {
		return GetMessage("min", int(p))
	}
	return ""
}

func checkMax(value any, param string) string {
	p, err := strconv.ParseFloat(param, 64)
	if err != nil {
		return ""
	}
	if i, ok := toInt(value); ok && float64(i) > p {
		return GetMessage("max", int(p))
	}
	if f, ok := toFloat(value); ok && f > p {
		return GetMessage("max", int(p))
	}
	if s, ok := value.(string); ok && len(s) > int(p) {
		return GetMessage("max", int(p))
	}
	return ""
}

func checkLen(value any, param string) string {
	p, err := strconv.Atoi(param)
	if err != nil {
		return "Invalid length parameter"
	}

	switch v := value.(type) {
	case string:
		if len(v) != p {
			return GetMessage("len", p)
		}
	default:
		return "Unsupported type for len check"
	}

	return ""
}

func checkGt(value any, param string) string {
	p, err := strconv.ParseFloat(param, 64)
	if err != nil {
		return ""
	}
	if f, ok := toFloat(value); ok && f <= p {
		return GetMessage("gt", p)
	}
	return ""
}

func checkGte(value any, param string) string {
	p, err := strconv.ParseFloat(param, 64)
	if err != nil {
		return ""
	}
	if f, ok := toFloat(value); ok && f < p {
		return GetMessage("gte", p)
	}
	return ""
}

func checkLt(value any, param string) string {
	p, err := strconv.ParseFloat(param, 64)
	if err != nil {
		return ""
	}
	if f, ok := toFloat(value); ok && f >= p {
		return GetMessage("lt", p)
	}
	return ""
}

func checkLte(value any, param string) string {
	p, err := strconv.ParseFloat(param, 64)
	if err != nil {
		return ""
	}
	if f, ok := toFloat(value); ok && f > p {
		return GetMessage("lte", p)
	}
	return ""
}
