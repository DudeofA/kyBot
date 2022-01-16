package handlers

import (
	"kyBot/kyDB"
	"kyBot/servers"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

func InteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {

	case discordgo.InteractionApplicationCommand:

		data := i.ApplicationCommandData()
		switch data.Name {
		case "add-server-status":
			var serverType, host, port string
			serverMap := make(map[string]string)
			for _, option := range data.Options {
				serverMap[option.Name] = option.StringValue()
			}
			serverType = serverMap["type"]
			host = serverMap["host"]
			port = serverMap["port"]

			servers.AddServer(s, i, serverType, host, port)

		default:
			log.Warnln("aw fuck idk what this is: ", data.Name)
		}

	case discordgo.InteractionMessageComponent:

		// Respond before updating the server as pinging it takes >= 3 seconds
		resp := &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredMessageUpdate,
		}
		err := s.InteractionRespond(i.Interaction, resp)
		if err != nil {
			log.Errorf("Error responding to the interaction: %s", err.Error())
		}

		switch i.MessageComponentData().CustomID {
		case "refresh_server":
			var server servers.Server
			result := kyDB.DB.Where(&servers.Server{StatusMessageID: i.Message.ID}).Limit(1).Find(&server)
			if result.RowsAffected == 1 {
				server.Update(s)
			}
		case "delete_server":
			var server servers.Server
			result := kyDB.DB.Where(&servers.Server{StatusMessageID: i.Message.ID}).Limit(1).Find(&server)
			if result.RowsAffected == 1 {
				server.Remove(s)
			}

		default:
			log.Warnf("Unknown Message Component: %s", i.MessageComponentData().CustomID)
		}

	default:
		log.Warnf("Unknown Interaction type: %s", i.Type.String())
	}
}
