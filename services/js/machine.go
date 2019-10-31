// Copyright (C) 2018 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package js

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/ddliu/motto"
	"github.com/robertkrimen/otto"
	"github.com/sarifsystems/sarif/sarif"
)

type Machine struct {
	sarif.Client
	Modules *motto.Motto
	VM      *otto.Otto

	StateLock    sync.Mutex
	OutputBuffer string
	Listeners    []string
}

func NewMachine(c sarif.Client) *Machine {
	m := motto.New()
	return &Machine{
		Modules: m,
		VM:      m.Otto,
		Client:  c,
	}
}

func (m *Machine) Enable() error {
	m.StateLock.Lock()
	defer m.StateLock.Unlock()

	m.VM.Set("print", m.vmPrint)
	m.VM.Set("require", m.vmRequire)
	v, _ := m.VM.Run("console || {}")
	if v.IsObject() {
		v.Object().Set("log", m.vmDebug)
	}

	m.RegisterModule("sarif-proto", map[string]interface{}{
		"subscribe": m.vmSubscribe,
		"publish":   m.vmPublish,
		"request":   m.vmRequest,
		"reply":     m.vmReply,
		"print":     m.vmPrint,
		"dump":      m.vmDebug,
		"debug":     m.vmDebug,
		"device_id": m.DeviceId(),
	})
	return nil
}

func (m *Machine) RegisterModule(name string, module map[string]interface{}) {
	m.Modules.AddModule(name, func(vm *motto.Motto) (otto.Value, error) {
		return vm.ToValue(module)
	})
}

func (m *Machine) Disable() error {
	m.StateLock.Lock()
	defer m.StateLock.Unlock()
	m.VM = nil
	return m.Disconnect()
}

func (m *Machine) vmRequire(call otto.FunctionCall) otto.Value {
	name := call.Argument(0).String()
	mod, err := m.Modules.Require(name, "")
	if err != nil {
		panic(m.VM.MakeTypeError(err.Error()))
	}
	return mod
}

func (m *Machine) vmPrint(call otto.FunctionCall) otto.Value {
	for i, arg := range call.ArgumentList {
		m.OutputBuffer += arg.String()
		if i > 0 {
			m.OutputBuffer += " "
		}
	}
	m.OutputBuffer += "\n"
	return otto.Value{}
}

func (m *Machine) vmSubscribe(call otto.FunctionCall) otto.Value {
	action, err := call.Argument(0).ToString()
	m.vmThrowIf(err)
	device, err := call.Argument(1).ToString()
	m.vmThrowIf(err)
	handler := call.Argument(2)

	m.Subscribe(action, device, func(msg sarif.Message) {
		m.vmHandle(msg, handler)
	})
	return otto.Value{}
}

func (m *Machine) vmThrowIf(err error) {
	if err != nil {
		panic(m.VM.MakeTypeError(err.Error()))
	}
}

func (m *Machine) vmPublish(call otto.FunctionCall) otto.Value {
	msg := objectToMessage(call.Argument(0))

	if msg.Action != "" {
		m.Publish(msg)
	}
	return otto.Value{}
}

func (m *Machine) vmReply(call otto.FunctionCall) otto.Value {
	msg := objectToMessage(call.Argument(0))
	reply := objectToMessage(call.Argument(1))

	m.Reply(msg, reply)
	return otto.Value{}
}

func (m *Machine) vmRequest(call otto.FunctionCall) otto.Value {
	msg := objectToMessage(call.Argument(0))
	handler := call.Argument(1)

	if msg.Action != "" {
		go func() {
			for reply := range m.Request(msg) {
				m.vmHandle(reply, handler)
			}
		}()
	}
	return otto.Value{}
}

func (m *Machine) vmHandle(msg sarif.Message, handler otto.Value) {
	m.StateLock.Lock()
	defer m.StateLock.Unlock()

	v := messageToObject(m.VM, msg)
	_, err := handler.Call(handler, v)
	if err != nil {
		m.InformListeners("err/internal", err.Error())
		m.Log("err", "handle err: "+err.Error())
	}

	for _, l := range strings.Split(m.FlushOut(), "\n") {
		if l != "" {
			m.Log("info", "print: "+l)
		}
	}
}

func (m *Machine) FlushOut() string {
	out := strings.TrimSpace(m.OutputBuffer)
	m.OutputBuffer = ""
	return out
}

func (m *Machine) Do(code string) (string, error, interface{}) {
	m.StateLock.Lock()
	defer m.StateLock.Unlock()

	m.FlushOut()
	ret, err := m.VM.Run(code)
	out := m.FlushOut()
	if err != nil {
		m.InformListeners("err/internal", err.Error())
		return out, err, nil
	}

	if ret.IsDefined() {
		exp, _ := ret.Export()
		return out, nil, exp
	}

	return out, nil, nil
}

func (m *Machine) vmDebug(call otto.FunctionCall) otto.Value {
	out := ""

	for i, arg := range call.ArgumentList {
		if i > 0 {
			out += " "
		}

		if arg.IsPrimitive() {
			out += arg.String()
		} else {
			v, _ := arg.Export()
			text, _ := json.MarshalIndent(v, "", "    ")
			out += string(text)
		}
	}
	m.InformListeners("js/debug", out)
	return otto.Value{}
}

func (m *Machine) Attach(listener string) {
	m.Listeners = append(m.Listeners, listener)
}

func (m *Machine) InformListeners(action, message string) {
	for _, l := range m.Listeners {
		m.Publish(sarif.Message{
			Action:      action,
			Destination: l,
			Text:        message,
		})
	}
}
