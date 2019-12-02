package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jasonlvhit/gocron"

	"github.com/xploitedd/leic-discordbot/handlers"
	ai "github.com/xploitedd/leic-discordbot/misc/ai"
	quotes "github.com/xploitedd/leic-discordbot/misc/quotes"
)

var discord *discordgo.Session

func main() {
	// the configuration is the first thing to be loaded
	LoadConfig()

	// create a new discord session
	discord, err = discordgo.New("Bot " + *Config.DiscordToken)
	if err != nil {
		fmt.Println("error creating discord session:", err)
		return
	}

	// register the handlers
	discord.AddHandler(messageCreate)
	registerCommands()

	// load other things
	err = quotes.LoadQuotes()
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
		quote := quotes.RandomQuote(m.GuildID)
		if quote == nil {
			s.ChannelMessageSend(m.ChannelID, "Nenhuma citação está disponível de momento!")
			return
		}

		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("> <%s> %s", *quote.Emote, *quote.Quote))
	}).SetDescription("Cita uma das grandes lendas da LEIC no ISEL").SetGuildOnly(true)

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
	}).SetDescription("Faz ban a alguém de quem não gostes!").SetGuildOnly(true)

	handlers.RegisterCommand("falar", func(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
		query := strings.Join(args, " ")
		ai.SendTextQuery(s, m, query)
	}).SetDescription("Para quando te sentes sozinho e precisas de alguém para falar").SetMinArgs(1)

	handlers.RegisterCommand("lotaria", func(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
		guild, _ := s.Guild(m.GuildID)
		emojislen := len(guild.Emojis)
		var msg *discordgo.Message
		var err error
		if emojislen < 4 {
			msg, err = s.ChannelMessageSend(m.ChannelID, "`A jogar na lotaria...`")
			if err != nil {
				return
			}

			time.Sleep(1200 * time.Millisecond)
		} else {
			fmt.Println(guild.Emojis[0].MessageFormat())
			msg, err = s.ChannelMessageSend(m.ChannelID, guild.Emojis[rand.Intn(emojislen)].MessageFormat())
			if err != nil {
				return
			}

			time.Sleep(300 * time.Millisecond)
			msg, _ = s.ChannelMessageEdit(m.ChannelID, msg.ID, msg.Content+guild.Emojis[rand.Intn(emojislen)].MessageFormat())
			time.Sleep(300 * time.Millisecond)
			msg, _ = s.ChannelMessageEdit(m.ChannelID, msg.ID, msg.Content+guild.Emojis[rand.Intn(emojislen)].MessageFormat())
			time.Sleep(300 * time.Millisecond)
			msg, _ = s.ChannelMessageEdit(m.ChannelID, msg.ID, msg.Content+guild.Emojis[rand.Intn(emojislen)].MessageFormat())
			time.Sleep(300 * time.Millisecond)
		}

		number := rand.Intn(101)
		if number > 97 {
			s.ChannelMessageEdit(m.ChannelID, msg.ID, "Ganhaste a Lotaria!")
		} else {
			s.ChannelMessageEdit(m.ChannelID, msg.ID, "Perdeste a Lotaria!")
		}
	}).SetDescription("Para quando te sentes com sorte").SetGuildOnly(true)
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
	if Config.PlayingWith != nil {
		len := len(Config.PlayingWith)
		if len > 0 {
			return Config.PlayingWith[rand.Intn(len)]
		}
	}

	return nil
}
