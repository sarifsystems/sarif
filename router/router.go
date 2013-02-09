package router

import (
	"log"
	"github.com/xconstruct/stark"
)

type Conn interface {
	Read() (*stark.Message, error)
	Write(*stark.Message) error
}

type LoggedConn struct {
	Conn
}

func (c *LoggedConn) Write(msg *stark.Message) error {
	log.Printf(" --> %v\n", msg)
	return c.Conn.Write(msg)
}

func (c *LoggedConn) Read() (*stark.Message, error) {
	msg, err := c.Conn.Read()
	log.Printf(" <-- %v\n", msg)
	return msg, err
}

type Router struct {
	Name string
	Route map[string]Conn
}

func NewRouter(name string) *Router {
	return &Router{
		name,
		make(map[string]Conn),
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

	next := path.Next()
	if next == "" {
		// TODO: Capabilities routing
		return nil
	}
	if r.Route[next] != nil {
		if err := r.Route[next].Write(msg); err != nil {
			return err
		}
		return nil
	}

	return &ErrDestination{next}
}

func (r *Router) Connect(conn Conn) {
	log.Printf("router/connect\n")
	go func() {
		name := stark.GenerateUUID()
		for {
			msg, err := conn.Read()
			if err != nil {
				delete(r.Route, name)
				log.Printf("router/disconnect: %v\n", err)
				return
			}
			if msg.Action == "route.hello" {
				newName, _ := msg.Data["name"].(string)
				log.Printf("router/hello: %s now known as %s\n", name, newName)
				delete(r.Route, name)
				name = newName
				r.Route[name] = conn
			}
			if err = r.Write(msg); err != nil {
				log.Printf("router: %v\n", err)
			}
		}
	}()
}
