package main

import (
	"flag"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/scotow/notigo"
)

var (
	apiKey         = flag.String("k", "", "hypixel API key")
	checkInterval  = flag.Duration("d", time.Minute, "time between two checks")
	discordWebhook = flag.String("w", "", "webhook url used to notify users on discord")
)

var (
	discordNameRegex = regexp.MustCompile(`^\d{18}$`)
	discordNotifier  *DiscordNotifier
)

func check(user *User) {
	online, err := user.isOnline(*apiKey)
	if err != nil {
		log.Println(err)
		return
	}

	if online {
		user.last = ""
		return
	}

	items, err := user.hasItems(*apiKey)
	if err != nil {
		log.Println(err)
		return
	}

	itemsStr := strings.Join(items, ", ")
	if len(items) > 0 && user.last != itemsStr {
		if discordNotifier != nil && discordNameRegex.MatchString(user.notif) {
			err := discordNotifier.send(user.notif, itemsStr)
			if err != nil {
				log.Println(err)
			}
		} else {
			n := notigo.NewNotification("Hypixel - Skyblock", fmt.Sprintf("You still have the following item(s) on you: %s.", itemsStr))
			key := notigo.Key(user.notif)
			err := key.Send(n)
			if err != nil {
				log.Println(err)
			}
		}
	}
	user.last = itemsStr
}

func checkLoop(user *User) {
	for {
		check(user)
		time.Sleep(*checkInterval)
	}
}

func main() {
	flag.Parse()

	if *apiKey == "" {
		log.Fatalln("invalid hypixel api key")
	}

	if *discordWebhook != "" {
		discordNotifier = &DiscordNotifier{url: *discordWebhook}
	}

	args := flag.Args()
	if len(args) == 0 {
		log.Fatalln("invalid number of user")
	}

	users := make([]*User, 0, len(args))
	for _, a := range args {
		u, err := parseUser(a)
		if err != nil {
			log.Fatalln(err)
		}
		users = append(users, u)
	}

	for _, u := range users {
		go checkLoop(u)
	}

	<-make(chan struct{})
}
