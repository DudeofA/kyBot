package main

import (
	"math/rand"

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

		case "readycheck":
			AddReadyCheck(i)

		case "coinflip":
			results := "tails"
			if rand.Intn(2) == 0 {
				results = "heads"
			}
			resp := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: results,
				},
			}
			err := s.InteractionRespond(i.Interaction, resp)
			if err != nil {
				log.Errorf("Error responding to the interaction: %s", err.Error())
			}

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

		switch i.MessageComponentData().CustomID {

		case "refresh_server":
			var server Server
			result := db.Where(&Server{StatusMessageID: i.Message.ID}).Limit(1).Find(&server)
			if result.RowsAffected == 1 {
				server.Update()
			}

		case "ready":
			var readyCheck ReadyCheck
			result := db.Preload("ReadyStatuses").Where(&ReadyCheck{StatusMessageID: i.Message.ID}).Limit(1).Find(&readyCheck)
			if result.RowsAffected == 1 {
				for idx, user := range readyCheck.ReadyStatuses {
					if i.Member.User.ID == user.UserID {
						readyCheck.ReadyStatuses[idx].Status = "Ready"
						db.Save(&readyCheck.ReadyStatuses[idx])
						readyCheck.Update()
					}
				}
			}

		case "not_ready":
			var readyCheck ReadyCheck
			result := db.Preload("ReadyStatuses").Where(&ReadyCheck{StatusMessageID: i.Message.ID}).Limit(1).Find(&readyCheck)
			if result.RowsAffected == 1 {
				for idx, user := range readyCheck.ReadyStatuses {
					if i.Member.User.ID == user.UserID {
						readyCheck.ReadyStatuses[idx].Status = "Not Ready"
						db.Save(&readyCheck.ReadyStatuses[idx])
						readyCheck.Update()
					}
				}
			}

		case "delete_readycheck":
			var readyCheck ReadyCheck
			result := db.Preload("ReadyStatuses").Where(&ReadyCheck{StatusMessageID: i.Message.ID}).Limit(1).Find(&readyCheck)
			if result.RowsAffected == 1 {
				readyCheck.Remove()
			}

		default:
			log.Warnf("Unknown Message Component: %s", i.MessageComponentData().CustomID)
		}

	default:
		log.Warnf("Unknown Interaction type: %s", i.Type.String())
	}
}
