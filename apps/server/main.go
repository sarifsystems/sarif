package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/xconstruct/stark/router"
	"github.com/xconstruct/stark/transport"

	_ "github.com/xconstruct/stark/transport/local"
	_ "github.com/xconstruct/stark/transport/net"

	"github.com/xconstruct/stark/service/mpd"
	"github.com/xconstruct/stark/service/natural"
	"github.com/xconstruct/stark/service/reminder"
	"github.com/xconstruct/stark/service/terminal"
	"github.com/xconstruct/stark/service/xmpp"
)

func main() {
	// Setup router
	router := router.NewRouter("router")

	// Listen on various protocols, as defined in the config
	listeners := getConfigMap("listeners")
	for url, _ := range listeners {
		listener, err := transport.Listen(url)
		if err != nil {
			log.Println(err)
		}
		go router.Listen(listener)
	}

	// Start integrated services
	t := terminal.New("local://")
	t.Start()

	m := mpd.New("local://")
	m.Start()

	n := natural.New("local://")
	n.Start()

	rm := reminder.New("local://")
	rm.Start()

	xs, _ := xmpp.NewService("local://", getConfigMap("xmpp"))
	xs.Start()

	// Loop
	select {}
}

var config map[string]interface{}

func readConfig() {
	file, err := os.Open("config.json")
	if err != nil {
		panic(err)
	}
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		panic(err)
	}
}

func getConfig(field string) interface{} {
	if config == nil {
		readConfig()
	}
	return config[field]
}

func getConfigMap(field string) map[string]interface{} {
	if config == nil {
		readConfig()
	}
	val, _ := config[field].(map[string]interface{})
	return val
}
