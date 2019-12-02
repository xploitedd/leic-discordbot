package handlers

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"

	dialogflow "cloud.google.com/go/dialogflow/apiv2"
	"google.golang.org/api/option"
	dialogflowpb "google.golang.org/genproto/googleapis/cloud/dialogflow/v2"

	misc "github.com/xploitedd/leic-discord/misc"
)

func login() (*dialogflow.SessionsClient, error) {
	ctx := context.Background()
	sessionClient, err := dialogflow.NewSessionsClient(ctx, option.WithCredentialsFile("dialogflow.json"))
	return sessionClient, err
}

// SendTextQuery allows to communicate with the bot via text
func SendTextQuery(s *discordgo.Session, m *discordgo.MessageCreate, query string) {
	session, err := login()
	defer session.Close()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Ocorreu um erro ao comunicar com o bot!")
		return
	}

	sessionPath := fmt.Sprintf("projects/%s/agent/sessions/%s", *misc.Config.DialogFlowID, m.Author.ID)
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
}
