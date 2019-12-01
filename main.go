package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"google.golang.org/api/option"

	dialogflow "cloud.google.com/go/dialogflow/apiv2"
	dialogflowpb "google.golang.org/genproto/googleapis/cloud/dialogflow/v2"

	"github.com/bwmarrin/discordgo"
	"github.com/jasonlvhit/gocron"

	"github.com/xploitedd/leic-discordbot/handlers"
	"github.com/xploitedd/leic-discordbot/misc"
)

type configuration struct {
	DiscordToken  *string   `json:"discord_token"`
	DialogFlowID  *string   `json:"dialogflow_id"`
	CommandPrefix *string   `json:"command_prefix"`
	PlayingWith   []*string `json:"playing_with"`
}

// config saves the main configuration file
var config configuration
var discord *discordgo.Session

func main() {
	// load configuration file
	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		fmt.Println("error reading the configuration file:", err)
		return
	}

	// try to parse the json
	config = configuration{}
	err = json.Unmarshal(data, &config)
	if err != nil {
		fmt.Println("error while parsing the config file:", err)
		return
	}

	// check if the required configuration fields are available
	if config.DiscordToken == nil || config.CommandPrefix == nil || config.DialogFlowID == nil {
		fmt.Println("invalid configuration file!")
		return
	}

	// create a new discord session
	discord, err = discordgo.New("Bot " + *config.DiscordToken)
	if err != nil {
		fmt.Println("error creating discord session:", err)
		return
	}

	// register the handlers
	discord.AddHandler(messageCreate)
	registerCommands()

	// load other things
	err = misc.LoadQuotes()
	if err != nil {
		fmt.Println("error occurred while loading quotes!")
	}

	// connect to the websocket
	err = discord.Open()
	if err != nil {
		fmt.Println("error connecting to websocket:", err)
		return
	}

	// run cron jobs
	gocron.Start()
	gocron.Every(5).Minutes().Do(playingMessageTask)
	gocron.RunAll()

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	gocron.Remove(playingMessageTask)
	discord.Close()
}

func registerCommands() {
	handlers.RegisterCommand("ajuda", func(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
		helpcmd := ""
		for name, command := range handlers.Commands {
			if command.Description != nil {
				helpcmd += *config.CommandPrefix + name + " >> " + *command.Description + "\n"
			}
		}

		hostname, _ := os.Hostname()
		s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
			Color: 0xeb4034,
			Title: "LEIC - ISEL -> Ajuda",
			Fields: []*discordgo.MessageEmbedField{
				&discordgo.MessageEmbedField{
					Name:  "Comandos Disponíveis",
					Value: helpcmd,
				},
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text: fmt.Sprintf("Message processed by %s", hostname),
			},
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: s.State.User.AvatarURL("64"),
			},
		})
	}).SetDescription("Obtem informação sobre outros comandos")

	handlers.RegisterCommand("citar", func(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
		quote := misc.RandomQuote(m.GuildID)
		if quote == nil {
			s.ChannelMessageSend(m.ChannelID, "Nenhuma citação está disponível de momento!")
			return
		}

		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("> <%s> %s", *quote.Emote, *quote.Quote))
	}).SetDescription("Cita uma das grandes lendas da LEIC no ISEL")

	handlers.RegisterCommand("ban", func(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
		mentions := m.Mentions
		if len(mentions) == 0 {
			s.ChannelMessageSend(m.ChannelID, "Por favor identifica alguém para ir embora!")
			return
		}

		user := mentions[0]
		if user.ID == m.Author.ID {
			s.ChannelMessageSend(m.ChannelID, "Nós compreendemos que não gostes de ti mesmo, mas é preciso saíres daqui?")
			return
		}

		if user.ID == s.State.User.ID {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("E se te fosses banir a ti mesmo, oh murcão? %s", m.Author.Mention()))
			return
		}

		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Olá %s! Acabaste de ser banido: https://www.youtube.com/watch?v=FXPKJUE86d0", user.Mention()))
	}).SetDescription("Faz ban a alguém de quem não gostes!")

	handlers.RegisterCommand("falar", func(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
		ctx := context.Background()

		sessionClient, err := dialogflow.NewSessionsClient(ctx, option.WithCredentialsFile("dialogflow.json"))
		if err != nil {
			fmt.Println(err)
			s.ChannelMessageSend(m.ChannelID, "Ocorreu um erro ao iniciar o pedido!")
			return
		}

		defer sessionClient.Close()

		query := strings.Join(args, " ")
		sessionPath := fmt.Sprintf("projects/%s/agent/sessions/%s", *config.DialogFlowID, m.Author.ID)
		textInput := dialogflowpb.TextInput{Text: query, LanguageCode: "pt-PT"}
		queryTextInput := dialogflowpb.QueryInput_Text{Text: &textInput}
		queryInput := dialogflowpb.QueryInput{Input: &queryTextInput}
		request := dialogflowpb.DetectIntentRequest{Session: sessionPath, QueryInput: &queryInput}

		message, msgerr := s.ChannelMessageSend(m.ChannelID, "`A processar...`")
		if msgerr != nil {
			return
		}

		response, err := sessionClient.DetectIntent(ctx, &request)
		if err != nil {
			fmt.Println(err)
			s.ChannelMessageEdit(m.ChannelID, message.ID, "Ocorreu um erro ao comunicar com o bot!")
			return
		}

		s.ChannelMessageEdit(m.ChannelID, message.ID, response.GetQueryResult().GetFulfillmentText())
	}).SetDescription("Para quando te sentes sozinho e precisas de alguém para falar").SetMinArgs(1)

	handlers.RegisterCommand("lotaria", func(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
		message, err := s.ChannelMessageSend(m.ChannelID, "<:DealWithMatilde:634114750799937536>")
		if err != nil {
			return
		}

		time.Sleep(200 * time.Millisecond)
		message, _ = s.ChannelMessageEdit(m.ChannelID, message.ID, message.Content+" <:HappyHernani:505418232421941268>")
		time.Sleep(200 * time.Millisecond)
		message, _ = s.ChannelMessageEdit(m.ChannelID, message.ID, message.Content+" <:SmugCoutinho:634011216460644373> ")
		time.Sleep(200 * time.Millisecond)
		message, _ = s.ChannelMessageEdit(m.ChannelID, message.ID, message.Content+" <:SmileAC:642387975900102666>")
		time.Sleep(200 * time.Millisecond)
		message, _ = s.ChannelMessageEdit(m.ChannelID, message.ID, "Perdeste a Lotaria!")
	}).SetDescription("Para quando te sentes com sorte")
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID || m.Author.Bot {
		return
	}

	handlers.ParseCommand(*config.CommandPrefix, s, m)
}

func playingMessageTask() {
	playingmsg := getPlayingMessage()
	if playingmsg != nil {
		_, time := gocron.NextRun()
		fmt.Printf("Playing %s\n", *playingmsg)
		fmt.Printf("Next game in %s\n", time.String())
		discord.UpdateStatus(0, *playingmsg)
	}
}

func getPlayingMessage() *string {
	if config.PlayingWith != nil {
		len := len(config.PlayingWith)
		if len > 0 {
			return config.PlayingWith[rand.Intn(len)]
		}
	}

	return nil
}
