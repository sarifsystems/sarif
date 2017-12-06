package main

import (
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"time"

	"github.com/sarifsystems/sarif/core"
	"github.com/sarifsystems/sarif/pkg/failsafe"
	"github.com/sarifsystems/sarif/sarif"
)

type SMTPConfig struct {
	User      string
	Password  string
	Host      string
	Recipient string
}

type Config struct {
	SMTP     SMTPConfig
	Duration string
}

type App struct {
	*core.App
	Client   *sarif.Client
	Config   Config
	Failsafe *failsafe.Failsafe
}

type DeadlineInfo struct {
	Deadline time.Time
}

func (p DeadlineInfo) Text() string {
	return "The deadline is at " + p.Deadline.Format(time.RFC3339)
}

func New() *App {
	var err error
	app := &App{
		App: core.NewApp("sarif", "failsafe"),
	}
	app.Init()
	app.Client, err = app.ClientDial(sarif.ClientInfo{
		Name: "failsafe/" + sarif.GenerateId(),
	})
	app.Must(err)

	app.App.Config.Get("failsafe", &app.Config)
	dur, err := time.ParseDuration(app.Config.Duration)
	app.Must(err)

	fs := failsafe.New(dur)
	app.Failsafe = fs

	app.Client.Subscribe("failsafe/checkin", "", app.HandleCheckIn)
	app.Client.Subscribe("failsafe/deadline", "", func(msg sarif.Message) {
		app.Client.Reply(msg, sarif.CreateMessage("failsafe/info", DeadlineInfo{
			Deadline: fs.Deadline(),
		}))
	})

	fs.After(-2*time.Hour, func() {
		msg := sarif.CreateMessage("failsafe/warning", nil)
		msg.Destination = "user"
		msg.Text = "Time for your daily check-in! Two hours remaining. Please confirm."
		app.Client.Publish(msg)
	})

	fs.After(-1*time.Hour, func() {
		msg := sarif.CreateMessage("failsafe/warning", nil)
		msg.Destination = "user"
		msg.Text = "Time for your daily check-in! One hour remaining. Please confirm."
		app.Client.Publish(msg)
	})

	fs.After(0, func() {
		msg := sarif.CreateMessage("notifications/new/failsafe/alert", nil)
		msg.Text = "Alert! Deadline exceeded. Failsafe armed. Please check-in immediately."
		app.Client.Publish(msg)
	})

	for i := time.Duration(1); i < 40; i++ {
		j := i * 3 * time.Hour
		fs.After(j, func() {
			remaining := fmt.Sprintf("%d:00 hours remaining until activation.", j/time.Hour)
			app.SendMail("Failsafe Alert", "Deadline exceeded. Failsafe armed. Please check-in immediately.\n\n"+remaining)
		})
	}

	fs.After(30*time.Minute, func() {
		app.SendMail("Failsafe", "You're dead!")
		log.Print("you're dead!")
	})

	return app
}

type TimestampPayload struct {
	Time time.Time `json:"time"`
}

func isRecent(t, base time.Time) bool {
	if t.IsZero() || base.IsZero() {
		return false
	}
	diff := t.Sub(base)
	return diff >= 0 && diff < 30*time.Second
}

func (app *App) HandleCheckIn(msg sarif.Message) {
	if !app.Challenge(msg.Source) {
		app.Client.Reply(msg, sarif.CreateMessage("failsafe/nack", nil))
	}

	t := app.Failsafe.CheckIn()
	app.Client.Reply(msg, sarif.CreateMessage("failsafe/ack", DeadlineInfo{t}))
}

func (app *App) Challenge(device string) bool {
	now := time.Now()

	reply, ok := <-app.Client.Request(sarif.Message{
		Action:      "failsafe/challenge",
		Destination: device,
	})
	if !ok {
		return false
	}

	var p TimestampPayload
	reply.DecodePayload(&p)
	return isRecent(p.Time, now)
}

func (app *App) SendMail(subject, body string) {
	cfg := app.Config.SMTP
	host := cfg.Host + ":587"
	auth := smtp.PlainAuth("", cfg.User, cfg.Password, cfg.Host)

	to := []string{cfg.Recipient}
	msg := "To: " + cfg.Recipient + "\n" +
		`From: "Sarif Failsafe" <` + cfg.User + ">\n" +
		"Subject: " + subject + "\n\n" +
		body

	msg = strings.Replace(msg, "\n", "\r\n", -1)
	err := smtp.SendMail(host, auth, cfg.User, to, []byte(msg))
	if err != nil {
		log.Print(err)
		app.Client.Log("err/internal", err.Error())
	}
}

func (app *App) Run() {
	app.Failsafe.Run()
}

func main() {
	app := New()
	app.Run()
}
