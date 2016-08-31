// Copyright (C) 2015 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/sarifsystems/sarif/pkg/content"
	"github.com/sarifsystems/sarif/pkg/schema"
	"github.com/sarifsystems/sarif/sarif"
)

type ContentPayload struct {
	Content schema.Content `json:"content,omitempty"`
}

func (app *App) Edit() {
	// Request document from given action
	action := flag.Arg(1)
	putAction := flag.Arg(2)

	msg, ok := <-app.Client.Request(sarif.CreateMessage(action, nil))
	if !ok {
		app.Log.Fatalln("No response received at " + action)
	}
	if msg.IsAction("err") {
		app.Log.Fatalln(msg.Action + ": " + msg.Text)
	}

	var ctp ContentPayload
	app.Must(msg.DecodePayload(&ctp))

	// Extract content from response
	var err error
	if ctp.Content.Url != "" {
		ct, err := content.Get(ctp.Content)
		app.Must(err)
		ctp.Content.Data = ct.Data
	} else if len(msg.Payload.Raw) > 0 {
		var p interface{}
		msg.DecodePayload(&p)
		ctp.Content.Data, err = json.MarshalIndent(p, "", "    ")
		app.Must(err)
	} else {
		ctp.Content.Data = []byte(msg.Text)
	}

	// Create temp file with content
	f, err := tempFile(os.TempDir(), "tars-", "-"+ctp.Content.Name)
	app.Must(err)
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()
	_, err = f.Write(ctp.Content.Data)
	app.Must(err)
	app.Must(f.Sync())
	fi, err := f.Stat()
	app.Must(err)
	lastMod := fi.ModTime()

	// Start editor and continuously check progress
	editor := os.Getenv("EDITOR")
	cmd := exec.Command(editor, f.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	app.Must(cmd.Start())

	stop := make(chan bool, 2)
	go func() {
		app.Must(cmd.Wait())
		stop <- true
	}()

	lastErr := ""
	run := true
	for run {
		select {
		case <-time.After(time.Second):
		case <-stop:
			run = false
		}

		fi, err := f.Stat()
		app.Must(err)
		if lastMod.Before(fi.ModTime()) && putAction != "" {
			lastMod = fi.ModTime()
			_, err := f.Seek(0, 0)
			app.Must(err)
			data, err := ioutil.ReadAll(f)
			app.Must(err)
			ct := content.PutData(data)
			lastErr = ""
			msg, ok := <-app.Client.Request(sarif.CreateMessage(putAction, ContentPayload{ct}))
			if !ok {
				lastErr = "Could not save: no response received at " + ctp.Content.PutAction
			} else if msg.IsAction("err") {
				lastErr = msg.Action + ": " + msg.Text
			}
		}
	}

	if lastErr != "" {
		app.Log.Fatalln(lastErr)
	}
}

func tempFile(dir, prefix, suffix string) (*os.File, error) {
	for index := 1; index < 10000; index++ {
		path := filepath.Join(dir, fmt.Sprintf("%s%03d%s", prefix, index, suffix))
		if _, err := os.Stat(path); err != nil {
			return os.Create(path)
		}
	}
	return nil, fmt.Errorf("could not create file of the form %s%03d%s", prefix, 1, suffix)
}
