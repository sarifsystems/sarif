// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package hostscan

import (
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

var re_nmap_hosts = regexp.MustCompile(`Host: ([\d\.]+) \(([\w\.\-]*)\)`)

type Host struct {
	Id     int64     `json:"-"`
	Ip     string    `json:"ip"`
	Name   string    `json:"name,omitempty" sql:"index"`
	Time   time.Time `json:"time" sql:"index"`
	Status string    `json:"status"`
}

func (Host) TableName() string {
	return "hostscan"
}

func (h Host) String() string {
	n := h.Name
	if n == "" {
		n = h.Ip
	}
	return fmt.Sprintf("%s is %s since %s.", n, h.Status, h.Time)
}

type HostScan struct {
	db              *gorm.DB
	MinDownInterval time.Duration
}

func New(db *gorm.DB) *HostScan {
	return &HostScan{db, 9 * time.Minute}
}

func (h *HostScan) Setup() error {
	return h.db.AutoMigrate(&Host{}).Error
}

func (h *HostScan) FindLocalNetwork() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		a := addr.String()
		if strings.HasPrefix(a, "192.168.") {
			return a, nil
		}
	}
	return "", nil
}

func (h *HostScan) ScanCurrentHosts() ([]Host, error) {
	addr, err := h.FindLocalNetwork()
	if err != nil {
		return nil, err
	}
	cmd := exec.Command("nmap", "-T3", "-sn", "-oG", "-", addr)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	matches := re_nmap_hosts.FindAllStringSubmatch(string(out), -1)

	now := time.Now()
	hosts := make([]Host, len(matches))
	for i, match := range matches {
		hosts[i].Ip = match[1]
		hosts[i].Name = match[2]
		hosts[i].Status = "up"
		hosts[i].Time = now
	}

	return hosts, nil
}

func (h *HostScan) InsertHosts(hosts []Host) error {
	tx := h.db.Begin()
	for _, h := range hosts {
		if err := tx.Save(&h).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}

func (h *HostScan) Update() ([]Host, error) {
	curr, err := h.ScanCurrentHosts()
	if err != nil {
		return nil, err
	}
	last, err := h.LastStatusAll()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	changed := make([]Host, 0)
	for _, host := range last {
		if now.Sub(host.Time) < h.MinDownInterval {
			continue
		}
		if host.Status == "up" && !HostInSlice(host, curr) {
			host.Status = "down"
			host.Time = time.Now()
			changed = append(changed, host)
		}
	}
	for _, host := range curr {
		if !HostInSlice(host, last) {
			host.Status = "up"
			host.Time = time.Now()
			changed = append(changed, host)
		}
	}

	if err := h.InsertHosts(changed); err != nil {
		return changed, err
	}

	return changed, nil
}

func (h *HostScan) LastStatus(name string) (Host, error) {
	var host Host
	host.Name = name
	err := h.db.Where(host).Order("time DESC").First(&host).Error
	return host, err
}

func (h *HostScan) LastStatusAll() ([]Host, error) {
	after := time.Now().AddDate(0, -1, 0)
	ids := make([]int64, 0)
	rows, err := h.db.Model(Host{}).Select("MAX(id)").Where("time > ?", after).Group("ip").Rows()
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id int64
		rows.Scan(&id)
		ids = append(ids, id)
	}
	hosts := make([]Host, 0)
	if len(ids) == 0 {
		return hosts, nil
	}
	err = h.db.Where("id IN (?)", ids).Find(&hosts).Error
	return hosts, err
}

func HostInSlice(host Host, hosts []Host) bool {
	for _, curr := range hosts {
		if curr.Ip == host.Ip && curr.Name == host.Name && curr.Status == "up" {
			return true
		}
	}
	return false
}
