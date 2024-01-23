// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package sys

type TypeConfig struct {
	DbType        string `json:"db_type,omitempty"`
	DbName        string `json:"db_name,omitempty"`
	DbUser        string `json:"db_user,omitempty"`
	DbPass        string `json:"db_pass,omitempty"`
	DbHost        string `json:"db_host,omitempty"`
	DbPort        int    `json:"db_port,omitempty"`
	WebHost       string `json:"web_host,omitempty"`
	WebPort       int    `json:"web_port,omitempty"`
	WebPath       string `json:"web_path,omitempty"`
	WebTlsPublic  string `json:"web_tls_public,omitempty"`
	WebTlsPrivate string `json:"web_tls_private,omitempty"`
	Start         bool   `json:"start,omitempty"`
	Setup         bool   `json:"setup,omitempty"`
	SetupUser     string `json:"setup_user,omitempty"`
	SetupPass     string `json:"setup_pass,omitempty"`
	SetupPath     string `json:"setup_path,omitempty"`
	ConfigFile    string `json:"config_file,omitempty"`
	Language      string `json:"language,omitempty"`
	LogLevel      int    `json:"log_level,omitempty"`
}
