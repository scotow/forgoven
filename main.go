package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/scotow/notigo"
)

var opts struct {
	ApiKeys         []string      `short:"k" description:"Hypixel API key(s)" required:"true"`
	CheckInterval   time.Duration `short:"d" description:"Time between two checks" default:"1m"`
	DiscordWebhook  string        `short:"w" description:"Webhook url used to notify users on discord"`
	Users           []string      `short:"u" description:"USERNAME|MINECRAFT_UUID:SKYBLOCK_PROFILE:DISCORD_USER_ID|NOTIGO_KEY:ITEM,...:AUCTION,..." required:"true"`
	DiscordBotToken string        `short:"t" description:"Discord token used to update channel topic with online players"`
	DiscordChannel  string        `short:"c" description:"Discord channel id used to update channel topic with online players"`
	Zoo             bool          `short:"z" description:"Check for Zoo pet and send it on Discord"`
}

var (
	keys Keys

	discordNameRegex    = regexp.MustCompile(`^\d{18}$`)
	discordNotifier     *DiscordNotifier
	discordTopicChanger *DiscordTopicChanger
)

func checkForgotItems(user *User) {
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
			err := discordNotifier.sendForget(user.notif, itemsStr, len(items))
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
	if len(user.items) == 0 {
		return
	}

	log.Println("items check loop started")
	for {
		checkForgotItems(user)
		time.Sleep(opts.CheckInterval)
	}
}

func checkAuctions(user *User) {
	newAuctions, err := user.newCompletedAuctions(keys.nextKey())
	if err != nil {
		log.Println(err)
	}

	if user.online {
		return
	}

	if len(newAuctions) == 0 {
		return
	}

	itemsStr := strings.Join(newAuctions, ", ")
	if discordNotifier != nil && discordNameRegex.MatchString(user.notif) {
		err := discordNotifier.sendAuctionCompleted(user.notif, itemsStr, len(newAuctions))
		if err != nil {
			log.Println(err)
		}
	} else {
		n := notigo.NewNotification("Hypixel - Skyblock", fmt.Sprintf("Your following %s has been sold at the auction house: %s.", plural("item", len(newAuctions)), itemsStr))
		key := notigo.Key(user.notif)
		err := key.Send(n)
		if err != nil {
			log.Println(err)
		}
	}
}

func auctionsCheckLoop(user *User) {
	if len(user.auctions) == 0 {
		return
	}

	log.Println("auctions check loop started")
	for {
		checkAuctions(user)
		time.Sleep(opts.CheckInterval)
	}
}

func onlineCheckLoop(users []*User) {
	log.Println("online check loop started")

	lastOnlineStr := onlineString(users)
	updateTopic(lastOnlineStr)

	for {
		time.Sleep(opts.CheckInterval)

		online := onlineString(users)
		if online != lastOnlineStr {
			updateTopic(online)
			lastOnlineStr = online
		}
	}
}

func zooCheckLoop() {
	log.Println("zoo check loop started")

	lastDate, _, err := fetchLatestPets()
	if err != nil {
		log.Println(err)
	}

	for {
		time.Sleep(time.Minute)

		currentDate, pets, err := fetchLatestPets()
		if err != nil {
			log.Println(err)
			continue
		}

		if currentDate == lastDate {
			continue
		}
		lastDate = currentDate

		err = discordNotifier.sendZoo(pets)
		if err != nil {
			log.Println(err)
		}
	}
}

func main() {
	parser := flags.NewParser(&opts, flags.Default)
	parser.Name = "forgoven"
	parser.Usage = "-k HYPIXEL_API_KEY... [-d CHECK_INTERVAL] [-w DISCORD_WEBHOOK_URL [-z]] [-t DISCORD_BOT_TOKEN -c DISCORD_CHANNEL_ID] -u USERNAME|MINECRAFT_UUID:SKYBLOCK_PROFILE:DISCORD_USER_ID|NOTIGO_KEY:ITEM,...:AUCTION,... ..."

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

		if opts.Zoo {
			go zooCheckLoop()
		}
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
		go auctionsCheckLoop(u)
	}

	if opts.DiscordBotToken != "" && opts.DiscordChannel != "" {
		discordTopicChanger = NewDiscordTopicChanger(opts.DiscordBotToken, opts.DiscordChannel)
		go onlineCheckLoop(users)
	}

	<-make(chan struct{})
}
