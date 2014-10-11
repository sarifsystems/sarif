// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package core

import "github.com/xconstruct/stark/services"

type ErrModuleNotFound struct {
	Module string
}

func (e ErrModuleNotFound) Error() string {
	return "module '" + e.Module + "' not found'"
}

var (
	modules = make(map[string]*services.Module)
)

func (app *App) RegisterModule(mod *services.Module) {
	app.modules[mod.Name] = mod
}

func (app *App) GetModule(name string) (*services.Module, error) {
	m, ok := app.modules[name]
	if !ok {
		return m, ErrModuleNotFound{name}
	}
	return m, nil
}

type enabler interface {
	Enable() error
}

type disabler interface {
	Disable() error
}
