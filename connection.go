package irc

import (
	"bufio"
	"io"
	"log"
	"net"
	"time"
)

type Handler interface {
	ServeIRC(req Message, res chan<- Message) (skip bool)
}

type defaultHandler struct{}

func (defaultHandler) ServeIRC(Message, chan<- Message) bool {
	return false
}

type Conn struct {
	conn       net.Conn
	address    string
	Msg        chan Message
	Send       chan Message
	Done       chan struct{}
	sendDone   chan struct{}
	resHandler chan Handler
	handlerMap map[string]map[string]Handler
}

type handlerMap map[string]map[string]*Handler

func Dial(address, nick, user string) (*Conn, error) {
	var err error
	var c = &Conn{
		address:    address,
		resHandler: make(chan Handler, 1),
		Done:       make(chan struct{}),
		sendDone:   make(chan struct{}),
	}
	if err = c.connect(); err != nil {
		return nil, err
	}
	c.recvLoop()
	c.login(nick, user)
	return c, nil

}

func (c *Conn) connect() error {
	log.Print("connecting to ", c.address)
	var err error

	c.conn, err = net.Dial("tcp4", c.address)
	if err != nil {
		return err
	}

	c.sendLoop()
	return nil
}

func (c *Conn) reconnect() error {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return err
		}
	}
	return c.connect()
}

func (c *Conn) login(nick, user string) {
	c.Send <- Message{
		Command:  "USER",
		Parms:    Parms{user, "0", "*"},
		Trailing: user,
	}
	c.Send <- Message{
		Command: "NICK",
		Parms:   Parms{nick},
	}

	for m := range c.Msg {
		if m.Command == RPL_ENDOFMOTD {
			log.Print(m)
			break
		}
	}

}

func (c *Conn) Join(channel string) {
	log.Print("Join ", channel)
	c.Send <- Message{
		Command: "JOIN",
		Parms:   Parms{channel},
	}
}

func (c Conn) AutoResponse(h Handler) {
	c.resHandler <- h
}

func (c *Conn) Quit() {
	if c.Send != nil {
		log.Print("send QUIT message")
		c.Send <- Message{
			Command:  "QUIT",
			Trailing: "watch this!",
		}
	}
}

func (c *Conn) Close() {
	if c.Send != nil {
		c.Quit()
	}
	if _, open := <-c.Done; open {
		close(c.Done)
	}
	c.conn.Close()
}

func (c *Conn) recvLoop() {
	log.Print("recvLoop start")
	c.Msg = make(chan Message, 10)
	var resHandler Handler = &defaultHandler{}

	go func() {
		defer func() {
			if c.Send != nil {
				close(c.Send)
				<-c.sendDone
			}
			c.conn.Close()
			c.Done <- struct{}{}
			log.Print("recvLoop close")
		}()

		buf := bufio.NewReader(c.conn)
		for {
			c.conn.SetDeadline(time.Now().Add(300 * time.Second))
			b, _, err := buf.ReadLine()
			switch err.(type) {
			case nil:
			case net.Error:
				if err.(net.Error).Timeout() {
					close(c.Send)
					<-c.sendDone
					if err = c.reconnect(); err == nil {
						continue
					}
					log.Print("recvLoop: ", err)
					return
				}

			default:
				if err == io.EOF {
					return
				}
				log.Print("recvLoop :", err)
				return
			}

			m, err := ParseMessage(b)
			if err != nil {
				log.Printf("recvLoop: %s\nraw: %s", err, b)
				continue
			}

			select {
			case resHandler = <-c.resHandler:
			default:
			}

			switch {
			case m.Command == "PING":
				c.Send <- Message{Command: "PONG", Trailing: m.Trailing}
			case !resHandler.ServeIRC(m, c.Send):
				c.Msg <- m

			}
		}
	}()

	return
}

func (c *Conn) sendLoop() {
	log.Print("sendLoop start")
	c.Send = make(chan Message, 10)
	go func() {
		defer func() {
			c.Send = nil
			c.sendDone <- struct{}{}
			log.Print("sendLoop close")
		}()
		ticker := time.Tick(500 * time.Millisecond)
		for m := range c.Send {
			<-ticker
			if _, err := c.conn.Write([]byte(m.String())); err != nil {
				log.Print("sendLoop: ", err)
				return
			}
		}
	}()

	return
}
