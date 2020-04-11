/*	log.go
_________________________________
Log/debug code for Kylixor Discord Bot
Andrew Langhill
kylixor.com
*/

package main

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Log - log information/errors from functions
func (k *K) Log(msgType string, msg string) {
	fmtTime := time.Now().Format("Jan 2 3:04:05PM 2006")
	fmtMsg := fmt.Sprintf("[%s] [%s]: %s\n", fmtTime, msgType, msg)

	_, err := k.logfile.WriteString(fmtMsg)
	if err != nil {
		panic(err)
	}
}

// LogMsg - log messages in the log channel
func LogMsg(s *discordgo.Session, m *discordgo.MessageCreate) {

	//Get channel
	channel, err := s.Channel(m.ChannelID)
	if err != nil {
		k.Log("ERR", "Cannot find channel in LogMsg")
		return
	}

	//If message is a direct message (DM), there will be no guild ID or channel name
	var guildName string
	if m.GuildID == "" {
		guildName = "Direct"
		channel.Name = "Message"
	} else {
		guild, err := s.Guild(m.GuildID)
		if err != nil {
			k.Log("ERR", "Cannot find guild in LogMsg")
			return
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
	k.Log("MSG", fmt.Sprintf("%s - %s - %s - %s: %s", guildName, channel.Name, name, m.ID, fullMsg))
}

// LogVoice - log voice events in the log channel
func LogVoice(s *discordgo.Session, v *discordgo.VoiceStateUpdate) {
	//Get KDB user data
	user := k.kdb.ReadUser(s, v.UserID)

	//If user hasn't changed channels, do nothing (usually a mute or deafen)
	if user.CurrentCID == v.ChannelID {
		return
	}

	//Get guild name and ID
	guild := k.kdb.ReadGuild(s, v.GuildID)

	//Get the event that occurred - LeftChannel, JoinedChannel
	var event string
	var channelName string
	//If update is a leave voice channel event
	if v.ChannelID == "" {
		oldChannel, err := s.Channel(user.CurrentCID)
		if err != nil {
			oldChannel.Name = "N/A"
		}
		channelName = oldChannel.Name
		event = fmt.Sprintf("Left Channel: %s", channelName)
	} else {
		channel, err := s.Channel(v.ChannelID)
		if err != nil {
			k.Log("ERR (LogVoice)", "Cannot find guild")
		}
		channelName = channel.Name
		event = fmt.Sprintf("Joined Channel: %s", channelName)
	}

	//Generate remaining formatting
	k.Log("VOICE", fmt.Sprintf("%s - %s - %s: %s", guild.Name, channelName, user.Name+"#"+user.Discriminator, event))

	//Update KDB with voice channel info
	user.LastSeenCID = user.CurrentCID
	user.CurrentCID = v.ChannelID
	user.Update()
}

// LogDB - log database manipulation
func LogDB(itemType, name, id, action string) {
	fmtLog := fmt.Sprintf("[%s] \"%s\" [%s] %s %s", itemType, name, id, action, k.botConfig.DBName)
	k.Log("KDB", fmtLog)
}
