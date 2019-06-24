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
		logChannel, err := s.State.Channel(botConfig.LogID)
		if err != nil {
			panic(err)
		}
		//Check if bot has permission to Send Messages and is accessable
		perm, err := MemberHasPermission(s, logChannel.GuildID, self.ID, 2048)
		if err != nil {
			panic(err)
		}

		//If log channel is valid and bot has permission
		if perm {
			//Format Name
			name := m.Author.Username + "#" + m.Author.Discriminator
			channel, err := s.Channel(m.ChannelID)
			if err != nil {
				panic(err)
			}
			//If message is a direct message (DM), there will be no guild ID or channel name
			var guildName string
			if m.GuildID == "" {
				guildName = "Direct"
				channel.Name = "Message"
			} else {
				guild, err := s.Guild(m.GuildID)
				if err != nil {
					panic(err)
				}
				guildName = guild.Name
			}

			//Generate remaining formatting
			message := m.ContentWithMentionsReplaced()
			cleanMsg := strings.Replace(message, "`", "\\`", -1)
			cleanMsg = strings.Replace(cleanMsg, "\n", "\n# ", -1)
			fmtMsg := fmt.Sprintf("```ini\n[ %s ] - Message - %s - %s - %s - %s\n# %s\n```",
				time.Now().Format("Jan 2 3:04:05PM 2006"),
				guildName,
				channel.Name,
				name,
				m.ID,
				cleanMsg)

			//Send log
			s.ChannelMessageSend(botConfig.LogID, fmtMsg)
		} else {
			fmt.Printf("No permission/unaccessable log channel: %s", m.ID)
		}
	}
}
