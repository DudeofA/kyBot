/* 	meme.go
_________________________________
Code for meme generation of Kylixor Discord Bot
Andrew Langhill
kylixor.com
*/

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type MemeResp struct {
	Success bool
	Message string
}

// GenerateMeme - Take in argument from command, returns either link or relevant error
func GenerateMeme(data string) (msg string) {
	memeHost := k.botConfig.KyAPI
	if memeHost == "" {
		return "kyAPI host not defined in conf.json, define it to use this function"
	}
	reqURI := ""
	input := strings.Split(data, "|")
	for i := range input {
		input[i] = strings.TrimSpace(input[i])
	}
	switch len(input) {
	case 0:
		// Gen random meme
		reqURI = fmt.Sprintf("http://%s/meme/gen", memeHost)
	case 2:
		// Gen random meme with captions
		toptext := url.PathEscape(input[0])
		bottext := url.PathEscape(input[1])
		reqURI = fmt.Sprintf("http://%s/meme/gen?toptext=%s&bottext=%s", memeHost, toptext, bottext)
	case 3:
		// Gen meme with captions
		meme := url.PathEscape(input[0])
		toptext := url.PathEscape(input[1])
		bottext := url.PathEscape(input[2])
		reqURI = fmt.Sprintf("http://%s/meme/gen?meme=%s&toptext=%s&bottext=%s", memeHost, meme, toptext, bottext)
	default:
		// Display help
		commandSyntax := "Command Syntax:\n```\n!meme <meme name> | <top captions> | <bot caption>\nOR !meme <top caption> | <botcaption> (random meme will be chosen)\nYou can also use a space to have ```"
		return commandSyntax
	}

	resp, err := http.Get(reqURI)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var m MemeResp
	err = json.NewDecoder(resp.Body).Decode(&m)
	if err != nil {
		k.Log("ERR", "Error decoding response from kyAPI")
		k.Log("ERR", err.Error())
		return "Error creating meme"
	}

	link := m.Message
	return link
}
