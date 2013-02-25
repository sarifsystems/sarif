// Package router provides a service that listens on multiple
// connections and forwards messages based on simple rules.
package router

import (
	"log"
	"github.com/xconstruct/stark"
	"github.com/xconstruct/stark/transport"
)

type connInfo struct {
	conn transport.Conn
	dest string
	actions []string
}

// A router listens on multiple connections and forwards messages based
// on simple rules.
type Router struct {
	Name string
	Conns map[transport.Conn]connInfo
}

// Creates a new router with the desired name.
func NewRouter(name string) *Router {
	return &Router{
		name,
		make(map[transport.Conn]connInfo, 0),
	}
}

// ErrDestination is returned when the destination for a message was not found.
type ErrDestination struct {
	Dest string
}

func (e *ErrDestination) Error() string {
	return "destination not found: " + e.Dest
}

// Write sends the message to one or more connections based on the rules.
func (r *Router) Write(msg *stark.Message) error {
	log.Println(msg)

	route := stark.ParseRoute(msg.Source, msg.Destination)
	// Exact destination found
	next := route.Forward(r.Name)
	if next != "" {
		for _, info := range r.Conns {
			if info.dest != next {
				continue
			}
			if err := info.conn.Write(msg); err != nil {
				return err
			}
			return nil
		}
	}

	// Action-based routing
	for _, info := range r.Conns {
		if info.actions == nil {
			continue
		}
		for _, action := range info.actions {
			if action == msg.Action {
				if err := info.conn.Write(msg); err != nil {
					return err
				}
				return nil
			}
		}
	}

	return &ErrDestination{"unknown"}
}

func (r *Router) handle(conn transport.Conn, msg *stark.Message) error {
	switch (msg.Action) {
	case "route.hello":
		name, _ := msg.Data["name"].(string)
		actions, _ := msg.Data["actions"].([]string)
		log.Printf("router/connect: %s connected\n", name)
		r.Conns[conn] = connInfo{conn, name, actions}
	default:
		if err := r.Write(msg); err != nil {
			return err
		}
	}
	return nil
}

// Connect accepts a new connection and starts listening on it for incoming
// messages.
func (r *Router) Connect(conn transport.Conn) {
	go func() {
		msg := stark.NewMessage()
		msg.Action = "route.hello"
		msg.Data["name"] = r.Name
		msg.Data["actions"] = "route.route"
		msg.Message = "Hello from service " + r.Name
		err := conn.Write(msg)
		if err != nil {
			log.Printf("router/disconnect: %v\n", err)
			return
		}

		for {
			msg, err := conn.Read()
			if err != nil {
				log.Printf("router/disconnect: %v\n", err)
				delete(r.Conns, conn)
				return
			}
			r.handle(conn, msg)
		}
	}()
}

func (r *Router) Listen(listener transport.Listener) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		r.Connect(conn)
	}
	return nil
}
