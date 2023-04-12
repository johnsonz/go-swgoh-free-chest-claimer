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
	datetimeLayout     = "2006-01-02 15:04:05"
	datetimeZoneLayout = "2006-01-02 15:04:05 -0700"
	contentSucceed     = "Congratulations! You have claimed your daily free chest in web store.\n" +
		"> Nickname: %s\n" +
		"> Item: %s\n" +
		"> Time: %s\n"
	contentFalied = "Someting went wrong when claiming daily free chest in web store.\n" +
		"> Nickname: %s\n" +
		"> Item: %s\n" +
		"> Message: %s\n"
)

var NumMapping = []string{"I", "II", "III", "IV", "V", "VI", "VII", "VIII", "IX", "X"}
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
			var nickname, email string
			content := "> Please wait..."

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: content,
				},
			})

			var embeds []*discordgo.MessageEmbed
			if option, ok := optionMap["nickname"]; ok {
				nickname = option.StringValue()
			}
			if option, ok := optionMap["email"]; ok {
				email = option.StringValue()
			}

			for index, player := range config.Players {
				if player.Nickname == nickname || player.Email == email || player.DiscordId == i.Member.User.ID {
					config.Players[index].LastClaimedDate = time.Now().Format(dateLayout)
					items := claim(player)
					embeds = append(embeds, generateMessageEmbed(items, player.Nickname))
					content = ""
					if len(embeds) == 0 {
						content = "> There is no any free chest now."
					}
					_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
						Content: &content,
						Embeds:  &embeds,
					})
					checkErr("send message to discord error: ", err, Info)
					return
				}
			}
			content = "> Nickname or email is incorrect"
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &content,
			})
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

func sendDiscordMessageEmbed(embed *discordgo.MessageEmbed) {
	if config.Discord.AutoMsgToChannel != "" {
		dgSession.ChannelMessageSendEmbed(config.Discord.AutoMsgToChannel, embed)
	}
}

func generateMessage(items []ItemClaim, playerName string) string {
	msg := ""
	for _, item := range items {
		if item.IsSucceed {
			msg += fmt.Sprintf(contentSucceed, playerName, item.Name, time.Now().Format(datetimeLayout))
		} else {
			msg += fmt.Sprintf(contentFalied, playerName, item.Name, item.Message)
			if item.StartTime == 0 {
				msg += "> Time: " + time.Now().Format(datetimeLayout) + "\n"
			} else {
				msg += "> Start Time: " + time.Unix(item.StartTime, 0).Format(datetimeZoneLayout) + "\n"
			}
		}
	}
	return msg
}
func generateMessageEmbed(items []ItemClaim, playerName string) *discordgo.MessageEmbed {
	var fields []*discordgo.MessageEmbedField
	for i, item := range items {
		name := NumMapping[i]
		if item.IsSucceed {
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:  name,
				Value: fmt.Sprintf("```Item: %s\nMessage: %s```", item.Name, "Successful"),
			})
		} else {
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:  name,
				Value: fmt.Sprintf("```Item: %s\nMessage: %s\nStart: %s```", item.Name, item.Message, time.Unix(item.StartTime, 0).Format(dateLayout)),
			})
		}
	}
	return &discordgo.MessageEmbed{
		Color:     0xff0000,
		Title:     fmt.Sprintf("Daily free chests for [%s]", playerName),
		Fields:    fields,
		Timestamp: time.Now().Format(time.RFC3339),
	}
}
