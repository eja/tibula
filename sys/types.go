// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package sys

type TypeConfig struct {
	DbType        string `json:"db_type"`
	DbName        string `json:"db_name"`
	DbUser        string `json:"db_user"`
	DbPass        string `json:"db_pass"`
	DbHost        string `json:"db_host"`
	DbPort        int    `json:"db_port"`
	WebHost       string `json:"web_host"`
	WebPort       int    `json:"web_port"`
	WebPath       string `json:"web_path"`
	WebTlsPublic  string `json:"web_tls_public"`
	WebTlsPrivate string `json:"web_tls_private"`
	Start         bool   `json:"start"`
	Setup         bool   `json:"setup"`
	SetupUser     string `json:"setup_user"`
	SetupPass     string `json:"setup_pass"`
	SetupPath     string `json:"setup_path"`
	ConfigFile    string `json:"config_file"`
	Language      string `json:"language"`
	LogLevel      int    `json:"log_level"`
}
