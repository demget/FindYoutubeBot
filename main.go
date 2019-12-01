package main

import (
	"log"
	"os"

	tb "github.com/demget/telebot"
)

func main() {
	pref, err := tb.NewSettings("bot.json", "data")
	if err != nil {
		log.Fatalln(err)
	}
	pref.Token = os.Getenv("TOKEN")

	b, err := tb.NewBot(pref)
	if err != nil {
		log.Fatalln(err)
	}

	b.Handle("/start", onStart(b))
	b.Handle(tb.OnQuery, onQuery(b))

	b.Poller = tb.NewMiddlewarePoller(b.Poller, onUpdate(b))
	b.Start()
}

func onUpdate(b *tb.Bot) func(u *tb.Update) bool {
	return func(u *tb.Update) bool {
		if u.Message != nil {
			log.Println(u.Message.Sender.ID, u.Message.Text)
		}
		if u.Query != nil {
			log.Println(u.Query.From.ID, u.Query.Text)
		}
		return true
	}
}

func onStart(b *tb.Bot) func(m *tb.Message) {
	return func(m *tb.Message) {
		b.Send(m.Sender,
			b.Text("start", m.Sender.FirstName),
			b.InlineMarkup("search"),
			tb.ModeHTML)
	}
}

func onQuery(b *tb.Bot) func(q *tb.Query) {
	return func(q *tb.Query) {
		if q.Text == "" {
			return
		}

		videos, err := searchVideos(q.Text)
		if err != nil {
			log.Println(err)
			return
		}

		var results tb.Results
		for _, video := range videos {
			results = append(results, b.InlineResult("video", video))
		}

		err = b.Answer(q, &tb.QueryResponse{Results: results})
		if err != nil {
			log.Println(err)
		}
	}
}
