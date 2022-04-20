package main

import (
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

func InteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var err error
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
		err = s.InteractionRespond(i.Interaction, resp)
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

		case "toggle_wordle_reminder":
			var user User
			db.Where(&User{ID: i.Member.User.ID}).Limit(1).Find(&user)
			err = user.ToggleWordleReminder()
			if err != nil {
				log.Errorf("error enabling wordle reminders for user [%s]: %s", i.User.ID, err)
				return
			}
			wordle, err := GetWordle(i.Message.ChannelID)
			if err != nil {
				log.Error(err)
			}
			wordle.UpdateStatus()

		default:
			log.Warnf("Unknown Message Component: %s", i.MessageComponentData().CustomID)
		}

	default:
		log.Warnf("Unknown Interaction type: %s", i.Type.String())
	}
}
