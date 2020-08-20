package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type DiscordNotifier struct {
	url string
}

type DiscordMessage struct {
	Content string `json:"content"`
}

func (dn *DiscordNotifier) send(str string) error {
	message := DiscordMessage{
		Content: str,
	}
	data, err := json.Marshal(&message)
	if err != nil {
		return err
	}

	_, err = http.Post(dn.url, "application/json", bytes.NewBuffer(data))
	return err
}

func (dn *DiscordNotifier) sendForget(user string, itemsStr string, count int) error {
	return dn.send(fmt.Sprintf("<@%s>, you still have the following %s on you: %s.", user, plural("item", count), itemsStr))
}

func (dn *DiscordNotifier) sendAuctionCompleted(user string, itemsStr string, count int) error {
	return dn.send(fmt.Sprintf("<@%s>, your following %s has been sold at the auctions house: %s.", user, plural("item", count), itemsStr))
}

func (dn *DiscordNotifier) sendZoo(pets []string) error {
	for i, p := range pets {
		pets[i] = fmt.Sprintf("    - %s", p)
	}

	return dn.send(fmt.Sprintf("Oringo has brought the following pets to the village:\n%s\n\nCheck the prices here: https://hypixel-skyblock.fandom.com/wiki/Traveling_Zoo.", strings.Join(pets, "\n")))
}

type DiscordTopicChanger struct {
	botToken string
	channel  string
	wait     *time.Timer
	lock     sync.Mutex
}

func NewDiscordTopicChanger(botToken string, channel string) *DiscordTopicChanger {
	return &DiscordTopicChanger{
		botToken: botToken,
		channel:  channel,
		wait:     nil,
		lock:     sync.Mutex{},
	}
}

func (dpc *DiscordTopicChanger) change(topic string) {
	dpc.lock.Lock()
	defer dpc.lock.Unlock()

	if dpc.wait != nil {
		_ = dpc.wait.Stop()
		dpc.wait = nil
	}

	payload, err := json.Marshal(struct {
		Topic string `json:"topic"`
	}{
		Topic: topic,
	})
	if err != nil {
		log.Println(err)
	}

	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("https://discord.com/api/channels/%s", dpc.channel), bytes.NewBuffer(payload))
	if err != nil {
		log.Println(err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bot %s", dpc.botToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Forgoven")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		sec, _ := strconv.Atoi(resp.Header.Get("X-RateLimit-Reset-After"))
		dpc.wait = time.AfterFunc(time.Second*time.Duration(sec), func() {
			dpc.change(topic)
		})
	}
}
