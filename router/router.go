package router

import (
	"log"
	"github.com/xconstruct/stark"
)

type connInfo struct {
	conn stark.Conn
	dest string
	actions []string
}

type Router struct {
	Name string
	Conns map[stark.Conn]connInfo
}

func NewRouter(name string) *Router {
	return &Router{
		name,
		make(map[stark.Conn]connInfo, 0),
	}
}

type ErrDestination struct {
	Dest string
}

func (e *ErrDestination) Error() string {
	return "destination not found: " + e.Dest
}

func (r *Router) Write(msg *stark.Message) error {
	path := stark.GetPath(msg)
	log.Println(msg)

	// Exact destination found
	next := path.Next()
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

func (r *Router) Connect(conn stark.Conn) {
	go func() {
		for {
			msg, err := conn.Read()
			if err != nil {
				log.Printf("router/disconnect: %v\n", err)
				delete(r.Conns, conn)
				return
			}
			if msg.Action == "route.hello" {
				name, _ := msg.Data["name"].(string)
				actions, _ := msg.Data["actions"].([]string)
				log.Printf("router/connect: %s connected\n", name)
				r.Conns[conn] = connInfo{conn, name, actions}
				continue
			}
			if err = r.Write(msg); err != nil {
				log.Printf("router: %v\n", err)
			}
		}
	}()
}
