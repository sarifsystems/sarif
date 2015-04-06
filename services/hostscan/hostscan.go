// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package hostscan

import (
	"errors"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

var re_nmap_hosts = regexp.MustCompile(`Host: ([\d\.]+) \(([\w\.\-]*)\)`)

type Host struct {
	Ip        string    `json:"ip"`
	Name      string    `json:"name,omitempty"`
	Time      time.Time `json:"time"`
	UpdatedAt time.Time `json:"-"`
	Status    string    `json:"status"`
}

func (h Host) String() string {
	n := h.Name
	if n == "" {
		n = h.Ip
	}
	return fmt.Sprintf("%s is %s since %s.", n, h.Status, h.Time.Format(time.RFC3339))
}

type HostScan struct {
	MinDownInterval time.Duration
	status          map[string]Host
}

func New() *HostScan {
	return &HostScan{
		9 * time.Minute,
		make(map[string]Host),
	}
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
		hosts[i].UpdatedAt = now
	}

	return hosts, nil
}

func (h *HostScan) Update() ([]Host, error) {
	curr, err := h.ScanCurrentHosts()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	changed := make([]Host, 0)
	for _, host := range curr {
		last, ok := h.status[host.Ip]
		if !ok || last.Status == "down" {
			h.status[host.Ip] = host
			changed = append(changed, host)
		} else {
			last.UpdatedAt = now
			h.status[host.Ip] = last
		}
	}
	for _, last := range h.status {
		if last.Status == "up" && now.Sub(last.UpdatedAt) > h.MinDownInterval {
			last.Status = "down"
			last.Time = now.Add(-h.MinDownInterval / 2)
			last.UpdatedAt = now
			h.status[last.Ip] = last
			changed = append(changed, last)
		}
	}

	return changed, nil
}

func (h *HostScan) LastStatus(name string) (Host, error) {
	for _, host := range h.status {
		if host.Ip == name || host.Name == name || host.Name == name+".lan" {
			return host, nil
		}
	}
	return Host{}, errors.New("Host " + name + " not found")
}

func (h *HostScan) LastStatusAll() ([]Host, error) {
	after := time.Now().AddDate(0, -1, 0)

	hosts := make([]Host, 0)
	for _, host := range h.status {
		if host.Time.After(after) {
			hosts = append(hosts, host)
		}
	}
	return hosts, nil
}
