// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package location

import "time"

type ClusterStatus int

const (
	UnconfirmedCluster ClusterStatus = iota
	ConfirmedCluster
	CompletedCluster
)

type Cluster struct {
	Start  Location      `json:"start,omitempty"`
	End    Location      `json:"end,omitempty"`
	Status ClusterStatus `json:"status"`
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
		g.current.Status = UnconfirmedCluster
		g.current.Start = l
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
			Start:  l,
			Status: UnconfirmedCluster,
		}
		return changed
	}

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
