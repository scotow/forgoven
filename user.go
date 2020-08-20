package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type User struct {
	id               uuid.UUID
	shortId          string
	name             string
	profile          string
	notif            string
	items            []string
	auctions         []string
	notifiedAuctions []string
	online           bool
	last             string
}

type MojangResponse struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func parseUser(arg string, apiKey string) (*User, error) {
	parts := strings.Split(arg, ":")
	if len(parts) < 3 {
		return nil, errors.New("invalid user config")
	}

	id, err := uuid.Parse(parts[0])
	if err != nil {
		id, err = fetchUuid(parts[0])
		if err != nil {
			return nil, err
		}
	}

	pr, err := fetchProfile(id, apiKey)
	if err != nil {
		return nil, err
	}
	name := pr.Player.Name

	if parts[1] == "" {
		return nil, errors.New("invalid skyblock profile")
	}
	if parts[2] == "" {
		return nil, errors.New("invalid notigo key or discord name")
	}

	items := make([]string, 0)
	if len(parts) >= 4 {
		for _, i := range strings.Split(parts[3], ",") {
			if len(i) > 0 {
				items = append(items, i)
			}
		}
	}

	auctions := make([]string, 0)
	if len(parts) >= 5 {
		for _, a := range strings.Split(parts[4], ",") {
			if len(a) > 0 {
				auctions = append(auctions, a)
			}
		}
	}

	u := new(User)
	u.id = id
	u.shortId = strings.ReplaceAll(u.id.String(), "-", "")
	u.name = name
	u.profile = parts[1]
	u.notif = parts[2]
	u.items = items
	u.auctions = auctions
	u.notifiedAuctions = make([]string, 0)

	return u, nil
}

type OnlineResponse struct {
	Success bool `json:"success"`
	Session struct {
		Online bool `json:"online"`
	} `json:"session"`
}

func (u *User) updateOnline(apiKey string) error {
	resp, err := http.Get(fmt.Sprintf("https://api.hypixel.net/status?key=%s&uuid=%s", apiKey, u.id.String()))
	if err != nil {
		return errors.New("api is offline")
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("invalid status code: %d", resp.StatusCode))
	}

	var or OnlineResponse
	err = json.NewDecoder(resp.Body).Decode(&or)
	if err != nil {
		return err
	}

	if or.Success == false {
		return errors.New("api is offline")
	}

	u.online = or.Session.Online
	return nil
}

type ProfileResponse struct {
	Success bool `json:"success"`
	Player  struct {
		Id    string `json:"uuid"`
		Name  string `json:"displayname"`
		Stats struct {
			SkyBlock struct {
				Profiles map[string]SkyblockProfile `json:"profiles"`
			} `json:"SkyBlock"`
		} `json:"stats"`
	} `json:"player"`
}

type SkyblockProfile struct {
	ProfileId string `json:"profile_id"`
	CuteName  string `json:"cute_name"`
}

type SkyblockResponse struct {
	Success bool `json:"success"`
	Profile struct {
		Members map[string]struct {
			Inventory  SkyblockContainer `json:"inv_contents"`
			Enderchest SkyblockContainer `json:"ender_chest_contents"`
		} `json:"members"`
	} `json:"profile"`
}

type SkyblockAuctionResponse struct {
	Success  bool `json:"success"`
	Auctions []struct {
		Id         string `json:"uuid"`
		Name       string `json:"item_name"`
		Bin        bool   `json:"bin"`
		HighestBid int    `json:"highest_bid_amount"`
	} `json:"auctions"`
}

type SkyblockContainer struct {
	Data string `json:"data"`
}

func fetchProfile(id uuid.UUID, apiKey string) (*ProfileResponse, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.hypixel.net/player?key=%s&uuid=%s", apiKey, id.String()))
	if err != nil {
		return nil, errors.New("api is offline")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("invalid status code: %d", resp.StatusCode))
	}

	var pr ProfileResponse
	err = json.NewDecoder(resp.Body).Decode(&pr)
	if err != nil {
		return nil, err
	}

	if pr.Success == false {
		return nil, errors.New("api is offline")
	}

	return &pr, nil
}

func (u *User) hasItems(apiKey string) ([]string, error) {
	pr, err := fetchProfile(u.id, apiKey)
	if err != nil {
		return nil, err
	}

	if pr.Player.Stats.SkyBlock.Profiles == nil {
		return nil, errors.New("cannot find skyblock profile")
	}

	var sbProfile SkyblockProfile
	found := false
	for _, v := range pr.Player.Stats.SkyBlock.Profiles {
		if v.CuteName == u.profile {
			sbProfile = v
			found = true
			break
		}
	}
	if !found {
		return nil, errors.New("cannot find skyblock profile")
	}

	resp, err := http.Get(fmt.Sprintf("https://api.hypixel.net/skyblock/profile?key=%s&profile=%s", apiKey, sbProfile.ProfileId))
	if err != nil {
		return nil, errors.New("api is offline")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("invalid status code: %d", resp.StatusCode))
	}

	var sr SkyblockResponse
	err = json.NewDecoder(resp.Body).Decode(&sr)
	if err != nil {
		return nil, err
	}

	if pr.Success == false {
		return nil, errors.New("api is offline")
	}

	if sr.Profile.Members == nil {
		return nil, errors.New("cannot find skyblock member")
	}

	member, ok := sr.Profile.Members[pr.Player.Id]
	if !ok {
		return nil, errors.New("cannot find skyblock member")
	}

	ir, err := gzip.NewReader(base64.NewDecoder(base64.StdEncoding, bytes.NewBuffer([]byte(member.Inventory.Data))))
	if err != nil {
		return nil, err
	}

	inv, err := ioutil.ReadAll(ir)
	if err != nil {
		return nil, err
	}

	er, err := gzip.NewReader(base64.NewDecoder(base64.StdEncoding, bytes.NewBuffer([]byte(member.Enderchest.Data))))
	if err != nil {
		return nil, err
	}

	enderchest, err := ioutil.ReadAll(er)
	if err != nil {
		return nil, err
	}

	items := make([]string, 0)
	for _, i := range u.items {
		if bytes.Contains(inv, []byte(i)) || bytes.Contains(enderchest, []byte(i)) {
			items = append(items, i)
		}
	}

	return items, nil
}

func (u *User) newCompletedAuctions(apiKey string) ([]string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.hypixel.net/skyblock/auction?key=%s&player=%s", apiKey, u.shortId))
	if err != nil {
		return nil, errors.New("api is offline")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("invalid status code: %d", resp.StatusCode))
	}

	var sar SkyblockAuctionResponse
	err = json.NewDecoder(resp.Body).Decode(&sar)
	if err != nil {
		return nil, err
	}

	if sar.Success == false {
		return nil, errors.New("api is offline")
	}

	newAuctions := make([]string, 0)
	for _, a := range sar.Auctions {
		if !a.Bin {
			continue
		}
		if a.HighestBid == 0 {
			continue
		}
		if u.alreadyHasNotifiedAuction(a.Id) {
			continue
		}

		newAuctions = append(newAuctions, a.Name)
		u.notifiedAuctions = append(u.notifiedAuctions, a.Id)
	}

	if len(sar.Auctions) == 0 {
		u.notifiedAuctions = make([]string, 0)
	}

	return newAuctions, nil
}

func (u *User) alreadyHasNotifiedAuction(id string) bool {
	for _, na := range u.notifiedAuctions {
		if na == id {
			return true
		}
	}
	return false
}

func fetchUuid(name string) (uuid.UUID, error) {
	var id uuid.UUID

	req, err := http.NewRequest("GET", fmt.Sprintf("http://api.mojang.com/users/profiles/minecraft/%s", name), nil)
	if err != nil {
		return id, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.2; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.90 Safari/537.36")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return id, err
	}

	if resp.StatusCode != http.StatusOK {
		return id, errors.New("invalid mojang response")
	}

	var mr MojangResponse
	err = json.NewDecoder(resp.Body).Decode(&mr)
	if err != nil {
		return id, err
	}

	id, err = uuid.Parse(mr.Id)
	if err != nil {
		return id, err
	}

	return id, nil
}
