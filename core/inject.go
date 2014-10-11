package core

import (
	"errors"
	"reflect"
)

type Injector struct {
	instances map[reflect.Type]reflect.Value
	factories map[reflect.Type]reflect.Value
}

func NewInjector() *Injector {
	return &Injector{
		make(map[reflect.Type]reflect.Value),
		make(map[reflect.Type]reflect.Value),
	}
}

func (in *Injector) Instance(instance interface{}) {
	in.instances[reflect.TypeOf(instance)] = reflect.ValueOf(instance)
}

func (in *Injector) Factory(factory interface{}) {
	v := reflect.ValueOf(factory)
	if v.Kind() != reflect.Func {
		panic("inject: expected factory func, not " + v.Type().String())
	}
	t := reflect.TypeOf(factory)
	if t.NumIn() != 0 {
		panic("inject: expected factory func with 0 parameters")
	}
	if t.NumOut() != 1 {
		panic("inject: expected factory with 1 result")
	}
	in.factories[t.Out(0)] = v
}

func (in *Injector) Inject(container interface{}) error {
	vcont := reflect.ValueOf(container)
	if vcont.Kind() != reflect.Ptr {
		return errors.New("inject: expected struct pointer, not " + vcont.Kind().String())
	}
	vcont = reflect.Indirect(vcont)
	if vcont.Kind() != reflect.Struct {
		return errors.New("inject: expected struct pointer, not " + vcont.Kind().String() + " pointer")
	}

	vtype := vcont.Type()
	for i := 0; i < vtype.NumField(); i++ {
		ftype := vtype.Field(i).Type
		if instance, ok := in.instances[ftype]; ok {
			vcont.Field(i).Set(instance)
			continue
		}
		if factory, ok := in.factories[ftype]; ok {
			vcont.Field(i).Set(factory.Call(nil)[0])
			continue
		}
		return errors.New("Could not populate '" + vtype.Field(i).Name + "': no " + vtype.Field(i).Type.String())
	}

	return nil
}

func (in *Injector) Create(factory interface{}) (interface{}, error) {
	v := reflect.ValueOf(factory)
	if v.Kind() != reflect.Func {
		return nil, errors.New("inject: expected factory func, not " + v.Type().String())
	}
	t := reflect.TypeOf(factory)
	if t.NumIn() != 1 {
		return nil, errors.New("inject: expected factory func with 1 parameters")
	}

	vin := reflect.New(t.In(0).Elem())
	if err := in.Inject(vin.Interface()); err != nil {
		return nil, err
	}
	out := v.Call([]reflect.Value{vin})
	return out[0].Interface(), nil
}
