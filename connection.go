package irc

import (
	"bufio"
	"log"
	"net"
)

type Conn struct {
	conn               net.Conn
	Msg                <-chan Message
	Send               chan<- Message
	recvDone, sendDone <-chan struct{}
}

func Dial(address string) (c Conn, err error) {
	c.conn, err = net.Dial("tcp4", address)
	if err != nil {
		return
	}
	c.Msg, c.recvDone = recvLoop(c.conn)
	c.Send, c.sendDone = sendLoop(c.conn)
	return
}

func (c *Conn) Close() {
	close(c.Send)
	c.Msg = nil
	c.conn.Close()
	<-c.recvDone
	<-c.sendDone
	log.Print("cu")
}

func recvLoop(conn net.Conn) (<-chan Message, <-chan struct{}) {
	ch := make(chan Message, 10)
	done := make(chan struct{})

	go func() {
		defer func() {
			log.Print("recv loop close")
			done <- struct{}{}
		}()
		buf := bufio.NewReader(conn)
		for {
			b, _, err := buf.ReadLine()
			if err != nil {
				log.Print("ReadLine(): ", err)
				return
			}
			m, err := ParseMessage(b)
			if err != nil {
				log.Print("ParseMessage(): ", err)
				return
			}
			ch <- m
		}
	}()

	return ch, done
}

func sendLoop(conn net.Conn) (chan<- Message, <-chan struct{}) {
	ch := make(chan Message, 10)
	done := make(chan struct{})

	go func() {
		defer func() {
			log.Print("send loop close")
			done <- struct{}{}
		}()
		for {
			msg, open := <-ch
			switch {
			case open:
				log.Print("send: ", msg.String())
				conn.Write([]byte(msg.String()))
				continue
			case !open:
				return
			}
		}
	}()

	return ch, done
}
