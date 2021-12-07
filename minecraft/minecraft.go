package minecraft

import (
	"fmt"
	"kyBot/kyDB"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/iverly/go-mcping/mcping"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	REFRESH_EMOJI = "🔄"
)

type MinecraftServer struct {
	gorm.Model
	IP             string `gorm:"primaryKey"` // IP address of server
	Port           uint16 // Port of server
	CurrentPlayers int    // Current number of players Status
	MaxPlayers     int    // Maximum allower players
	PlayerList     string // New line delimited list of players
	MOTD           string // Message of the day
	Version        string // Minecraft version
	Status         bool   // True if server is pingable
	ChannelID      string // Channel ID of the server status message
	MessageID      string // Message ID of the server status message
}

func AddServer(s *discordgo.Session, chanID string, ip string, port uint16) {
	if port < 1 {
		port = 25565
	}

	server := MinecraftServer{
		IP:        strings.ToLower(ip),
		Port:      port,
		ChannelID: chanID,
	}
	// If server not found in database, add it
	if err := kyDB.DB.Take(&server).Error; err != nil {
		kyDB.DB.Create(&server)
	}

	server.UpdateServer(s)

}

// Loop through all servers and update their status
func UpdateAllServers(s *discordgo.Session) {
	var mc_servers []MinecraftServer
	_ = kyDB.DB.Find(&mc_servers)
	for _, server := range mc_servers {
		server.UpdateServer(s)
	}
}

func (server *MinecraftServer) UpdateServer(s *discordgo.Session) {
	ping := mcping.NewPinger()
	response, err := ping.Ping(server.IP, server.Port)
	if err != nil {
		log.Warningf("Error pinging server, setting offline...: %s:%d\n%s", server.IP, server.Port, err.Error())
		server.Status = false
		server.CurrentPlayers = 0
		server.Version = "N/A"
		server.PlayerList = ""
	} else {
		server.CurrentPlayers = response.PlayerCount.Online
		server.MaxPlayers = response.PlayerCount.Max
		server.MOTD = response.Motd
		server.Version = response.Version
		server.Status = true
		var playerString string
		for _, player := range response.Sample {
			playerString += "```" + player.Name + "```\n"
		}
		server.PlayerList = playerString
	}

	server.UpdateServerMsg(s)

	kyDB.DB.Model(&server).Updates(MinecraftServer{
		CurrentPlayers: server.CurrentPlayers,
		MaxPlayers:     server.MaxPlayers,
		MOTD:           server.MOTD,
		Version:        server.Version,
		Status:         server.Status,
		MessageID:      server.MessageID,
	})
}

func (server *MinecraftServer) UpdateServerMsg(s *discordgo.Session) {
	ipEmbed := discordgo.MessageEmbedField{
		Name:  "Server IP",
		Value: server.IP,
	}
	versionEmbed := discordgo.MessageEmbedField{
		Name:  "Version",
		Value: server.Version,
	}
	status := "Offline"
	color := 0xff0000
	if server.Status {
		status = "Online"
		color = 0x00ff00
	}
	statusEmbed := discordgo.MessageEmbedField{
		Name:  "Status",
		Value: status,
	}
	playersStr := fmt.Sprintf("%d/%d Online\n%s", server.CurrentPlayers, server.MaxPlayers, server.PlayerList)
	playersEmbed := discordgo.MessageEmbedField{
		Name:  "Players",
		Value: playersStr,
	}
	data := discordgo.MessageEmbed{
		Title:       "Minecraft Server Status",
		Description: server.MOTD,
		Color:       color,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://cdn.icon-icons.com/icons2/2699/PNG/512/minecraft_logo_icon_168974.png",
		},
		Fields: []*discordgo.MessageEmbedField{
			&ipEmbed,
			&versionEmbed,
			&statusEmbed,
			&playersEmbed,
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Last updated: %s", time.Now().Local().Format("Jan 2 2006 - 3:04 PM MST")),
		},
	}

	var msg *discordgo.Message
	var err error
	if server.MessageID == "" {
		// Send embed
		msg, err = s.ChannelMessageSendEmbed(server.ChannelID, &data)
		if err != nil {
			log.Errorln("Could not send message", err.Error())
			return
		}
		// Add Reaction for refreshing data
		err = s.MessageReactionAdd(server.ChannelID, msg.ID, REFRESH_EMOJI)
		if err != nil {
			log.Errorln("Could not add reaction to server update", err.Error())
		}
	} else {
		msg, err = s.ChannelMessageEditEmbed(server.ChannelID, server.MessageID, &data)
		if err != nil {
			log.Errorln("Could not edit message", err.Error())
			return
		}
		needsEmote := true
		for _, reaction := range msg.Reactions {
			if reaction.Emoji.Name == REFRESH_EMOJI {
				needsEmote = false
			}
		}
		if needsEmote {
			err = s.MessageReactionAdd(server.ChannelID, msg.ID, REFRESH_EMOJI)
			if err != nil {
				log.Errorln("Could not add reaction to server update", err.Error())
			}
		}
	}

	server.MessageID = msg.ID
}
