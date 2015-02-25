// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package proto

import (
	"fmt"
	"io"
	"strings"
)

type subTree struct {
	Topic   map[string]*subTree
	Writers map[Writer]struct{}
}

func newSubtree() *subTree {
	return &subTree{
		make(map[string]*subTree, 0),
		make(map[Writer]struct{}),
	}
}

// Subscribes a connection to a specific topic. The subscription applies for all
// subtopics, e.g. a subscription for "location/geofence" also matches
// "location/geofence/home/enter".
// A "+" wildcard matches all branches, thus "location/geofence/+/enter" matches
// "location/geofence/home/enter" and "location/geofence/work/enter", but not
// "location/geofence/work/enter".
func (t *subTree) Subscribe(topic []string, c Writer) {
	if _, ok := t.Writers[c]; ok {
		// Writerection already subscribed to some parent topic? Nothing to do
		return
	}

	if topic != nil && len(topic) > 0 {
		// Descend further down the topic tree, creating nodes along the way
		st, ok := t.Topic[topic[0]]
		if !ok {
			st = newSubtree()
			t.Topic[topic[0]] = st
		}
		st.Subscribe(topic[1:], c)
		return
	}

	// Target node reached? Unsubscribe from narrower subtopics and make new
	// subscription
	t.Unsubscribe(nil, c)
	t.Writers[c] = struct{}{}
}

// Unsubscribe removes the subscription to a topic and all its subtopics for a
// specific connection. Thus, unsubscribing from "location" would remove both
// subscriptions to "location" and "location/geofence". An empty topics removes
// all subscriptions.
func (t *subTree) Unsubscribe(topic []string, c Writer) {
	if topic != nil && len(topic) > 0 {
		// Descend further along the topic path
		if st, ok := t.Topic[topic[0]]; ok {
			st.Unsubscribe(topic[1:], c)
			if len(st.Writers) == 0 && len(st.Topic) == 0 {
				// Delete tree if it becomes empty
				delete(t.Topic, topic[0])
			}
		}
		return
	}

	// Target node reached? Unsubscribe from it and all child trees
	delete(t.Writers, c)
	for top, st := range t.Topic {
		st.Unsubscribe(nil, c)
		if len(st.Writers) == 0 && len(st.Topic) == 0 {
			// Delete tree if it becomes empty
			delete(t.Topic, top)
		}
	}
}

// Get returns a child tree by walking along the topic path.
func (t *subTree) Get(topic []string) *subTree {
	if len(topic) == 0 {
		return t
	}

	if st, ok := t.Topic[topic[0]]; ok {
		return st.Get(topic[1:])
	}
	return nil
}

// Walk visits all children of a subtopic and calls function f on them.
//
// For example, walking along "location/geofence" would also match
// "location/geofence/home", but not "location" or "location/update",
// A "+" wildcard matches all branches: e.g. "topic/+/something" matches both
// "topic/one/something" and "topic/two/something".
func (t *subTree) Walk(topic []string, depth int, f func(Writer)) {
	if topic != nil && len(topic) > 0 {
		// Descend further along the topic path
		if topic[0] == "+" {
			for _, st := range t.Topic {
				st.Walk(topic[1:], depth+1, f)
			}
		} else {
			if st, ok := t.Topic[topic[0]]; ok {
				st.Walk(topic[1:], depth+1, f)
			}
		}
		return
	}

	for c := range t.Writers {
		f(c)
	}
	for _, st := range t.Topic {
		st.Walk(topic, depth+1, f)
	}
}

// Call applies a function f to all children matching the topic along the descent.
// This is the same rule for delivering messages to subscribers.
//
// For example, "topic/subtopic" would match children registered to "", ""topic"
// and "topic/subtopic", but not "topic/subtopic/deeper".
func (t *subTree) Call(topic []string, f func(Writer)) {
	if topic != nil && len(topic) > 0 {
		// Descend further along the topic path
		if st, ok := t.Topic[topic[0]]; ok {
			st.Call(topic[1:], f)
		}
		if st, ok := t.Topic["+"]; ok {
			st.Call(topic[1:], f)
		}
	}

	for c := range t.Writers {
		f(c)
	}
}

func (t *subTree) Print(w io.Writer, depth int) error {
	if depth == 0 {
		if _, err := fmt.Fprintf(w, "/ (%d)\n", len(t.Writers)); err != nil {
			return err
		}
	}

	indent := strings.Repeat(" ", (depth+1)*4)
	for c := range t.Writers {
		if _, err := fmt.Fprintf(w, "%s- Writer: %s\n", indent, c); err != nil {
			return err
		}
	}
	for topic, st := range t.Topic {
		if _, err := fmt.Fprintf(w, "%s+ %s (%d)\n", indent, topic, len(st.Writers)); err != nil {
			return err
		}
		if err := st.Print(w, depth+1); err != nil {
			return err
		}
	}
	return nil
}

func (t *subTree) GetTopics(root string, topics []string) []string {
	for top, st := range t.Topic {
		top = root + "/" + top
		if len(st.Writers) > 0 {
			topics = append(topics, top)
		} else {
			topics = st.GetTopics(top, topics)
		}
	}
	return topics
}
