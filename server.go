package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/iverly/go-mcping/mcping"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"layeh.com/gumble/gumble"
)

const (
	REFRESH_EMOJI = "🔄"
	DELETE_EMOJI  = "🗑"
)

var (
	MUMBLE_DEFAULT_PORT = uint16(64738)
	MUMBLE_ICON_URL     = "https://cdn.icon-icons.com/icons2/1381/PNG/512/mumble_94248.png"

	MINECRAFT_DEFAULT_PORT = uint16(25565)
	MINECRAFT_ICON_URL     = "https://cdn.icon-icons.com/icons2/2699/PNG/512/minecraft_logo_icon_168974.png"
)

type Server struct {
	Host            string `gorm:"primaryKey"` // IP address of server
	Port            uint16 // Port of server
	Type            string // Server Type
	CurrentUsers    int    // Current number of players Status
	MaxUsers        int    // Maximum allower players
	UserList        string // New line delimited list of players
	MOTD            string // Message of the day
	Version         string // Minecraft version
	Status          bool   // True if server is pingable
	Ping            int64  // Ping in milliseconds
	StatusChannelID string // Channel ID of the server status message
	StatusMessageID string // Message ID of the server status message
}

func init() {
	addServerCommand := &discordgo.ApplicationCommand{
		Name:        "add-server-status",
		Type:        discordgo.ChatApplicationCommand,
		Description: "Add a status message of a server to this channel",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "type",
				Description: "Pick what kind of server to be monitored in this channel",
				Required:    true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{
						Name:  "Minecraft",
						Value: "minecraft",
					},
					{
						Name:  "Mumble",
						Value: "mumble",
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "host",
				Description: "IP Address or hostname of server",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "port",
				Description: "Optional port of server",
				Required:    false,
			},
		},
	}
	AddCommand(addServerCommand)
}

func AddServer(i *discordgo.InteractionCreate, serverType string, host string, portString string) {
	var port uint16
	if portString == "" {
		switch serverType {
		case "minecraft":
			port = MINECRAFT_DEFAULT_PORT
		case "mumble":
			port = MUMBLE_DEFAULT_PORT
		}
	} else {
		port64, err := strconv.ParseUint(portString, 10, 16)
		if err != nil {
			log.Errorf("Error converting port to integer: %s", portString)
			return
		}
		port = uint16(port64)
	}

	server := Server{
		Type:            serverType,
		Host:            strings.ToLower(host),
		Port:            port,
		StatusChannelID: i.ChannelID,
	}

	// If server not found in database, add it
	if err := db.Take(&server).Error; err != nil {
		db.Create(&server)
		server.ping()
		db.Model(&server).Where(&Server{Host: server.Host}).Updates(&server)
		msg := server.buildEmbedMsg()

		resp := &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					msg.Embed,
				},
				Content:    "",
				Components: msg.Components,
			},
		}
		err := s.InteractionRespond(i.Interaction, resp)
		if err != nil {
			log.Errorf("Error responding to the interaction: %s", err.Error())
		}
		interaction, err := s.InteractionResponse(i.Interaction)
		if err != nil {
			log.Errorf("Error getting interaction response message info: %s", err.Error())
		}
		server.StatusMessageID = interaction.ID
		db.Model(&server).Where(&Server{Host: server.Host}).Updates(&Server{StatusMessageID: server.StatusMessageID})
	}
}

func (server *Server) Update() {
	server.ping()
	db.Model(&server).Where(&Server{Host: server.Host}).Updates(&server)
	msg := server.buildEmbedMsg()
	server.updateStatusMessage(msg)
}

func (server *Server) Remove() {
	err := s.ChannelMessageDelete(server.StatusChannelID, server.StatusMessageID)
	if err != nil {
		log.Errorf("Error removing server status window: %s", err.Error())
	}
	db.Delete(server)
}

func (server *Server) updateStatusMessage(updateContent *discordgo.MessageSend) {

	var statusMsg *discordgo.Message
	var err error

	_, err = s.ChannelMessage(server.StatusChannelID, server.StatusMessageID)
	if err != nil {
		// Send status
		statusMsg, err = s.ChannelMessageSendComplex(server.StatusChannelID, updateContent)
		if err != nil {
			log.Errorln("Could not send server status message", err.Error())
			return
		}
	} else {
		// Edit existing message
		edit := &discordgo.MessageEdit{
			Components: &updateContent.Components,
			ID:         server.StatusMessageID,
			Channel:    server.StatusChannelID,
			Embed:      updateContent.Embed,
		}
		statusMsg, err = s.ChannelMessageEditComplex(edit)
		if err != nil {
			log.Errorln("Could not edit server message", err.Error())
			return
		}
	}

	server.StatusMessageID = statusMsg.ID
	db.Model(&server).Where(&Server{Host: server.Host}).Updates(&Server{StatusMessageID: server.StatusMessageID})
}

func (server *Server) ping() {
	switch server.Type {
	case "minecraft":
		ping := mcping.NewPinger()
		resp, err := ping.PingWithTimeout(server.Host, server.Port, 2*time.Second)
		if err != nil {
			server.Status = false
			server.Version = "N/A"
			server.UserList = ""
			server.CurrentUsers = 0
		} else {
			server.Status = true
			server.Version = resp.Version
			server.CurrentUsers = resp.PlayerCount.Online
			server.MaxUsers = resp.PlayerCount.Max
			server.MOTD = resp.Motd
			if len(resp.Sample) > 0 {
				playerString := "```"
				for _, player := range resp.Sample {
					playerString += player.Name + "\n"
				}
				playerString += "```"
				server.UserList = playerString
			} else {
				server.UserList = ""
			}
		}
	case "mumble":
		port := fmt.Sprintf("%d", server.Port)
		resp, err := gumble.Ping(net.JoinHostPort(server.Host, port), 1*time.Second, 2*time.Second)
		if err != nil {
			server.Status = false
			server.Version = "N/A"
			server.Ping = 0
			server.CurrentUsers = 0
		} else {
			major, minor, patch := resp.Version.SemanticVersion()
			server.Status = true
			server.Version = fmt.Sprintf("%d.%d.%d", major, minor, patch)
			server.Ping = resp.Ping.Milliseconds()
			server.CurrentUsers = resp.ConnectedUsers
			server.MaxUsers = resp.MaximumUsers
		}
	}
}

func (server *Server) buildEmbedMsg() (msg *discordgo.MessageSend) {
	var url string
	switch server.Type {
	case "minecraft":
		url = MINECRAFT_ICON_URL
	case "mumble":
		url = MUMBLE_ICON_URL
	}

	ip := server.Host
	if server.Port != MINECRAFT_DEFAULT_PORT && server.Port != MUMBLE_DEFAULT_PORT {
		ip = fmt.Sprintf("%s:%d", server.Host, server.Port)
	}
	ipEmbed := &discordgo.MessageEmbedField{
		Name:  "Server IP",
		Value: ip,
	}
	versionEmbed := &discordgo.MessageEmbedField{
		Name:  "Version",
		Value: server.Version,
	}
	status := "Offline"
	color := 0xff0000
	if server.Status {
		status = "Online"
		color = 0x00ff00
	}
	statusEmbed := &discordgo.MessageEmbedField{
		Name:  "Status",
		Value: status,
	}
	playersStr := fmt.Sprintf("%d/%d Online\n%s", server.CurrentUsers, server.MaxUsers, server.UserList)
	playersEmbed := &discordgo.MessageEmbedField{
		Name:  "Players",
		Value: playersStr,
	}
	caser := cases.Title(language.AmericanEnglish)
	title := fmt.Sprintf("%s Server Status", caser.String(server.Type))
	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: server.MOTD,
		Color:       color,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: url,
		},
		Fields: []*discordgo.MessageEmbedField{
			ipEmbed,
			versionEmbed,
			statusEmbed,
			playersEmbed,
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Last updated: %s", time.Now().Local().Format("Jan 2 2006 - 3:04 PM MST")),
		},
	}

	refresh_button := &discordgo.Button{
		Label: "Refresh",
		Style: 1,
		Emoji: &discordgo.ComponentEmoji{
			Name:     REFRESH_EMOJI,
			ID:       "",
			Animated: false,
		},
		CustomID: "refresh_server",
	}

	delete_button := &discordgo.Button{
		Label:    "Delete Server",
		Style:    4,
		Disabled: false,
		Emoji: &discordgo.ComponentEmoji{
			Name:     DELETE_EMOJI,
			ID:       "",
			Animated: false,
		},
		URL:      "",
		CustomID: "delete_server",
	}

	msg = &discordgo.MessageSend{
		Embed:   embed,
		Content: "",
		Components: []discordgo.MessageComponent{
			&discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					refresh_button,
					delete_button,
				},
			},
		},
	}

	return msg
}
