// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package sys

import (
	"strconv"
)

type TypeConfig struct {
	DbType        string `json:"db_type,omitempty"`
	DbName        string `json:"db_name,omitempty"`
	DbUser        string `json:"db_user,omitempty"`
	DbPass        string `json:"db_pass,omitempty"`
	DbHost        string `json:"db_host,omitempty"`
	DbPort        int    `json:"db_port,omitempty"`
	DbSetupUser   string `json:"db_setup_user,omitempty"`
	DbSetupPass   string `json:"db_setup_pass,omitempty"`
	DbSetupPath   string `json:"db_setup_path,omitempty"`
	WebHost       string `json:"web_host,omitempty"`
	WebPort       int    `json:"web_port,omitempty"`
	WebPath       string `json:"web_path,omitempty"`
	WebTlsPublic  string `json:"web_tls_public,omitempty"`
	WebTlsPrivate string `json:"web_tls_private,omitempty"`
	ConfigFile    string `json:"config_file,omitempty"`
	Language      string `json:"language,omitempty"`
	LogLevel      int    `json:"log_level,omitempty"`
	LogFile       string `json:"log_file,omitempty"`
	GoogleSsoId   string `json:"google_sso_id,omitempty"`
}

type TypeCommand struct {
	Start   bool `json:"start,omitempty"`
	DbSetup bool `json:"db_setup,omitempty"`
	Wizard  bool `json:"wizard,omitempty"`
	Help    bool `json:"help,omitempty"`
}

func String(nameValue interface{}) string {
	switch v := nameValue.(type) {
	case string:
		return v
	case int, int64, float32, float64:
		return strconv.FormatInt(Number(nameValue), 10)
	case []uint8:
		return string(v)
	default:
		return ""
	}
}

func Number(nameValue interface{}) int64 {
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
		value := String(v)
		num, err := strconv.ParseInt(value, 10, 64)
		if err == nil {
			return num
		}
	}
	return 0
}

func Float(nameValue interface{}) float64 {
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

func Bool(value interface{}) bool {
	return Number(value) > 0
}
