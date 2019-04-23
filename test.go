package main

import (
    "bytes"
	"fmt"
    "os/exec"

	"github.com/bwmarrin/discordgo"
)

func Test(s *discordgo.Session, m *discordgo.MessageCreate) {
	s.ChannelMessageSend(m.ChannelID, "Testing Starting...")

    s.ChannelMessageSend(m.ChannelID, "Running update command...")
    cmd := exec.Command("ssh", "andrew@hermes", "/home/andrew/test")
    var out bytes.Buffer
    cmd.Stdout = &out
    err := cmd.Run()
    if err != nil {
        s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Updated failed: err - %s", err))
    }

    s.ChannelMessageSend(m.ChannelID, "Output: " + out.String())

    s.ChannelMessageSend(m.ChannelID, "Testing Done")

}
