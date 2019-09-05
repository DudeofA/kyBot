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

//LogValid - tests if the log is valid
func LogValid(s *discordgo.Session) bool {
	var perm bool
	//Check to see if Log channel is defined by first checking the state, then DiscordAPI
	if botConfig.LogID != "" {
		logChannel, err := s.State.Channel(botConfig.LogID)
		if err != nil {
			logChannel, err = s.Channel(botConfig.LogID)
			if err != nil {
				fmt.Printf("Log channel does not exist or is not defined")
				return false
			}
		}
		//Check if bot has permission to Send Messages and is accessable
		perm, err = MemberHasPermission(s, logChannel.GuildID, self.ID, 2048)
		if err != nil {
			fmt.Printf("Error checking permissions")
			return false
		}
	}
	return perm
}

//PrintLog - Send/Format the log message
func PrintLog(s *discordgo.Session, logType string, logTime time.Time, guild string, channel string, user string, msgID string, data string) {
	fmtTime := logTime.Format("Jan 2 3:04:05PM 2006")
	cleanMsg := strings.Replace(data, "`", "\\`", -1)
	cleanMsg = strings.Replace(cleanMsg, "\n", "\n# ", -1)
	fmtMsg := fmt.Sprintf("```ini\n[ %s ] - %s - %s - %s - %s - %s\n# %s\n```",
		fmtTime,
		logType,
		guild,
		channel,
		user,
		msgID,
		cleanMsg)

	//Send log
	s.ChannelMessageSend(botConfig.LogID, fmtMsg)
}

//LogMsg - log messages in the log channel
func LogMsg(s *discordgo.Session, m *discordgo.MessageCreate) {
	if LogValid(s) {

		//Get channel
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

		//Format name
		name := m.Author.Username + "#" + m.Author.Discriminator

		//Append attachment URLs to the end of the messages
		fullMsg := m.ContentWithMentionsReplaced()
		if len(m.Attachments) > 0 {
			for _, v := range m.Attachments {
				fullMsg += "\nProxyURL: " + v.ProxyURL + "\nURL: " + v.URL
			}
		}

		//Generate remaining formatting
		PrintLog(s,
			"Message",
			time.Now(),
			guildName,
			channel.Name,
			name,
			m.ID,
			fullMsg)
	}

	//Update KDB
	kdb.GetUser(s, m.Author.ID)
}

//LogVoice - log voice events in the log channel
func LogVoice(s *discordgo.Session, v *discordgo.VoiceStateUpdate) {
	//If logs are able to be written to a channel
	if LogValid(s) {

		//Get KDB user data
		KDBuser := kdb.GetUser(s, v.UserID)

		//If user hasn't changed channels, do nothing (usually a mute or deafen)
		if KDBuser.CurrentCID == v.ChannelID {
			return
		}

		//Get guild name and ID
		guild, err := s.Guild(v.GuildID)
		if err != nil {
			panic(err)
		}

		//Get user name and ID
		user, err := s.User(v.UserID)
		if err != nil {
			panic(err)
		}

		//Get the event that occured - LeftChannel, JoinedChannel
		var event string
		var channelName string
		//If update is a leave voice channel event
		if v.ChannelID == "" {
			oldChannel, err := s.Channel(KDBuser.CurrentCID)
			if err != nil {
				oldChannel.Name = "N/A"
			}
			channelName = oldChannel.Name
			event = fmt.Sprintf("Left Channel: %s", channelName)
		} else {
			channel, err := s.Channel(v.ChannelID)
			if err != nil {
				panic(err)
			}
			channelName = channel.Name
			event = fmt.Sprintf("Joined Channel: %s", channelName)
		}

		//Generate remaining formatting
		PrintLog(s,
			"Voice",
			time.Now(),
			guild.Name,
			channelName,
			user.Username+"#"+user.Discriminator,
			"",
			event)
	} else {
		fmt.Printf("No permission/unaccessable log channel: \"%s\"", botConfig.LogID)
	}

	//Update KDB with voice channel info
	KDBuser.LastSeenCID = KDBuser.CurrentCID
	KDBuser.CurrentCID = v.ChannelID
	kdb.Write()
}

// LogTxt - log information/errors from functions
func LogTxt(s *discordgo.Session, msgType string, msg string) {
	if LogValid(s) {
		fmtTime := time.Now().Format("Jan 2 3:04:05PM 2006")
		cleanMsg := strings.Replace(msg, "`", "\\`", -1)
		cleanMsg = strings.Replace(cleanMsg, "\n", "\n# ", -1)
		fmtMsg := fmt.Sprintf("```ini\n[ %s ] - %s\n# %s\n```", fmtTime, msgType, cleanMsg)

		s.ChannelMessageSend(botConfig.LogID, fmtMsg)
	}
}
