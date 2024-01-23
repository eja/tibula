// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package main

import (
	"fmt"
	"github.com/eja/tibula/sys"
	"golang.org/x/term"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

func exit() {
	fmt.Println("Press enter to continue.")
	fmt.Scanln()
	os.Exit(0)
}

func prompt(message string) string {
	fmt.Printf("%s: ", message)
	var input string
	fmt.Scanln(&input)
	return input
}

func password(message string) string {
	fmt.Printf("%s: ", message)
	pass, _ := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	return strings.TrimSpace(string(pass))
}

func openBrowser(url string) {
	switch os := osType(); os {
	case "darwin":
		exec.Command("open", url).Start()
	case "linux":
		exec.Command("xdg-open", url).Start()
	case "windows":
		exec.Command("cmd", "/c", "start", url).Start()
	}
}

func osType() string {
	return strings.ToLower(runtime.GOOS)
}

func runCommand(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running command %s: %v\n", command, err)
		return err
	}
	return nil
}

func main() {
	var tibulaCommand = "./tibula"
	if osType() == "windows" {
		tibulaCommand = ".\tibula.exe"
	}

	sys.Configure()
	if _, err := os.Stat("tibula.json"); err == nil {
		start := prompt("A configuration file already exists. Do you want to start Tibula instead? (Y/n)")
		if strings.ToLower(start) != "n" {
			if err := sys.ConfigRead("tibula.json"); err != nil {
				fmt.Println(err)
				exit()
			}
			if sys.Options.WebTlsPrivate != "" && sys.Options.WebTlsPublic != "" {
				openBrowser(fmt.Sprintf("https://%s:%d", sys.Options.WebHost, sys.Options.WebPort))
			} else {
				openBrowser(fmt.Sprintf("http://%s:%d", sys.Options.WebHost, sys.Options.WebPort))
			}
			runCommand(tibulaCommand, "--config", "tibula.json", "--start")
			exit()
		}
	}
	//setup
	setupUser := prompt("Choose an administrator username (admin)")
	if setupUser != "" {
		sys.Options.SetupUser = setupUser
	}
	sys.Options.SetupPass = password("Enter the administrator password")
	if sys.Options.SetupPass == "" {
		fmt.Println("Password cannot be empty.")
		exit()
	}
	passConfirm := password("Repeat password")
	if passConfirm != sys.Options.SetupPass {
		fmt.Println("Passwords do not match.")
		exit()
	}
	//web
	webHost := prompt("Web address to listen to (localhost)")
	if webHost != "" {
		sys.Options.WebHost = webHost
	}
	webPort := prompt("Web port to listen to (35248)")
	if webPort != "" {
		sys.Options.WebPort, _ = strconv.Atoi(webPort)
	}
	webPrivate := prompt("Web https private certificate path (none)")
	if webPrivate != "" {
		sys.Options.WebTlsPrivate = webPrivate
	}
	webPublic := prompt("Web https public certificate path (none)")
	if webPublic != "" {
		sys.Options.WebTlsPublic = webPublic
	}
	//db
	dbType := prompt("Choose a database engine between sqlite and mysql (sqlite)")
	if dbType == "mysql" {
		sys.Options.DbType = dbType
		sys.Options.DbName = prompt("Database name")
		sys.Options.DbUser = prompt("Database username")
		sys.Options.DbPass = password("Database password")
		dbHost := prompt("Database hostname (localhost)")
		if dbHost != "" {
			sys.Options.DbHost = dbHost
		}
		dbPort := prompt("Database port (3306)")
		if dbPort != "" {
			sys.Options.DbPort, _ = strconv.Atoi(dbPort)
		}
	} else {
		sys.Options.DbPort = 0
		dbName := prompt("Database file name (tibula.db)")
		if dbName != "" {
			sys.Options.DbName = dbName
		}
	}
	//misc
	language := prompt("Choose default language (en)")
	if language != "" {
		sys.Options.Language = language
	}
	logLevel := prompt("Choose log level between 1=Error, 2=Warn, 3=Info, 4=Debug, 5=Trace (3)")
	if logLevel != "" {
		sys.Options.LogLevel, _ = strconv.Atoi(logLevel)
	}
	jsonFile := prompt("Config file (tibula.json)")
	if jsonFile == "" {
		jsonFile = "tibula.json"
	}

	//setup
	if err := sys.ConfigWrite(jsonFile); err != nil {
		fmt.Printf("Cannot write the configuration file, %v\n", err)
		exit()
	}
	if err := runCommand(tibulaCommand, "--config", jsonFile, "--setup"); err != nil {
		exit()
	}
	sys.Options.SetupUser = ""
	sys.Options.SetupPass = ""
	if err := sys.ConfigWrite(jsonFile); err != nil {
		fmt.Printf("Cannot write the configuration file, %v\n", err)
		exit()
	}

	//start
	if sys.Options.WebTlsPrivate != "" && sys.Options.WebTlsPublic != "" {
		openBrowser(fmt.Sprintf("https://%s:%d", sys.Options.WebHost, sys.Options.WebPort))
	} else {
		openBrowser(fmt.Sprintf("http://%s:%d", sys.Options.WebHost, sys.Options.WebPort))
	}
	if err := runCommand(tibulaCommand, "--config", jsonFile, "--start"); err != nil {
		exit()
	}
}
