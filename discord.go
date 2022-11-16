package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	datetimeLayout   = "2006-01-02 15:04:05"
	contentSucceeded = "Congratulations! You have claimed your daily free chest in web store.\n" +
		"> Nickname: %s\n" +
		"> Item: %s\n" +
		"> Time: %s\n"
	contentFalied = "Someting went wrong when claiming daily free chest in web store.\n" +
		"> Nickname: %s\n" +
		"> Item: %s\n" +
		"> Message: %s\n" +
		"> Time: %s\n"
)

var dgSession *discordgo.Session

func init() {
	go openDiscord()
}

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "help",
			Description: "Help command",
		},
		{
			Name:        "claim",
			Description: "Command for claiming free chest in web store",
			Options: []*discordgo.ApplicationCommandOption{

				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "nickname",
					Description: "Your nickname",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "email",
					Description: "Your email",
					Required:    false,
				},
			},
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"help": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Hello there!",
				},
			})
		},

		"claim": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := i.ApplicationCommandData().Options

			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
			for _, opt := range options {
				optionMap[opt.Name] = opt
			}
			nickname := ""
			email := ""
			content := ""
			if option, ok := optionMap["nickname"]; ok {
				nickname = option.StringValue()
			}
			if option, ok := optionMap["email"]; ok {
				email = option.StringValue()
			}
			if nickname == "" && email == "" {
				content = "> Nickname or email is required"
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: content,
					},
				})
				return
			} else {
				content = "> Please wait..."
			}
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: content,
				},
			})

			for i, player := range config.Players {
				if player.Nickname == nickname || player.Email == email {
					config.Players[i].LastClaimedDate = time.Now().Format(dateLayout)
					if ok, item, msg := claim(player); ok {
						content = fmt.Sprintf(contentSucceeded, player.Nickname, item, time.Now().Format(datetimeLayout))
					} else {
						content = fmt.Sprintf(contentFalied, player.Nickname, item, msg, time.Now().Format(datetimeLayout))
					}
				}
			}
			_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &content,
			})
			checkErr("send message to discord error: ", err, Info)
		},
	}
)
var registeredCommands = make([]*discordgo.ApplicationCommand, len(commands))

func openDiscord() {
	var err error
	dgSession, err = discordgo.New("Bot " + config.Discord.Token)
	checkErr("invalid discord bot parameters: ", err, Error)

	dgSession.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
	dgSession.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		//Log in
	})
	err = dgSession.Open()
	checkErr("open discord seesion error: ", err, Error)

	for i, v := range commands {
		cmd, err := dgSession.ApplicationCommandCreate(dgSession.State.User.ID, config.Discord.Guild, v)
		checkErr("create discord command error: ", err, Error)
		registeredCommands[i] = cmd
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("press Ctrl+C to exit")
	<-stop

	defer dgSession.Close()

	if config.Discord.RemoveCommands {
		log.Println("removing commands...")
		for _, v := range registeredCommands {
			err := dgSession.ApplicationCommandDelete(dgSession.State.User.ID, config.Discord.Guild, v.ID)
			checkErr("delete discord command error: ", err, Error)
		}
	}

	log.Println("gracefully shutting down.")
}

func sendDiscordMessage(content string) {
	if config.Discord.AutoMsgToChannel != "" {
		dgSession.ChannelMessageSend(config.Discord.AutoMsgToChannel, content)
	}
}
