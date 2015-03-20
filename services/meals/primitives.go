// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package meals

import (
	"errors"
	"fmt"
)

type Weight float64
type Volume float64
type Energy float64

const (
	Milligram Weight = 1.0 / 1000
	Gram      Weight = 1000 * Milligram
	Kilogram         = 1000 * Gram

	Joule     Energy = 1.0 / 1000
	Kilojoule        = 1000 * Joule
	Kcal             = 4184 * Joule

	Millilitre Volume = 1.0
	Litre      Volume = 1000 * Millilitre
)

func (w *Weight) Scan(src interface{}) error {
	if v, ok := src.(float64); ok {
		*w = Weight(v)
		return nil
	}
	return errors.New("weight expected float64 as source type")
}

func (w *Volume) Scan(src interface{}) error {
	if v, ok := src.(float64); ok {
		*w = Volume(v)
		return nil
	}
	return errors.New("volume expected float64 as source type")
}

func (w *Energy) Scan(src interface{}) error {
	if v, ok := src.(float64); ok {
		*w = Energy(v)
		return nil
	}
	return errors.New("energy expected float64 as source type")
}

func (v Weight) String() string {
	if v >= Kilogram {
		return fmt.Sprintf("%.5g kg", v/Kilogram)
	}
	return fmt.Sprintf("%.5g g", v)
}

func (v Volume) String() string {
	if v >= Litre {
		return fmt.Sprintf("%.5g l", v/Litre)
	}
	return fmt.Sprintf("%.5g ml", v)
}

func (v Energy) String() string {
	return fmt.Sprintf("%.5g kJ", v)
}

func (v Energy) StringKcal() string {
	return fmt.Sprintf("%.5g kcal", v/Kcal)
}
