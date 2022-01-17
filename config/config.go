package config

import (
	"os"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

var (
	TOKEN string
	APPID string
	DEBUG bool
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
	}
}
