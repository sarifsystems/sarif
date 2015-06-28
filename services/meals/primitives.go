// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package meals

import (
	"fmt"
	"strconv"
	"strings"
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

func (w *Weight) UnmarshalJSON(item []byte) error {
	s := strings.Trim(string(item), `"`)
	n, err := strconv.ParseFloat(s, 64)
	*w = Weight(n)
	return err
}

func (w *Volume) UnmarshalJSON(item []byte) error {
	s := strings.Trim(string(item), `"`)
	n, err := strconv.ParseFloat(s, 64)
	*w = Volume(n)
	return err
}

func (w *Energy) UnmarshalJSON(item []byte) error {
	s := strings.Trim(string(item), `"`)
	n, err := strconv.ParseFloat(s, 64)
	*w = Energy(n)
	return err
}

func convertToFloat64(src interface{}) (float64, error) {
	switch v := src.(type) {
	case float64:
		return v, nil
	case []uint8:
		return strconv.ParseFloat(string(v), 64)
	}
	return 0, fmt.Errorf("cannot convert %v (%T) to float64", src, src)
}

func (w *Weight) Scan(src interface{}) error {
	v, err := convertToFloat64(src)
	if err == nil {
		*w = Weight(v)
	}
	return err
}

func (w *Volume) Scan(src interface{}) error {
	v, err := convertToFloat64(src)
	if err == nil {
		*w = Volume(v)
	}
	return err
}

func (w *Energy) Scan(src interface{}) error {
	v, err := convertToFloat64(src)
	if err == nil {
		*w = Energy(v)
	}
	return err
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
