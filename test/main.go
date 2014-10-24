package main

import (
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/juggle-tux/irc"
)

func init() {
	log.SetPrefix("irc: ")
	log.SetFlags(log.Lshortfile)
}

func main() {
	conn, err := irc.Dial("irc.freenode.org:6667")
	if err != nil {
		log.Fatal(err)
	}
	//defer conn.Close()
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		for {
			log.Printf("%#s", <-conn.Msg)
		}
	}()
	user := irc.Message{
		Command:  "USER",
		Parms:    []string{"go-test", "0", "*"},
		Trailing: "go-test",
	}
	nick := irc.Message{
		Command: "NICK",
		Parms:   []string{"gotest"},
	}
	time.Sleep(time.Second)
	conn.Send <- user
	time.Sleep(time.Second)
	conn.Send <- nick
	<-c
	conn.Close()
}
