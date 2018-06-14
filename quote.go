package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"time"
)

func SaveQuote(s *discordgo.Session, message *discordgo.MessageCreate, quote string) (formatQuote string) {
	//s.ChannelMessageSend(message.ChannelID, quote)
	file, err := os.OpenFile("quotes.txt", os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(fmt.Sprintf("failed opening file: %s", err))
	}
	defer file.Close()

	//Add timestamp to quote
	formatQuote = fmt.Sprintf("```css\n[ %s ]\n%s\n```", time.Now().Format("Jan 2 3:04:05PM 2006"), quote)
	_, err = file.WriteString(formatQuote + "\n\n")
	if err != nil {
		panic(fmt.Sprintf("failed writing to file: %s", err))
	}
	return
}

func ListQuote(s *discordgo.Session, m *discordgo.MessageCreate, i ...int) (entries int) {

	quotes, err := ioutil.ReadFile("quotes.txt")
	if err != nil {
		panic(err)
	}

	lines := strings.Split(string(quotes), "\n\n")

	entries = 0
    var quoteList string
	for _, line := range lines {
        quoteList += "\n" + line
		entries++
	}
    s.ChannelMessageSend(m.ChannelID, quoteList)

	return
}

func ShowRandQuote(s *discordgo.Session, m *discordgo.MessageCreate) {

	rand.Seed(time.Now().UTC().UnixNano())
	quotes, err := ioutil.ReadFile("quotes.txt")
	if err != nil {
		panic(err)
	}

	lines := strings.Split(string(quotes), "\n\n")
	randNum := rand.Intn(len(lines) - 1)
	s.ChannelMessageSend(m.ChannelID, lines[randNum])
}
