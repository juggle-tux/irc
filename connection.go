package irc

import "os"

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
	Msg        chan Message
	Send       chan<- Message
	recvDone   chan struct{}
	sendDone   <-chan struct{}
	resHandler chan Handler
	handlerMap map[string]map[string]Handler
}

type handlerMap map[string]map[string]*Handler

func Dial(address, nick, user string) (c Conn, err error) {
	c.resHandler = make(chan Handler)
	c.handlerMap = make(map[string]map[string]Handler)

	log.Print("connecting to ", address)
	c.conn, err = net.Dial("tcp4", address)
	if err != nil {
		return
	}
	c.recvLoop()
	c.Send, c.sendDone = sendLoop(c.conn)

	// login
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
	return
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

func (c *Conn) Close() {
	c.Send <- Message{
		Command:  "QUIT",
		Trailing: "watch this!",
	}
	<-c.recvDone
	close(c.Send)
	<-c.sendDone
	c.conn.Close()
}

type tHandler struct{}

func (*tHandler) ServeIRC(Message, chan<- Message) bool {
	log.Print("test")
	return false
}

func (c *Conn) recvLoop() {
	c.Msg = make(chan Message, 10)
	c.recvDone = make(chan struct{})
	var resHandler Handler = &defaultHandler{}

	go func() {
		defer func() {
			log.Print("recv loop close")
			c.recvDone <- struct{}{}
		}()

		// raw dump
		lf, err := os.Create(c.conn.RemoteAddr().String() + ".raw")
		if err != nil {
			log.Fatal(err)
		}
		defer lf.Close()
		log.Print("logfile ", c.conn.RemoteAddr().String()+".raw")
		rlog := bufio.NewWriter(lf)

		tH := new(tHandler)
		c.handlerMap["PING"][""] = tH
		buf := bufio.NewReader(c.conn)
		for {
			b, _, err := buf.ReadLine()
			switch err {
			case nil:
			case io.EOF:
				return
			default:
				log.Print("ReadLine() ", err)
				continue
			}

			m, err := ParseMessage(b)
			if err != nil {
				log.Printf("ParseMessage(): %s\nraw: %s", err, b)
				continue
			}
			rlog.WriteString(m.String())
			rlog.Flush()

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
			if rh, ok := c.handlerMap[m.Command][m.Parms[0]]; ok {
				rh.ServeIRC(m, c.Send)
			}
		}
	}()

	return
}

func sendLoop(conn net.Conn) (chan<- Message, <-chan struct{}) {
	ch := make(chan Message, 10)
	done := make(chan struct{})

	go func() {
		defer func() {
			log.Print("send loop close")
			done <- struct{}{}
		}()
		ticker := time.Tick(500 * time.Millisecond)
		for m := range ch {
			<-ticker
			conn.Write([]byte(m.String()))
		}
	}()

	return ch, done
}
