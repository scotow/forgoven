package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type DiscordNotifier struct {
	url string
}

type DiscordMessage struct {
	Content string `json:"content"`
}

func (dn *DiscordNotifier) send(user string, items []string) error {
	message := DiscordMessage{
		Content: fmt.Sprintf("<@%s>, you still have the following item(s) on you: %s.", user, strings.Join(items, ", ")),
	}
	data, err := json.Marshal(&message)
	if err != nil {
		return err
	}

	_, err = http.Post(dn.url, "application/json", bytes.NewBuffer(data))
	return err
}
