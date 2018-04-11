package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
)

func SetAnthem(s *discordgo.Session, m *discordgo.MessageCreate) {
	usr, i := ReadUser(s, m, "MSG")
	var anthem string
	anthem = usr.Anthem
	if usr.Anthem == "" {
		anthem = "nothing"
	}
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(
		"Your current anthem is **%s**\nTo change anthem, please enter the number corresponding to the one you want",
		anthem))
	files, err := ioutil.ReadDir("clips/anthems")
	if err != nil {
		panic(err)
	}
	fileList := "[0] Nothing\n"
	for j, file := range files {
		fileList += "[" + strconv.Itoa(j+1) + "] " + file.Name() + "\n"
	}
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("```\nAnthem options:\n%s\n```", fileList))
	var stopHandling func()
	stopHandling = s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == usr.UserID {
			index, err := strconv.Atoi(m.Content)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Please enter a number in the list, I don't understand what you said")
				return
			}
			if index == 0 {
				USArray.Users[i].Anthem = ""
				s.ChannelMessageSend(m.ChannelID, "Anthem cleared")
				WriteUserFile()
				stopHandling()
				return
			}
			// _, err := ioutil.ReadFile("clips/anthems/" + files[index-1].Name())
			if index <= len(files) && index >= 1 {
				USArray.Users[i].Anthem = files[index-1].Name()
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Anthem set to %s", files[index-1].Name()))
				WriteUserFile()
				stopHandling()
				return
			} else {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Anthem '%s' not found. Exiting menu...",
					files[index-1].Name()))
				s.ChannelMessageSend(m.ChannelID, err.Error())
				stopHandling()
				return
			}
		}
	})
}

func PlayClip(s *discordgo.Session, m *discordgo.MessageCreate, clip string) {
	var stopChan = make(chan bool)

	if !config.Noise {
		s.ChannelMessageSend(m.ChannelID, "Noise commands are disabled")
		return
	}

	_, i := ReadUser(s, m, "MSG")
	if USArray.Users[i].NoiseCredits >= 100 {
		USArray.Users[i].NoiseCredits -= 100
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("💵 | You now have a total of **%d** %s coins", USArray.Users[i].NoiseCredits, config.Coins))
		WriteUserFile()
	} else {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(
			"You do not have enough credits to use this command\nYou need 100 coins (have %d)", USArray.Users[i].NoiseCredits))
		return
	}

	c, _ := s.State.Channel(m.ChannelID)
	g, _ := s.State.Guild(c.GuildID)
	for _, vs := range g.VoiceStates {
		if vs.UserID == m.Author.ID {
			curChan, _ := s.ChannelVoiceJoin(c.GuildID, vs.ChannelID, false, false)
			switch clip {

			case "yee":
				dgvoice.PlayAudioFile(curChan, "clips/yee.mp3", stopChan)
				// voiceChannel.Disconnect()
				return
			case "bitconnect":
				dgvoice.PlayAudioFile(curChan, "clips/bitconnect.wav", stopChan)
				// voiceChannel.Disconnect()
				return
			default:
				s.ChannelMessageSend(m.ChannelID, "File not found, unable to play clip...")
			}
		}
	}

	s.ChannelMessageSend(m.ChannelID, "I can't sing for you if you aren't in a voice channel :c")
	return

}

func PlayAnthem(s *discordgo.Session, v *discordgo.VoiceStateUpdate, anthem string) {
	var stopChan = make(chan bool)

	c, _ := s.State.Channel(v.ChannelID)
	voiceChan, err := s.ChannelVoiceJoin(c.GuildID, v.ChannelID, false, false)
	if err != nil {
		panic(err)
	}
	dgvoice.PlayAudioFile(voiceChan, fmt.Sprintf("clips/anthems/%s", anthem), stopChan)
	// voiceChan.Disconnect()

	_, i := ReadUser(s, v, "VOICE")
	USArray.Users[i].PlayAnthem = false
	WriteUserFile()

	go func() {
		time.Sleep(time.Minute * 10)
		USArray.Users[i].PlayAnthem = true
		WriteUserFile()
	}()
}
