// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package cfg

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
)

var Version = "17.1.5"
var Options TypeConfig

type TypeConfig struct {
	DbType     string `json:"db_type"`
	DbName     string `json:"db_name"`
	DbUser     string `json:"db_user"`
	DbPass     string `json:"db_pass"`
	DbHost     string `json:"db_host"`
	DbPort     int    `json:"db_port"`
	WebHost    string `json:"web_host"`
	WebPort    int    `json:"web_port"`
	WebPath    string `json:"web_path"`
	Start      bool   `json:"start"`
	Setup      bool   `json:"setup"`
	SetupUser  string `json:"setup_user"`
	SetupPass  string `json:"setup_pass"`
	SetupPath  string `json:"setup_path"`
	ConfigFile string `json:"config_file"`
	Language   string `json:"language"`
	LogLevel   int    `json:"log_level"`
}

func Configure() {
	help := flag.Bool("help", false, "show this message")
	flag.StringVar(&Options.DbType, "db-type", "sqlite", "database type: sqlite/mysql")
	flag.StringVar(&Options.DbName, "db-name", "tibula.db", "database name or filename")
	flag.StringVar(&Options.DbUser, "db-user", "", "database username")
	flag.StringVar(&Options.DbPass, "db-pass", "", "database password")
	flag.StringVar(&Options.DbHost, "db-host", "", "database hostname")
	flag.IntVar(&Options.DbPort, "db-port", 3306, "database port")
	flag.StringVar(&Options.WebPath, "web-path", "", "web path")
	flag.StringVar(&Options.WebHost, "web-host", "localhost", "web listen address")
	flag.IntVar(&Options.WebPort, "web-port", 35248, "web listen port")
	flag.BoolVar(&Options.Setup, "setup", false, "initialize the database")
	flag.StringVar(&Options.SetupPath, "setup-path", "", "setup files path")
	flag.StringVar(&Options.SetupUser, "setup-user", "admin", "setup admin username")
	flag.StringVar(&Options.SetupPass, "setup-pass", "", "setup admin password")
	flag.BoolVar(&Options.Start, "start", false, "start the web service")
	flag.StringVar(&Options.ConfigFile, "config", "", "json config file")
	flag.StringVar(&Options.Language, "language", "en", "default language code")
	flag.IntVar(&Options.LogLevel, "log-level", 3, "set the log level (1-5): 1=Error, 2=Warn, 3=Info, 4=Debug, 5=Trace")
	flag.Parse()

	if Options.ConfigFile != "" {
		jsonData, err := os.ReadFile(Options.ConfigFile)
		if err != nil {
			log.Fatalf("Error reading configuration file: %v", err)
		}
		err = json.Unmarshal(jsonData, &Options)
		if err != nil {
			log.Fatalf("Error unmarshaling configuration file: %v", err)
		}
	}

	if *help {
		Help()
	}
}

func Help() {
	fmt.Println("Copyright:", "2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>")
	fmt.Println("Version:", Version)
	fmt.Println("Usage: tibula [options]")
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	fmt.Println()
}
