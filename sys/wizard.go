// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package sys

import (
	"bufio"
	"errors"
	"fmt"
	"golang.org/x/term"
	"os"
	"strconv"
	"strings"
	"syscall"
)

func WizardPrompt(message string) string {
	fmt.Printf("%s: ", message)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text())
	}
	return ""
}

func WizardPassword(message string) string {
	fmt.Printf("%s: ", message)
	pass, _ := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	return strings.TrimSpace(string(pass))
}

func WizardSetup() error {
	var err error
	var tibulaDb = "tibula.db"
	var tibulaJson = ConfigFileName()

	if _, err := os.Stat(tibulaJson); err == nil {
		return errors.New("A configuration file already exists, remove it and try again.")
	}

	//admin
	setupUser := WizardPrompt("Choose an administrator username (admin)")
	if setupUser != "" {
		Options.DbSetupUser = setupUser
	}
	Options.DbSetupPass = WizardPassword("Enter the administrator password")
	if Options.DbSetupPass == "" {
		return errors.New("Password cannot be empty.")
	}
	passConfirm := WizardPassword("Repeat password")
	if passConfirm != Options.DbSetupPass {
		return errors.New("Passwords do not match.")
	}
	//web
	webHost := WizardPrompt("Web address to listen to (localhost)")
	if webHost != "" {
		Options.WebHost = webHost
	}
	webPort := WizardPrompt("Web port to listen to (35248)")
	if webPort != "" {
		Options.WebPort, err = strconv.Atoi(webPort)
		if err != nil {
			return fmt.Errorf("invalid port number: %s", webPort)
		}
	}
	webPrivate := WizardPrompt("Web https private certificate path (none)")
	if webPrivate != "" {
		Options.WebTlsPrivate = webPrivate
		webPublic := WizardPrompt("Web https public certificate path (none)")
		if webPublic != "" {
			Options.WebTlsPublic = webPublic
		}
	}
	//db
	dbType := WizardPrompt("Choose a database engine between sqlite and mysql (sqlite)")
	if dbType == "mysql" {
		Options.DbType = dbType
		Options.DbName = WizardPrompt("Database name")
		Options.DbUser = WizardPrompt("Database username")
		Options.DbPass = WizardPassword("Database password")
		dbHost := WizardPrompt("Database hostname (localhost)")
		if dbHost != "" {
			Options.DbHost = dbHost
		}
		dbPort := WizardPrompt("Database port (3306)")
		if dbPort != "" {
			Options.DbPort, err = strconv.Atoi(dbPort)
			if err != nil {
				return fmt.Errorf("invalid port number: %s", dbPort)
			}
		}
	} else {
		Options.DbPort = 0
		dbName := WizardPrompt(fmt.Sprintf("Database file name (%s)", tibulaDb))
		if dbName == "" {
			Options.DbName = tibulaDb
		} else {
			Options.DbName = dbName
		}
		if _, err := os.Stat(Options.DbName); err == nil {
			return errors.New("An sqlite database with the same name already exists, remove it and try again.")
		}
	}
	//misc
	language := WizardPrompt("Choose default language (en)")
	if language != "" {
		Options.Language = language
	}
	logLevel := WizardPrompt("Choose log level between 0=None 1=Error, 2=Warn, 3=Info, 4=Debug (3)")
	if logLevel != "" {
		Options.LogLevel, err = strconv.Atoi(logLevel)
		if err != nil {
			return fmt.Errorf("invalid log level: %s", logLevel)
		}
	}
	Options.LogFile = WizardPrompt("Choose a log file (stderr)")
	jsonFile := WizardPrompt(fmt.Sprintf("Config file (%s)", tibulaJson))
	if jsonFile == "" {
		jsonFile = tibulaJson
	}

	if err := Setup(); err != nil {
		return err
	}

	Options.ConfigFile = ""
	Options.DbSetupUser = ""
	Options.DbSetupPass = ""
	if err := ConfigWrite(jsonFile, &Options); err != nil {
		return fmt.Errorf("Cannot write the configuration file, %w\n", err)
	}
	Options.ConfigFile = jsonFile

	return nil
}
