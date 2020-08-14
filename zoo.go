package main

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	WikiEndpoint = "https://hypixel-skyblock.fandom.com/wiki/Traveling_Zoo"
)

func fetchLatestPets() (string, []string, error) {
	resp, err := http.Get(WikiEndpoint)
	if err != nil {
		return "", nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		return "", nil, errors.New(fmt.Sprintf("invalid status code: %d", resp.StatusCode))
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", nil, err
	}

	table := doc.Find("table.wikitable tbody")

	// Date.
	date := table.Children().Eq(1).Children().First().Text()

	// All possible pets.
	allPets := table.Children().First().Children().Slice(2, goquery.ToEnd).Map(func(i int, s *goquery.Selection) string {
		pet := strings.ReplaceAll(s.Text(), "Pet", "")
		return strings.TrimSpace(pet)
	})

	// Available pets.
	var availablePets []string
	table.Children().Eq(1).Children().Slice(2, goquery.ToEnd).Each(func(i int, s *goquery.Selection) {
		if s.Children().First().HasClass("blankCell") {
			return
		}

		rarity := strings.TrimSpace(s.Find("b").Text())
		name := allPets[i]

		availablePets = append(availablePets, fmt.Sprintf("%s %s %s", findEmoji(name), rarity, name))
	})
	sort.Sort(byRarity(availablePets))

	return date, availablePets, nil
}

func findEmoji(pet string) string {
	switch pet {
	case "Blue Whale":
		return ":whale2:"
	case "Lion":
		return ":lion_face:"
	case "Tiger":
		return ":tiger:"
	case "Giraffe":
		return ":giraffe:"
	case "Monkey":
		return ":monkey_face:"
	case "Elephant":
		return ":elephant:"
	default:
		return ":question:"
	}
}

type byRarity []string

func (s byRarity) Len() int {
	return len(s)
}
func (s byRarity) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s byRarity) Less(i, j int) bool {
	match := func(name string) int {
		if strings.Contains(name, "Common") {
			return 0
		}
		if strings.Contains(name, "Uncommon") {
			return 1
		}
		if strings.Contains(name, "Rare") {
			return 2
		}
		if strings.Contains(name, "Epic") {
			return 3
		}
		if strings.Contains(name, "Legendary") {
			return 4
		}
		return 64
	}

	return match(s[i]) < match(s[j])
}
