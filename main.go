package main

import (
	_ "github.com/mattn/go-sqlite3"
	"github.com/nlopes/slack"
	"github.com/joho/godotenv"
	"flag"
	"os"
	"log"
	"regexp"
	"fmt"
	"database/sql"
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
					if !regexp.MustCompile(regexStr).MatchString(ev.Text) {
						break
					}
					isStart = true
					channel = ev.Channel
					api.PostMessage(channel, "マジ卍", slack.NewPostMessageParameters())
				}
			default:
			}
		}
	}
}

type FileSaver struct {
	name string
	file *os.File
	db   *sql.DB
}

func NewFileSaver(name string) FileSaver {
	fileSaver := FileSaver{name: name}

	file, err := os.OpenFile(name+".db", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Println(err)
		return fileSaver
	}
	fileSaver.file = file

	db, err := sql.Open("sqlite3", name+".db")
	if err != nil {
		log.Println(err)
		return fileSaver
	}
	fileSaver.db = db

	query := "create table message ("
	query += "id integer primary key autoincrement"
	query += ", user varchar(255)"
	query += ", text text"
	query += ")"

	fileSaver.dbExec(query)

	return fileSaver
}

func (s *FileSaver) dbExec(q string) {
	if _, err := s.db.Exec(q); err != nil {
		log.Println(err)
		return
	}
}

func (s *FileSaver) Write(ev *slack.MessageEvent) {

}

func (s *FileSaver) Close() {
	s.db.Close()
	s.file.Close()
	os.Remove(s.name)
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
