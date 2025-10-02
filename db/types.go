// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package db

import (
	"strconv"
)

type TypeRun struct {
	Changes int64
	LastId  int64
}

type TypeRow map[string]string

type TypeRows []map[string]string

type TypeSelect struct {
	Key   string
	Value string
}

func (session *TypeSession) String(nameValue interface{}) string {
	switch v := nameValue.(type) {
	case string:
		return v
	case int, int64, float32, float64:
		return strconv.FormatInt(session.Number(nameValue), 10)
	case []uint8:
		return string(v)
	default:
		return ""
	}
}

func (session *TypeSession) Number(nameValue interface{}) int64 {
	switch v := nameValue.(type) {
	case int:
		return int64(v)
	case int64:
		return v
	case float32:
		return int64(v)
	case float64:
		return int64(v)
	default:
		value := session.String(v)
		num, err := strconv.ParseInt(value, 10, 64)
		if err == nil {
			return num
		}
	}
	return 0
}

func (session *TypeSession) Float(nameValue interface{}) float64 {
	switch v := nameValue.(type) {
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case float32:
		return float64(v)
	case float64:
		return v
	case string:
		num, err := strconv.ParseFloat(v, 64)
		if err == nil {
			return num
		}
	}
	return 0.0
}

func (session *TypeSession) Bool(value interface{}) bool {
	return session.Number(value) > 0
}
