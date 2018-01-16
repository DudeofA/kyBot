package main

import (
    "github.com/bwmarrin/discordgo"
//    "fmt"
//    "time"
)

func Test(s *discordgo.Session, m *discordgo.MessageCreate){

    //    s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("```css\n%s\n```", time.Now().Format("Jan _2 3:04:05 2006")))

    Vote(s, m, "derp")
}
