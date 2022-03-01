package main

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

var (
	TOKEN          string
	APPID          string
	DEBUG          bool
	DEBUG_GUILD_ID string
	VERSION        string
)

func init() {
	godotenv.Load()

	var found bool
	TOKEN, found = os.LookupEnv("DISCORD_TOKEN")
	if !found {
		log.Fatal("No token found, please set env DISCORD_TOKEN to a valid Discord bot token")
	}
	APPID, found = os.LookupEnv("APP_ID")
	if !found {
		log.Fatal("No app id found, please set env APP_ID to a valid Discord app id")
	}

	DEBUG = false
	_, found = os.LookupEnv("DEBUG")
	if found {
		DEBUG = true
		DEBUG_GUILD_ID, found = os.LookupEnv("DEBUG_GUILD_ID")
		if !found {
			log.Fatal("Debug mode, but DEBUG_GUILD_ID not set")
		}
	}

	VERSION = GetBotVersion()
}

func GetBotVersion() (ver string) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Unable to get current executable for working directory: %s", err.Error())
	}

	// Open the file and grab it line by line into textlines
	changelog, err := os.Open(filepath.Join(cwd, "CHANGELOG.md"))
	if err != nil {
		log.Fatalf("Unable to access changelog file: %s", err.Error())
	}
	defer changelog.Close()

	scanner := bufio.NewScanner(changelog)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		// Find version number between 2 square brackets
		re := regexp.MustCompile(`\[([^\[\]]*)\]`)
		if re.MatchString(scanner.Text()) {
			versionSplice := re.FindAllString(scanner.Text(), 1)
			ver = versionSplice[0]
			ver = strings.Trim(ver, "[")
			ver = strings.Trim(ver, "]")
			return ver
		}

	}
	log.Fatalf("Could not find current version in changelog: %s", err.Error())
	return ""
}
