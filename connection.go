package irc

import (
	"bufio"
	"io"
	"log"
	"net"
	"time"
)

type Conn struct {
	conn     net.Conn
	Msg      chan Message
	Send     chan<- Message
	recvDone chan struct{}
	sendDone <-chan struct{}
	resFunc  chan ResponseFunc
}

func Dial(address, nick, user string) (c Conn, err error) {
	c.resFunc = make(chan ResponseFunc)
	c.conn, err = net.Dial("tcp4", address)
	if err != nil {
		return
	}
	c.resFunc = make(chan ResponseFunc)
	c.recvLoop()
	c.Send, c.sendDone = sendLoop(c.conn)

	// login
	go func() {
		c.Send <- Message{
			Command:  "USER",
			Parms:    Parms{user, "0", "*"},
			Trailing: user,
		}
		c.Send <- Message{
			Command: "NICK",
			Parms:   Parms{nick},
		}
	}()
	for m := range c.Msg {
		if m.Command == "376" {
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

// ResponseFunc will be called by AutoResponse
// if the return of the func is true the Message will be still send to the Conn.Msg chan
type ResponseFunc func(Message, chan<- Message) bool

func (c *Conn) AutoResponse(re ResponseFunc) {
	c.resFunc <- re
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

func (c *Conn) recvLoop() {
	c.Msg = make(chan Message, 10)
	c.recvDone = make(chan struct{})
	resFunc := ResponseFunc(func(Message, chan<- Message) bool {
		return true
	})

	go func() {
		defer func() {
			log.Print("recv loop close")
			c.recvDone <- struct{}{}
		}()
		buf := bufio.NewReader(c.conn)
		for {
			b, _, err := buf.ReadLine()
			switch err {
			case nil:
				break
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

			select {
			case resFunc = <-c.resFunc:
			default:
			}

			switch {
			case m.Command == "PING":
				c.Send <- Message{Command: "PONG", Trailing: m.Trailing}
			case resFunc(m, c.Send):
				c.Msg <- m

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
		for {
			msg, open := <-ch
			switch {
			case open:
				log.Print("send: ", msg.String())
				<-ticker
				conn.Write([]byte(msg.String()))
				continue
			case !open:
				return
			}
		}
	}()

	return ch, done
}
