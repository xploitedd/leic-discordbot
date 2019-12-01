package misc

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
)

// Quote representation
// To be used when a quote is requested by the user
type Quote struct {
	Emote *string `json:"emote"`
	Quote *string `json:"quote"`
}

var quotes []Quote

// LoadQuotes loads all the quotes into an array
func LoadQuotes() error {
	data, err := ioutil.ReadFile("quotes.json")
	if err != nil {
		return err
	}

	quotes = make([]Quote, 0)
	err = json.Unmarshal(data, &quotes)
	if err != nil {
		return err
	}

	return nil
}

// RandomQuote obtains a random quote from the array
func RandomQuote() *Quote {
	if quotes == nil {
		return nil
	}

	len := len(quotes)
	if len == 0 {
		return nil
	}

	idx := rand.Intn(len)
	return &quotes[idx]
}
