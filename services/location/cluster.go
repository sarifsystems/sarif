// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package location

import (
	"fmt"
	"time"
)

type ClusterStatus int

const (
	UnconfirmedCluster ClusterStatus = iota
	ConfirmedCluster
	CompletedCluster
)

type Cluster struct {
	Location
	Status ClusterStatus `json:"status"`

	Start Location `json:"start,omitempty"`
	End   Location `json:"end,omitempty"`
}

func (c Cluster) Text() string {
	if c.Address == "" {
		c.Address = fmt.Sprintf("%.4f, %.4f", c.Latitude, c.Longitude)
	}

	switch c.Status {
	case ConfirmedCluster:
		ts := c.Start.Timestamp.Local().Format(time.RFC3339)
		return "Entered " + c.Address + " on " + ts
	case CompletedCluster:
		ts := c.End.Timestamp.Local().Format(time.RFC3339)
		return "Left " + c.Address + " on " + ts
	}
	return ""
}

func NewClusterGenerator() *ClusterGenerator {
	return &ClusterGenerator{
		MinInterval: 30 * time.Minute,
		MaxDistance: 500,
	}
}

type ClusterGenerator struct {
	MinInterval time.Duration
	MaxDistance float64

	current  Cluster
	clusters []Cluster
}

func (g *ClusterGenerator) Advance(l Location) bool {
	if g.current.Start.Timestamp.IsZero() {
		g.current = Cluster{
			Start:    l,
			Location: l,
			Status:   UnconfirmedCluster,
		}
		return false
	}

	dist := HaversineDistance(g.current.Start, l)
	if dist > g.MaxDistance {
		changed := g.current.Status == ConfirmedCluster
		if changed {
			g.current.Status = CompletedCluster
			g.clusters = append(g.clusters, g.current)
		}

		g.current = Cluster{
			Start:    l,
			Location: l,
			Status:   UnconfirmedCluster,
		}
		return changed
	}

	w := l.Accuracy / (g.current.Accuracy + l.Accuracy + 0.00001)
	g.current.Longitude = (1-w)*l.Longitude + w*g.current.Longitude
	g.current.Latitude = (1-w)*l.Latitude + w*g.current.Latitude
	g.current.Accuracy = (1-w)*l.Accuracy + w*g.current.Accuracy

	g.current.End = l
	if g.current.Status == UnconfirmedCluster && l.Timestamp.Sub(g.current.Start.Timestamp) >= g.MinInterval {
		g.current.Status = ConfirmedCluster
		return true
	}

	return false
}

func (g *ClusterGenerator) Current() Cluster {
	return g.current
}

func (g *ClusterGenerator) ClearCompleted() {
	g.clusters = make([]Cluster, 0)
}

func (g *ClusterGenerator) Completed() []Cluster {
	if g.clusters == nil {
		g.ClearCompleted()
	}
	return g.clusters
}

func (g *ClusterGenerator) LastCompleted() Cluster {
	if len(g.clusters) == 0 {
		return Cluster{}
	}
	return g.clusters[len(g.clusters)-1]
}
