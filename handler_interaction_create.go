package main

import (
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

			AddServer(i, serverType, host, port)

		case "add-wordle-channel":
			// AddWordleChannel(i)

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

		var server Server
		switch i.MessageComponentData().CustomID {

		case "refresh_server":
			result := db.Where(&Server{StatusMessageID: i.Message.ID}).Limit(1).Find(&server)
			if result.RowsAffected == 1 {
				server.Update()
			}

		case "enable_wordle_reminder":
			err = EnableWordleReminder(i)
			if err != nil {
				log.Errorf("error enabling wordle reminders for user [%s]: %s", i.User.ID, err)
				return
			}

		case "disable_wordle_reminder":
			err = DisableWordleReminder(i)
			if err != nil {
				log.Errorf("error disabling wordle reminders for user [%s]: %s", i.User.ID, err)
				return
			}

		default:
			log.Warnf("Unknown Message Component: %s", i.MessageComponentData().CustomID)
		}

	default:
		log.Warnf("Unknown Interaction type: %s", i.Type.String())
	}
}
