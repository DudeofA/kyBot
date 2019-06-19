/* 	log.go
_________________________________
Log/debug code for Kylixor Discord Bot
Andrew Langhill
kylixor.com
*/

package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

//LogMsg - log messages in the log channel
func LogMsg(s *discordgo.Session, m *discordgo.MessageCreate) {
	//Check to see if Log channel is defined
	if botConfig.LogID != "" && m.ChannelID != botConfig.LogID {
		//Check if bot has permission to Send Messages and is accessable
		perm, err := MemberHasPermission(s, m.GuildID, self.ID, 2048)
		if err != nil {
			panic(err)
		}

		if perm {
			//Post the Log message
			name := m.Author.Username + "#" + m.Author.Discriminator
			channel, _ := s.Channel(m.ChannelID)
			guild, _ := s.Guild(m.GuildID)
			message := m.ContentWithMentionsReplaced()
			cleanMsg := strings.Replace(message, "`", "\\`", -1)
			cleanMsg = strings.Replace(cleanMsg, "\n", "\n# ", -1)
			fmtMsg := fmt.Sprintf("```ini\n[ %s ] - Message - %s - %s - %s - %s\n# %s\n```",
				time.Now().Format("Jan 2 3:04:05PM 2006"),
				guild.Name,
				channel.Name,
				name,
				m.ID,
				cleanMsg)

			s.ChannelMessageSend(botConfig.LogID, fmtMsg)
		} else {
			s.ChannelMessageSend(m.ChannelID, "No permission/unaccessable log channel")
		}
	}
}
