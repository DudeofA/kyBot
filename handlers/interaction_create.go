package handlers

import (
	"kyBot/component"
	"kyBot/kyDB"

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

			component.AddServer(s, i, serverType, host, port)

		case "add-wordle-channel":
			component.AddWordleChannel(s, i)

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

		var server component.Server
		switch i.MessageComponentData().CustomID {

		case "refresh_server":
			result := kyDB.DB.Where(&component.Server{StatusMessageID: i.Message.ID}).Limit(1).Find(&server)
			if result.RowsAffected == 1 {
				server.Update(s)
			}

		case "join_wordle":
			wordle, err := component.GetWordle(i.Message.ChannelID)
			if err == nil {
				changed := wordle.AddUser(i.Member.User)
				if changed {
					wordle.UpdateStatus(s)
				}
			}
		case "leave_wordle":
			wordle, err := component.GetWordle(i.Message.ChannelID)
			if err == nil {
				changed := wordle.RemoveUser(i.Member.User)
				if changed {
					wordle.UpdateStatus(s)
				}
			}

		default:
			log.Warnf("Unknown Message Component: %s", i.MessageComponentData().CustomID)
		}

	default:
		log.Warnf("Unknown Interaction type: %s", i.Type.String())
	}
}
