// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package hostscan

import (
	"database/sql"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

const schema = `
CREATE TABLE IF NOT EXISTS hostscan (
	id int(10) NOT NULL AUTO_INCREMENT,
	ip varchar(16) NOT NULL,
	hostname varchar(100) NOT NULL,
	timestamp timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
	status enum('','up','down') NOT NULL,
	PRIMARY KEY (id),
	KEY timestamp (timestamp),
	KEY hostname (hostname)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
`

const schemaSqlite3 = `
CREATE TABLE IF NOT EXISTS hostscan (
	id INTEGER PRIMARY KEY,
	ip VARCHAR(16) NOT NULL,
	hostname VARCHAR(100) NOT NULL,
	timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	status VARCHAR(100) NOT NULL
);
CREATE INDEX IF NOT EXISTS timestamp ON hostscan (timestamp);
CREATE INDEX IF NOT EXISTS hostname ON hostscan (hostname);
`

var re_nmap_hosts = regexp.MustCompile(`Host: ([\d\.]+) \(([\w\.\-]*)\)`)

type Host struct {
	Ip        string    `json:"ip"`
	Name      string    `json:"name,omitempty"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"time"`
}

func (h Host) String() string {
	n := h.Name
	if n == "" {
		n = h.Ip
	}
	return fmt.Sprintf("%s is %s since %s.", n, h.Status, h.Timestamp)
}

type HostScan struct {
	db *sql.DB
}

func New(db *sql.DB) *HostScan {
	return &HostScan{db}
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
		hosts[i].Timestamp = now
	}

	return hosts, nil
}

func (h *HostScan) InsertHosts(hosts []Host) error {
	tx, err := h.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`INSERT INTO hostscan (ip, hostname, status, timestamp) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	for _, host := range hosts {
		_, err := stmt.Exec(host.Ip, host.Name, host.Status, host.Timestamp)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
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

	changed := make([]Host, 0)
	for _, host := range last {
		if host.Status == "up" && !HostInSlice(host, curr) {
			host.Status = "down"
			host.Timestamp = time.Now()
			changed = append(changed, host)
		}
	}
	for _, host := range curr {
		if !HostInSlice(host, last) {
			host.Status = "up"
			host.Timestamp = time.Now()
			changed = append(changed, host)
		}
	}

	if err := h.InsertHosts(changed); err != nil {
		return changed, err
	}

	return changed, nil
}

func (h *HostScan) LastStatus(name string) (Host, error) {
	row := h.db.QueryRow(`
		SELECT ip, hostname, timestamp, status FROM hostscan 
		WHERE ip = ? OR hostname = ?
		ORDER BY id DESC LIMIT 1
	`, name, name)
	var host Host
	err := row.Scan(&host.Ip, &host.Name, &host.Timestamp, &host.Status)
	return host, err

}

func (h *HostScan) LastStatusAll() ([]Host, error) {
	rows, err := h.db.Query(`
		SELECT d1.ip, d1.hostname, d1.status, d1.timestamp FROM hostscan d1
		JOIN (SELECT MAX(id) id FROM hostscan GROUP BY ip) d2
			ON d1.id = d2.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hosts := make([]Host, 0)
	for rows.Next() {
		var host Host
		rows.Scan(&host.Ip, &host.Name, &host.Status, &host.Timestamp)
		hosts = append(hosts, host)
	}
	return hosts, rows.Err()
}

func HostInSlice(host Host, hosts []Host) bool {
	for _, curr := range hosts {
		if curr.Ip == host.Ip && curr.Name == host.Name && curr.Status == "up" {
			return true
		}
	}
	return false
}

func SetupSchema(driver string, db *sql.DB) error {
	var err error
	if driver == "sqlite3" {
		_, err = db.Exec(schemaSqlite3)
	} else {
		_, err = db.Exec(schema)
	}
	return err
}
