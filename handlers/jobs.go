package handlers

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/jasonlvhit/gocron"
	"github.com/xploitedd/leic-discordbot/misc"
)

var discord *discordgo.Session

// StartJobs starts all registered cron jobs
func StartJobs(d *discordgo.Session) {
	discord = d

	gocron.Start()
	gocron.Every(5).Minutes().Do(runPlayingMessageJob)
	gocron.Every(1).Days().At("00:00").Do(runHolidaysJob)
	gocron.RunAll()
}

// StopJobs removes all jobs before shutdown
func StopJobs() {
	gocron.Remove(runPlayingMessageJob)
	gocron.Remove(runHolidaysJob)
}

func runPlayingMessageJob() {
	playingmsg := getPlayingMessage()
	if playingmsg != nil {
		_, time := gocron.NextRun()
		fmt.Printf("Playing %s\n", *playingmsg)
		fmt.Printf("Next game in %s\n", time.String())
		discord.UpdateStatus(0, *playingmsg)
	}
}

func getPlayingMessage() *string {
	if misc.Config.PlayingWith != nil {
		len := len(misc.Config.PlayingWith)
		if len > 0 {
			return misc.Config.PlayingWith[rand.Intn(len)]
		}
	}

	return nil
}

func runHolidaysJob() {
	time := time.Now()
	if time.Day() == 25 && time.Month() == 12 {
		// christmas holiday
		broadcastMessage("É vos desejado um feliz natal recheado de um óptimo estudo!")
	} else if time.Day() == 1 && time.Month() == 1 {
		// new year
		broadcastMessage("`Feliz ano novo! Que este ano traga uma época de exames com bastantes ~~negativas~~ positivas!`")
	}
}

func broadcastMessage(message string) {
	bot := discord.State.User.ID
	for _, guild := range discord.State.Guilds {
		channels, _ := discord.GuildChannels(guild.ID)
		for _, channel := range channels {
			if channel.Type == discordgo.ChannelTypeGuildText && !channel.NSFW {
				permissions, err := discord.State.UserChannelPermissions(bot, channel.ID)
				if err == nil && permissions&discordgo.PermissionSendMessages != 0 {
					_, err := discord.ChannelMessageSend(channel.ID, message)
					if err != nil {
						fmt.Println("error while broadcasting", err)
					}

					break
				}
			}
		}
	}
}
