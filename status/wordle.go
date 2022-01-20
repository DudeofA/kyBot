package status

import (
	"kyBot/commands"
	"kyBot/kyDB"
	"strings"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

const (
	STOP_EMOJI              = "🛑"
	WORDLE_EMOJI            = "🟩"
	WORDLE_URL              = "https://www.powerlanguage.co.uk/wordle/"
	WORDLE_GREEN            = 0x538d4e
	WORDLE_JOIN_EMOTE_NAME  = "aenezukojump"
	WORDLE_JOIN_EMOTE_ID    = "849514753042546719"
	WORDLE_LEAVE_EMOTE_NAME = "PES2_SadGeRain"
	WORDLE_LEAVE_EMOTE_ID   = "849698641869406261"
)

func init() {
	addServerCommand := &discordgo.ApplicationCommand{
		Name:        "add-wordle-reminder",
		Type:        discordgo.ChatApplicationCommand,
		Description: "Add a repeating wordle reminder message",
	}
	commands.AddCommand(addServerCommand)
}

func AddWordleReminder(s *discordgo.Session, i *discordgo.InteractionCreate) {
	resp := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "A reminder message will be posted every day in this channel",
		},
	}
	err := s.InteractionRespond(i.Interaction, resp)
	if err != nil {
		log.Errorf("Error responding to the interaction: %s", err.Error())
	}

	server := Server{
		Type:            "wordle",
		Host:            i.ChannelID,
		Port:            0,
		StatusChannelID: i.ChannelID,
	}

	if err := kyDB.DB.Take(&server).Error; err != nil {
		kyDB.DB.Create(&server)
	}

}

func SendWordleReminders(s *discordgo.Session) {
	var server_objects []Server
	_ = kyDB.DB.Where(&Server{Type: "wordle"}).Find(&server_objects)
	for _, server := range server_objects {
		msg := server.buildWordleEmbedMsg(s)
		update, err := s.ChannelMessageSendComplex(server.StatusChannelID, msg)
		if err != nil {
			log.Errorf("Error sending wordle update: %s", err.Error())
		}
		server.StatusMessageID = update.ID
		kyDB.DB.Model(&server).Where(&Server{Host: server.Host}).Updates(&Server{StatusMessageID: server.StatusMessageID})
	}
}

func (server *Server) AddUser(s *discordgo.Session, u *discordgo.User) {
	if server.UserList == "" {
		server.UserList += u.Mention()
	} else if !strings.Contains(server.UserList, u.Mention()) {
		server.UserList += "\n" + u.Mention()
	}
	kyDB.DB.Model(&server).Where(&Server{Host: server.Host}).Updates(&Server{UserList: server.UserList})
	msg := server.buildWordleEmbedMsg(s)
	server.updateStatusMessage(s, msg)
}

func (server *Server) RemoveUser(s *discordgo.Session, u *discordgo.User) {
	if server.UserList != "" && strings.Contains(server.UserList, u.Mention()) {
		server.UserList = strings.Replace(server.UserList, u.Mention(), "", 1)
		server.UserList = strings.Replace(server.UserList, "\n\n", "\n", 1)
	}
	kyDB.DB.Model(&server).Where(&Server{Host: server.Host}).Updates(&Server{UserList: server.UserList})
	msg := server.buildWordleEmbedMsg(s)
	server.updateStatusMessage(s, msg)
}

func (server *Server) buildWordleEmbedMsg(s *discordgo.Session) (msg *discordgo.MessageSend) {
	optInButton := &discordgo.Button{
		Label: "Join Game",
		Style: 1,
		Emoji: discordgo.ComponentEmoji{
			Name:     WORDLE_JOIN_EMOTE_NAME,
			ID:       WORDLE_JOIN_EMOTE_ID,
			Animated: true,
		},
		CustomID: "join_server",
	}
	optOutButton := &discordgo.Button{
		Label: "Leave Game",
		Style: 4,
		Emoji: discordgo.ComponentEmoji{
			Name:     WORDLE_LEAVE_EMOTE_NAME,
			ID:       WORDLE_LEAVE_EMOTE_ID,
			Animated: true,
		},
		CustomID: "leave_server",
	}

	var playerString string
	if server.UserList == "" {
		playerString = "None :("
	} else {
		playerString = server.UserList
	}
	wordleEmbed := &discordgo.MessageEmbed{
		URL:         WORDLE_URL,
		Title:       "CLICK HERE TO PLAY WORDLE",
		Description: "New wordle available now!",
		Timestamp:   "",
		Color:       WORDLE_GREEN,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Join In!",
				Value: "Click the button to join the game for tracking",
			},
			{
				Name:  "Players",
				Value: playerString,
			},
		},
	}
	msg = &discordgo.MessageSend{
		Embed: wordleEmbed,
		Components: []discordgo.MessageComponent{
			&discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					optInButton,
					optOutButton,
				},
			},
		},
	}
	return msg
}
