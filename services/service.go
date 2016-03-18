// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package services

type Module struct {
	Name        string
	Version     string
	Description string

	NewInstance interface{}
}

type Config interface {
	Exists() bool
	Set(v interface{}) error
	Get(v interface{}) (error, bool)

	Dir() string
}

type ModuleManager struct {
	Instantiate func(*Module) (interface{}, error)
	Modules     map[string]*Module
	Instances   map[string]interface{}
}

func NewModuleManager(instantiate func(*Module) (interface{}, error)) *ModuleManager {
	return &ModuleManager{
		instantiate,
		make(map[string]*Module),
		make(map[string]interface{}),
	}
}

func (s *ModuleManager) EnableModule(name string) error {
	i, ok := s.Instances[name]
	if ok {
		return nil
	}

	m, err := s.GetModule(name)
	if err != nil {
		return err
	}

	i, err = s.Instantiate(m)
	if err != nil {
		return err
	}
	s.Instances[name] = i

	if i, ok := i.(enabler); ok {
		return i.Enable()
	}
	return nil
}

func (s *ModuleManager) DisableModule(name string) error {
	i, ok := s.Instances[name]
	if !ok {
		return nil
	}
	s.Instances[name] = nil
	if i, ok := i.(disabler); ok {
		if err := i.Disable(); err != nil {
			return err
		}
	}
	delete(s.Instances, name)
	return nil
}

type ErrModuleNotFound struct {
	Module string
}

func (e ErrModuleNotFound) Error() string {
	return "module '" + e.Module + "' not found'"
}

func (s *ModuleManager) RegisterModule(mod *Module) {
	s.Modules[mod.Name] = mod
}

func (s *ModuleManager) GetModule(name string) (*Module, error) {
	m, ok := s.Modules[name]
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
