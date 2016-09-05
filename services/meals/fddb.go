// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package meals

import (
	"time"

	"github.com/sarifsystems/sarif/pkg/fddb"
)

func (s *Service) FddbLoop() {
	for {
		time.Sleep(6 * time.Hour)

		prev := time.Now().Add(-24 * time.Hour)
		if err := s.FetchFddb(prev); err != nil {
			s.Log("err", "[meals] fddb update failed: "+err.Error())
		}
		if err := s.FetchFddb(time.Now()); err != nil {
			s.Log("err", "[meals] fddb update failed: "+err.Error())
		}
	}
}

func (s *Service) FetchFddb(day time.Time) error {
	c := fddb.New(s.cfg.FDDB.ApiKey)
	c.SetLoginInfo(s.cfg.FDDB.Username, s.cfg.FDDB.Password)

	r, err := c.DiaryGetDay(day)
	if err != nil {
		return err
	}

	for _, el := range r.DiaryElements {
		data := el.Item.Data
		if data.ServingAmount <= 0 {
			continue
		}

		p := &Product{
			RefId: int64(el.Item.Id),
			Name:  el.Item.Description.FullName(),

			ServingWeight: Weight(data.ServingAmount) * Gram,
			ServingVolume: 0,

			Stats: Stats{
				Weight: Weight(data.Amount) * Gram,
				Energy: Energy(data.KJ) * Kilojoule,

				Fat:           Weight(data.FatGram) * Gram,
				Carbohydrates: Weight(data.KhGram) * Gram,
				Sugar:         Weight(data.SugarGram) * Gram,
				Protein:       Weight(data.ProteinGram) * Gram,
			},
		}

		db := s.DB.Where(Product{RefId: p.RefId}).Assign(p)
		if err := db.FirstOrCreate(&p).Error; err != nil {
			return err
		}

		srv := &Serving{
			RefId:        int64(el.Uid),
			Name:         el.Item.Description.FullName(),
			AmountWeight: Weight(data.ServingAmount) * Gram,
			Time:         time.Unix(el.Date, 0),

			ProductId: p.Id,
			Product:   p,
		}
		db = s.DB.Where(Serving{RefId: srv.RefId}).Assign(srv)
		if err := db.FirstOrCreate(&srv).Error; err != nil {
			return err
		}
	}
	return nil
}
