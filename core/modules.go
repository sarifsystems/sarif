package core

type ErrModuleNotFound struct {
	Module string
}

func (e ErrModuleNotFound) Error() string {
	return "module '" + e.Module + "' not found'"
}

type ModuleInstance interface {
	Enable() error
	Disable() error
}

type Module struct {
	Name        string
	Version     string
	NewInstance func(ctx *Context) (ModuleInstance, error)
}

var (
	modules = make(map[string]Module)
)

func RegisterModule(m Module) {
	modules[m.Name] = m
}

func GetModule(name string) (Module, error) {
	m, ok := modules[name]
	if !ok {
		return m, ErrModuleNotFound{name}
	}
	return m, nil
}
