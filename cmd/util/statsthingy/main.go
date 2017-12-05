package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/sarifsystems/sarif/core"
	"github.com/sarifsystems/sarif/sarif"
)

var stats = map[string]interface{}{}

func main() {
	app := core.NewApp("sarif", "tars")
	app.Init()

	c, err := app.ClientDial(sarif.ClientInfo{
		Name: "statsthingy/" + sarif.GenerateId(),
	})
	app.Must(err)

	if f, err := os.Open("stats.json"); err == nil {
		json.NewDecoder(f).Decode(&stats)
		f.Close()
	}

	c.Subscribe("stats", "", func(msg sarif.Message) {
		key := "default"
		if strings.HasPrefix(msg.Action, "stats/") {
			key = strings.TrimPrefix(msg.Action, "stats/")
		}

		var p interface{}
		msg.DecodePayload(&p)
		stats[key] = p
		writeStats()
	})
}

func writeStats() {
	b, err := json.Marshal(stats)
	if err != nil {
		log.Fatal(err)
	}
	err = WriteFileAtomic("stats.json", b, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func WriteFileAtomic(filename string, data []byte, perm os.FileMode) error {
	dir, name := path.Split(filename)
	f, err := ioutil.TempFile(dir, name)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err == nil {
		err = f.Sync()
	}
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if permErr := os.Chmod(f.Name(), perm); err == nil {
		err = permErr
	}
	if err == nil {
		err = os.Rename(f.Name(), filename)
	}
	// Any err should result in full cleanup.
	if err != nil {
		os.Remove(f.Name())
	}
	return err
}
