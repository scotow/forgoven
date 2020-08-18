package main

import (
	"fmt"
	"strings"
)

func onlineString(users []*User) string {
	online := make([]string, 0)
	for _, u := range users {
		if u.online {
			online = append(online, u.name)
		}
	}
	return strings.Join(online, ", ")
}

func updateTopic(onlineStr string) {
	if onlineStr == "" {
		discordTopicChanger.change("")
	} else {
		discordTopicChanger.change(fmt.Sprintf("Online on Hypixel: %s", onlineStr))
	}
}
