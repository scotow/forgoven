package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/scotow/notigo"
)

var opts struct {
	ApiKeys         []string      `short:"k" description:"Hypixel API key(s)" required:"true"`
	CheckInterval   time.Duration `short:"d" description:"Time between two checks" default:"1m"`
	DiscordWebhook  string        `short:"w" description:"Webhook url used to notify users on discord"`
	Users           []string      `short:"u" description:"USERNAME|MINECRAFT_UUID:SKYBLOCK_PROFILE:DISCORD_USER_ID|NOTIGO_KEY:ITEM..." required:"true"`
	DiscordBotToken string        `short:"t" description:"Discord token used to update channel topic with online players"`
	DiscordChannel  string        `short:"c" description:"Discord channel id used to update channel topic with online players"`
}

var (
	keys Keys

	discordNameRegex    = regexp.MustCompile(`^\d{18}$`)
	discordNotifier     *DiscordNotifier
	discordTopicChanger *DiscordTopicChanger
)

func check(user *User) {
	err := user.updateOnline(keys.nextKey())
	if err != nil {
		log.Println(err)
		return
	}

	if user.online {
		user.last = ""
		return
	}

	items, err := user.hasItems(keys.nextKey())
	if err != nil {
		log.Println(err)
		return
	}

	itemsStr := strings.Join(items, ", ")
	if len(items) > 0 && user.last != itemsStr {
		if discordNotifier != nil && discordNameRegex.MatchString(user.notif) {
			err := discordNotifier.send(user.notif, itemsStr, len(items))
			if err != nil {
				log.Println(err)
			}
		} else {
			n := notigo.NewNotification("Hypixel - Skyblock", fmt.Sprintf("You still have the following %s on you: %s.", plural("item", len(items)), itemsStr))
			key := notigo.Key(user.notif)
			err := key.Send(n)
			if err != nil {
				log.Println(err)
			}
		}
	}
	user.last = itemsStr
}

func plural(s string, count int) string {
	if count >= 2 {
		return s + "s"
	} else {
		return s
	}
}

func itemsCheckLoop(user *User) {
	for {
		check(user)
		time.Sleep(opts.CheckInterval)
	}
}

func onlineCheckLoop(users []*User) {
	lastOnlineStr := ""
	for {
		online := make([]string, 0)
		for _, u := range users {
			if u.online {
				online = append(online, u.name)
			}
		}
		onlineStr := strings.Join(online, ", ")

		if onlineStr != lastOnlineStr {
			if onlineStr == "" {
				discordTopicChanger.change("")
			} else {
				discordTopicChanger.change(fmt.Sprintf("Online on Hypixel: %s", onlineStr))
			}
			lastOnlineStr = onlineStr
		}
		time.Sleep(opts.CheckInterval)
	}
}

func main() {
	parser := flags.NewParser(&opts, flags.Default)
	parser.Name = "forgoven"
	parser.Usage = "-k HYPIXEL_API_KEY... [-d CHECK_INTERVAL] [-w DISCORD_WEBHOOK_URL] [-t DISCORD_BOT_TOKEN -c DISCORD_CHANNEL_ID] -u USERNAME|MINECRAFT_UUID:SKYBLOCK_PROFILE:DISCORD_USER_ID|NOTIGO_KEY:ITEM... ..."

	_, err := parser.Parse()
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

	if len(opts.ApiKeys) == 0 {
		log.Fatalln("invalid hypixel api key")
	}
	keys = NewKeys(opts.ApiKeys)

	if opts.DiscordWebhook != "" {
		discordNotifier = &DiscordNotifier{url: opts.DiscordWebhook}
	}

	if len(opts.Users) == 0 {
		log.Fatalln("invalid number of user")
	}

	users := make([]*User, 0, len(opts.Users))
	for _, a := range opts.Users {
		u, err := parseUser(a, keys.nextKey())
		if err != nil {
			log.Fatalln(err)
		}
		users = append(users, u)
	}

	for _, u := range users {
		go itemsCheckLoop(u)
	}

	if opts.DiscordBotToken != "" && opts.DiscordChannel != "" {
		discordTopicChanger = NewDiscordTopicChanger(opts.DiscordBotToken, opts.DiscordChannel)
		go onlineCheckLoop(users)
	}

	<-make(chan struct{})
}
