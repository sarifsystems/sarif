package main

import (
	"log"

	"github.com/mattn/go-gtk/gtk"
	"github.com/mattn/go-gtk/glib"

	"github.com/xconstruct/stark"
	"github.com/xconstruct/stark/service"

	_ "github.com/xconstruct/stark/transport/net"
)

func main() {
	serv := service.New(service.Info{
		Name: "launcher",
	})
	err := serv.Dial("tcp://")
	if err != nil {
		log.Fatalln(err)
	}

	gtk.Init(nil)
	window := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
	window.SetTitle("Stark Launcher")
	window.SetIconName("gtk-dialog-info")
	window.Connect("destroy", func(ctx *glib.CallbackContext) {
		gtk.MainQuit()
	}, "foo")

	hbox := gtk.NewHBox(false, 1)
	entry := gtk.NewEntry()
	entry.SetText("Hello, world")
	entry.GrabFocus()
	hbox.Add(entry)

	entry.Connect("activate", func(ctx *glib.CallbackContext) {
		text := entry.GetText()
		if text == "" {
			return
		}

		msg := stark.NewMessage()
		msg.Action = "natural.process"
		msg.Message = entry.GetText()
		err := serv.Write(msg)
		if err != nil {
			log.Println(err)
		}
		window.Destroy()
	}, "foo")

	window.Add(hbox)
	window.SetSizeRequest(400, 50)
	window.Move(500, 500)
	window.ShowAll()
	gtk.Main()
}
