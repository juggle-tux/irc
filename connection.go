package irc

import (
	"bufio"
	"io"
	"log"
	"net"
	"time"
)

var defaultHandler HandlerFunc = func(Message, chan<- Message) bool {
	return false
}

type Handler interface {
	ServeIRC(req Message, res chan<- Message) (skip bool)
}

type HandlerFunc func(Message, chan<- Message) bool

func (f HandlerFunc) ServeIRC(req Message, res chan<- Message) bool {
	return f(req, res)
}

type Client struct {
	conn       net.Conn
	address    string
	nick, user string

	Msg        chan Message
	send       chan Message
	Done       chan struct{}
	sendDone   chan struct{}
	resHandler chan Handler
}

func Dial(address, nick, user string) (*Client, error) {
	var c = &Client{
		address:    address,
		nick:       nick,
		user:       user,
		resHandler: make(chan Handler, 1),
		Done:       make(chan struct{}),
		sendDone:   make(chan struct{}),
	}
	if err := c.connect(); err != nil {
		return nil, err
	}

	return c, nil

}

func (c *Client) Close() {
	c.Quit()
	if _, open := <-c.Done; open {
		close(c.Done)
	}
}

func (c Client) Handle(h Handler) {
	c.resHandler <- h
}

func (c Client) HandleFunc(f func(Message, chan<- Message) bool) {
	c.Handle(HandlerFunc(f))
}

func (c *Client) Send(m Message) error {
	c.send <- m
	return nil
}

func (c *Client) SendChan() chan<- Message {
	return c.send
}

func (c *Client) Quit() {
	if c.send != nil {
		log.Print("send QUIT message")
		c.send <- Message{
			Command:  "QUIT",
			Trailing: "watch this!",
		}
	}
}

func (c *Client) connect() error {
	log.Print("connecting to ", c.address)
	var err error
	if c.conn, err = net.Dial("tcp4", c.address); err != nil {
		return err
	}

	c.sendLoop()
	c.send <- Message{
		Command:  "USER",
		Parms:    Parms{c.user, "0", "*"},
		Trailing: c.user,
	}
	c.send <- Message{
		Command: "NICK",
		Parms:   Parms{c.nick},
	}
	c.recvLoop()
	for m := range c.Msg {
		if m.Command == RPL_ENDOFMOTD {
			log.Print(m)
			break
		}
	}
	return nil
}

func (c *Client) reconnect() error {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return err
		}
	}
	for err := c.connect(); err != nil; err = c.connect() {
		log.Println(err)
	}
	return nil
}

func (c *Client) recvLoop() {
	log.Print("recvLoop start")
	c.Msg = make(chan Message, 10)

	go func() {
		defer func() {
			if c.send != nil {
				close(c.send)
				<-c.sendDone
			}
			if c.conn != nil {
				c.conn.Close()
			}
			close(c.Msg)
			c.Done <- struct{}{}
			log.Print("recvLoop close")
		}()

		resHandler := Handler(defaultHandler)
		buf := bufio.NewReader(c.conn)
		for {
			c.conn.SetDeadline(time.Now().Add(300 * time.Second))
			b, _, err := buf.ReadLine()
			switch err {
			case nil:
			case io.EOF:
				return
			default:
				close(c.send)
				<-c.sendDone
				if err := c.reconnect(); err != nil {
					log.Print(err)
					return
				}

			}

			m, err := ParseMessage(b)
			if err != nil {
				log.Printf("recvLoop: %s\nraw: %#nv", err, b)
				continue
			}

			select {
			case resHandler = <-c.resHandler:
			default:
			}

			switch {
			case m.Command == "PING":
				c.send <- Message{Command: "PONG", Trailing: m.Trailing}
			case !resHandler.ServeIRC(m, c.send):
				c.Msg <- m

			}
		}
	}()

	return
}

func (c *Client) sendLoop() {
	log.Print("sendLoop start")
	c.send = make(chan Message, 10)
	go func() {
		defer func() {
			c.send = nil
			c.sendDone <- struct{}{}
			log.Print("sendLoop close")
		}()
		ticker := time.Tick(500 * time.Millisecond)
		for m := range c.send {
			<-ticker
			if _, err := c.conn.Write([]byte(m.String())); err != nil {
				log.Print("sendLoop: ", err)
				return
			}
		}
	}()

	return
}
