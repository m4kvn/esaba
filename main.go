package main

import (
	"github.com/nlopes/slack"
	"github.com/joho/godotenv"
	"flag"
	"os"
	"log"
	"regexp"
	"fmt"
	"strings"
	"strconv"
	"time"
)

func main() {
	flags := loadFlags()
	api := slack.New(flags.SlackBotToken)

	client, err := api.AuthTest()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	} else {
		log.Println(SlackAuthTestIsOk)
		log.Printf("%#v\n", client)
	}

	channels := map[string]string{}
	params := slack.NewPostMessageParameters()
	params.AsUser = true

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	for {
		select {
		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.MessageEvent:
				regexStr := fmt.Sprintf(`^(<@%s> 神降臨)$`, client.UserID)
				if regexp.MustCompile(regexStr).MatchString(ev.Text) {
					if _, ok := channels[ev.Channel]; !ok {
						channels[ev.Channel] = ev.Timestamp
						api.PostMessage(ev.Channel, "はじまるよ！", params)
						break
					}
				}
				regexStr = fmt.Sprintf(`^(<@%s> 終わり)$`, client.UserID)
				if regexp.MustCompile(regexStr).MatchString(ev.Text) {
					if _, ok := channels[ev.Channel]; ok {
						startTime := channels[ev.Channel]
						endTime := ev.Timestamp
						getHistory(api, ev.Channel, startTime, endTime)
						api.PostMessage(ev.Channel, "おわり！", params)
						delete(channels, ev.Channel)
						break
					}
				}
			}
		}
	}
}

func getHistory(api *slack.Client, channel string, start string, end string) {
	history, err := api.GetChannelHistory(channel, slack.HistoryParameters{
		Latest: end,
		Oldest: start,
		Count:  1000,
	})
	if err != nil {
		log.Println(err)
		return
	}
	users := map[string]*slack.User{}
	for _, message := range history.Messages {
		user := message.User
		if _, ok := users[user]; !ok {
			userInfo, err := api.GetUserInfo(user)
			if err != nil {
				log.Println(err)
				break
			}
			users[user] = userInfo
		}
		text := message.Text
		unixTime := strings.Split(message.Timestamp, ".")
		sec, _ := strconv.ParseInt(unixTime[0], 10, 64)
		nsec, _ := strconv.ParseInt(unixTime[1], 10, 64)
		log.Printf("[%s] user: %s, text: %s\n",
			time.Unix(sec, nsec).Format("2006/1/2 15:04:05"), users[user].Profile.DisplayName, text)
	}
}

type Flag struct {
	SlackBotToken string
}

func loadFlags() Flag {
	godotenv.Load()
	slackBotToken := flag.String("token", os.Getenv(FlagSlackBotToken), FlagTokenDescription)
	flag.Parse()

	if *slackBotToken == "" {
		log.Println(SlackTokenIsRequire)
		os.Exit(1)
	}

	return Flag{
		SlackBotToken: *slackBotToken,
	}
}
