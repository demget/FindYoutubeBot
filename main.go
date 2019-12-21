package main

import (
	"log"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/SerhiiCho/timeago"
	tb "github.com/demget/telebot"
)

func main() {
	tmpl := &tb.TemplateText{
		Dir: "data",

		Funcs: template.FuncMap{
			"timeago": func(at time.Time) string {
				return timeago.Take(at.Format("2006-01-02 15:04:05"))
			},
		},
	}

	pref, err := tb.NewSettings("bot.json", tmpl)
	if err != nil {
		log.Fatalln(err)
	}
	pref.Token = os.Getenv("TOKEN")

	b, err := tb.NewBot(pref)
	if err != nil {
		log.Fatalln(err)
	}

	b.Handle("/start", onStart(b))
	b.Handle(tb.OnText, onText(b))
	b.Handle(tb.OnQuery, onQuery(b))
	b.Handle(b.InlineButton("get"), onGet(b))

	b.Poller = tb.NewMiddlewarePoller(b.Poller, onUpdate(b))
	b.Start()
}

func onUpdate(b *tb.Bot) func(u *tb.Update) bool {
	return func(u *tb.Update) bool {
		if u.Message != nil {
			log.Println(u.Message.Sender.ID, u.Message.Text)
		}
		if u.Callback != nil {
			data := strings.TrimPrefix(u.Callback.Data, "\f")
			log.Println(u.Callback.Sender.ID, data)
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

func onText(b *tb.Bot) func(m *tb.Message) {
	return func(m *tb.Message) {
		videos, err := searchVideos(m.Text, 10)
		if err != nil {
			log.Println(err)
			return
		}

		markup := &tb.ReplyMarkup{}
		markup.InlineKeyboard = make([][]tb.InlineButton, 2)

		for i, video := range videos {
			btn := b.InlineButton("get", struct {
				Index int
				SearchResult
			}{
				Index:        i + 1,
				SearchResult: video,
			})

			ind := i / 5 // 5 buttons per row
			markup.InlineKeyboard[ind] = append(markup.InlineKeyboard[ind], *btn)
		}

		_, err = b.Send(m.Sender,
			b.Text("search", videos),
			markup, tb.ModeMarkdown)
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func onGet(b *tb.Bot) func(c *tb.Callback) {
	return func(c *tb.Callback) {
		b.Send(c.Sender, "https://youtu.be/"+c.Data)
		b.Respond(c)
	}
}

func onQuery(b *tb.Bot) func(q *tb.Query) {
	return func(q *tb.Query) {
		if q.Text == "" {
			return
		}

		videos, err := searchVideos(q.Text, 25)
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
