package handlers

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// CommandHandler represents the function of a command
type CommandHandler func(s *discordgo.Session, m *discordgo.MessageCreate, args []string)

// Command represents each command
type Command struct {
	Name        string
	Handler     CommandHandler
	MinArgs     int
	Permission  int
	Description *string
	GuildOnly   bool
}

// Commands stores all the bot commands
var Commands = make(map[string]*Command)

// RegisterCommand allows to register a new command to the bot
func RegisterCommand(name string, handler CommandHandler) *Command {
	command := &Command{
		Name:      name,
		Handler:   handler,
		GuildOnly: false,
	}

	Commands[name] = command
	return command
}

// SetMinArgs allows to set a minimum number of arguments for the command to be successful
func (c *Command) SetMinArgs(args int) *Command {
	c.MinArgs = args
	return c
}

// SetPermission allows to set a permission for the command
func (c *Command) SetPermission(permission int) *Command {
	c.Permission = permission
	return c
}

// SetDescription allows to set a description for the command
// This description will be available on the help command
func (c *Command) SetDescription(description string) *Command {
	c.Description = &description
	return c
}

// SetGuildOnly allows for the command to be only executed while inside a guild
func (c *Command) SetGuildOnly(guildOnly bool) *Command {
	c.GuildOnly = guildOnly
	return c
}

// ParseCommand finds a command by its name and executes it
// it returns a boolean which is true if the message has a command
func ParseCommand(prefix string, s *discordgo.Session, m *discordgo.MessageCreate) bool {
	content := m.Content
	if strings.HasPrefix(content, prefix) {
		content = content[1:]
		splitted := strings.SplitN(content, " ", 2)
		command := Commands[splitted[0]]
		if command == nil {
			s.ChannelMessageSend(m.ChannelID, "Este comando não está disponível!")
			return true
		}

		if m.GuildID == "" {
			if command.GuildOnly {
				s.ChannelMessageSend(m.ChannelID, "Este comando apenas pode ser executado numa guild")
				return true
			}
		} else {
			if !hasPermission(s, m.Member, m.GuildID, command.Permission) {
				s.ChannelMessageSend(m.ChannelID, "Não tens permissões suficientes para executar este comando!")
				return true
			}
		}

		var args []string
		var argslen int
		if len(splitted) == 2 {
			args = strings.Split(splitted[1], " ")
			argslen = len(args)
		}

		if argslen < command.MinArgs {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Este comando necessita de pelo menos %d argumento(s)", command.MinArgs))
			return true
		}

		command.Handler(s, m, args)
		return true
	}

	return false
}

func hasPermission(s *discordgo.Session, member *discordgo.Member, guildID string, permission int) bool {
	for _, roleID := range member.Roles {
		role, err := s.State.Role(guildID, roleID)
		if err != nil {
			return false
		}

		if role.Permissions&permission == permission {
			return true
		}
	}

	return false
}
