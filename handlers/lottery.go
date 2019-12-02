package handlers

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/xploitedd/leic-discordbot/misc"
)

var lock bool = false

// RunLottery runs the lottery game
func RunLottery(s *discordgo.Session, m *discordgo.MessageCreate) {
	if lock {
		return
	}

	lock = true

	guild, _ := s.Guild(m.GuildID)
	emojislen := len(guild.Emojis)
	var msg *discordgo.Message
	var err error
	if emojislen < 4 {
		msg, err = s.ChannelMessageSend(m.ChannelID, "`A jogar na lotaria...`")
		if err != nil {
			return
		}

		time.Sleep(2000 * time.Millisecond)
	} else {
		fmt.Println(guild.Emojis[0].MessageFormat())
		msg, err = s.ChannelMessageSend(m.ChannelID, guild.Emojis[rand.Intn(emojislen)].MessageFormat())
		if err != nil {
			return
		}

		time.Sleep(500 * time.Millisecond)
		msg, _ = s.ChannelMessageEdit(m.ChannelID, msg.ID, msg.Content+guild.Emojis[rand.Intn(emojislen)].MessageFormat())
		time.Sleep(500 * time.Millisecond)
		msg, _ = s.ChannelMessageEdit(m.ChannelID, msg.ID, msg.Content+guild.Emojis[rand.Intn(emojislen)].MessageFormat())
		time.Sleep(500 * time.Millisecond)
		msg, _ = s.ChannelMessageEdit(m.ChannelID, msg.ID, msg.Content+guild.Emojis[rand.Intn(emojislen)].MessageFormat())
		time.Sleep(500 * time.Millisecond)
	}

	number := rand.Intn(101)
	if number > 89 || m.Author.ID == *misc.Config.OwnerID {
		s.ChannelMessageEdit(m.ChannelID, msg.ID, "Ganhaste a Lotaria!")
	} else {
		s.ChannelMessageEdit(m.ChannelID, msg.ID, "Perdeste a Lotaria!")
	}

	lock = false
}
