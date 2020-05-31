package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type DiscordNotifier struct {
	url string
}

type DiscordMessage struct {
	Content string `json:"content"`
}

func (dn *DiscordNotifier) send(user string, itemsStr string, count int) error {
	message := DiscordMessage{
		Content: fmt.Sprintf("<@%s>, you still have the following %s on you: %s.", user, plural("item", count), itemsStr),
	}
	data, err := json.Marshal(&message)
	if err != nil {
		return err
	}

	_, err = http.Post(dn.url, "application/json", bytes.NewBuffer(data))
	return err
}
