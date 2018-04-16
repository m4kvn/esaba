package main

import (
	"github.com/nlopes/slack"
	"github.com/joho/godotenv"
	"flag"
	"os"
	"log"
	"regexp"
	"fmt"
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

	var isStart = false
	var channel string

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	for {
		select {
		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.MessageEvent:
				if isStart {
					if ev.Channel != channel {
						break
					}
					regexStr := fmt.Sprintf(`^(<@%s> 終わり)$`, client.UserID)
					if regexp.MustCompile(regexStr).MatchString(ev.Text) {
						api.PostMessage(channel, ":innocent:", slack.NewPostMessageParameters())
						isStart = false
						channel = ""
						break
					}
					log.Printf("%#v\n", ev)
				} else {
					regexStr := fmt.Sprintf(`^(<@%s> 神降臨)$`, client.UserID)
					if regexp.MustCompile(regexStr).MatchString(ev.Text) {
						isStart = true
						channel = ev.Channel
					}
					api.PostMessage(channel, "マジ卍", slack.NewPostMessageParameters())
				}
			default:
			}
		}
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
