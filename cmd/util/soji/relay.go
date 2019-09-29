package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sarifsystems/sarif/core"
	"github.com/sarifsystems/sarif/sarif"
	"github.com/sarifsystems/sarif/services/schema/store"
)

type Relay struct {
	*app
	store *store.Store
}

type keyValuePayload struct {
	Key   string      `json:"key,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

func (r *Relay) Relay() {
	r.Info.DeviceId = "soji/" + sarif.GenerateId()
	r.Info.ActionChanged = "context/changed"
	r.Info.ActionPull = "store/get/context/current"
	r.Info.Read()

	r.Info.PID = os.Getpid()
	r.Must(r.Info.Write())

	c, err := r.ClientDial(sarif.ClientInfo{
		Name: r.Info.DeviceId,
	})
	r.Must(err)

	r.store = store.New(c)

	if r.Info.ActionChanged != "" {
		c.Subscribe(r.Info.ActionChanged, "", func(msg sarif.Message) {
			var kv keyValuePayload
			msg.DecodePayload(&kv)

			if kv.Key == "" {
				if strings.HasPrefix(msg.Action, r.Info.ActionChanged+"/") {
					kv.Key = strings.TrimPrefix(msg.Action, r.Info.ActionChanged+"/")
				}
			}
			if kv.Value == nil {
				kv.Value = msg.Text
			}
			if kv.Key == "" || kv.Value == nil {
				return
			}

			r.Info.Bar(kv.Key).Value = kv.Value
			r.Must(r.Info.Write())
		})
	}

	fmt.Println(os.Getpid())

	if r.Info.ActionPull != "" {
		tries := 0
		for {
			msg, ok := <-c.Request(sarif.CreateMessage(r.Info.ActionPull, nil))
			if !ok || msg.IsAction("err") {
				tries++
				if tries > 3 {
					break
				}
			} else {
				tries = 0
				var m map[string]interface{}
				msg.DecodePayload(&m)
				r.Merge("", m)
				r.Info.Bar("soji_status").Value = time.Now()
				r.Must(r.Info.Write())
			}
			time.Sleep(5 * time.Minute)
		}
	} else {
		core.WaitUntilInterrupt()
	}

	r.Info.PID = 0
	r.Must(r.Info.Write())
}

func (r *Relay) Merge(prefix string, m map[string]interface{}) {
	for k, v := range m {
		switch vv := v.(type) {
		case map[string]interface{}:
			r.Merge(k+"/", vv)
		default:
			r.Info.Bar(prefix + k).Value = v
		}
	}
}
