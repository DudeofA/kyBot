package main

import (
	"github.com/bwmarrin/discordgo"
    "fmt"
    "os"
    "log"
    "time"
    "strings"
    "io/ioutil"
)

func SaveQuote(s *discordgo.Session, message *discordgo.MessageCreate, quote string) (formatQuote string){
	//s.ChannelMessageSend(message.ChannelID, quote)
    file, err := os.OpenFile("quotes.txt", os.O_WRONLY|os.O_APPEND, 0644)
    if err != nil {
        log.Fatalf("failed opening file: %s", err)
    }
    defer file.Close()


    formatQuote = fmt.Sprintf("```css\n[ %s ]\n%s\n```", time.Now().Format("Jan 2 3:04:05PM 2006"), quote)
    _, err = file.WriteString(formatQuote + "\n\n")
    if err != nil {
        log.Fatalf("failed writing to file: %s", err)
    }
    return
}

func ListQuote(s *discordgo.Session, m *discordgo.MessageCreate, i ...int) (entries int){

    quotes, err := ioutil.ReadFile("quotes.txt")
    if err != nil{
        panic(err)
    }

    lines := strings.Split(string(quotes), "\n\n")

    entries = 0
    for _, line := range lines {
        s.ChannelMessageSend(m.ChannelID, line)
        entries++
    }

    return
}
