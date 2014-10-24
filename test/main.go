package main

import (
	//"fmt"
	//"time"
	"log"
	"os"
	"os/signal"

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
	defer conn.Close()
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		for {
			m := <-conn.Msg
			switch m.Command {
			case "PRIVMSG":
				if m.Parms[0] == "#go-bottest" && m.Trailing == "!hello" {
					conn.Send <- irc.Message{
						Command:  "PRIVMSG",
						Parms:    irc.Parms{"#go-bottest"},
						Trailing: "world",
					}
				}
				continue
			case "PING":
				conn.Send <- irc.Message{Command: "PONG", Trailing: m.Trailing}
			case "376":
				conn.Send <- irc.Message{Command: "JOIN", Parms: irc.Parms{"#go-bottest"}}
			default:
			}
			if m.Prefix.Nick != "" {
				log.Printf("%s |%s| %s:%s", m.Prefix.Nick, m.Command, m.Parms, m.Trailing)
			} else {
				log.Printf("%s |%s| %s:%s", m.Prefix, m.Command, m.Parms, m.Trailing)
			}
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
	conn.Send <- user
	conn.Send <- nick
	<-c
}
