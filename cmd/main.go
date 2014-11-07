package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/juggle-tux/irc"
)

// flags
var (
	clServer = flag.String("s", "irc.freenode.org:6667", "irc server")
	clNick   = flag.String("n", "", "nickname")
	clUser   = flag.String("u", "", "username if emty use same as nickname")
	clHelp   = flag.Bool("h", false, "show this")
)

var responseMap = map[string]func(*irc.Message) *irc.Message{
	"!hello": func(req *irc.Message) *irc.Message {
		return &irc.Message{
			Command:  "PRIVMSG",
			Parms:    req.Parms,
			Trailing: "hello " + req.Prefix.Nick,
		}
	},

	"!time": func(req *irc.Message) *irc.Message {
		return &irc.Message{
			Command:  "PRIVMSG",
			Parms:    req.Parms,
			Trailing: time.Now().String(),
		}
	},
	"!OP": func(req *irc.Message) *irc.Message {
		return &irc.Message{
			Command: "MODE",
			Parms: irc.Parms{
				0: req.Parms[0],
				1: "+o",
				2: req.Prefix.Nick,
			},
		}
	},
}

var ResHandler = func(req irc.Message, res chan<- irc.Message) bool {
	switch req.Command {
	case "MODE":
		log.Print(req)
	case "PRIVMSG":
		fmt.Println("<", req.Parms[0], "|", req.Prefix.Nick, ">", req.Trailing)
		switch req.Parms[0] {
		// Privat message
		case *clNick:
			msg := irc.Message{
				Command:  "PRIVMSG",
				Parms:    irc.Parms{0: req.Prefix.Nick},
				Trailing: "I'll not speak to you " + req.Prefix.Nick,
			}
			fmt.Println("-=>", msg.Parms[0], "<=-", msg.Trailing)
			res <- msg
		default:
			if mf, ok := responseMap[req.Trailing]; ok {
				msg := mf(&req)
				fmt.Println("-=>", msg.Parms[0], "<=-", msg.Trailing)
				res <- *msg
			}
		}
	case "JOIN":
		if req.Prefix.Nick == "JuggleTux" {
			res <- irc.Message{
				Command: "MODE",
				Parms: irc.Parms{
					0: req.Parms[0],
					1: "+o",
					2: req.Prefix.Nick,
				},
			}
		}
	}
	return false
}

func main() {
	flag.Parse()
	if *clHelp || *clNick == "" {
		flag.Usage()
		return
	}
	if *clUser == "" {
		*clUser = *clNick
	}

	conn, err := irc.Dial(*clServer, *clNick, *clUser)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	for _, ch := range flag.Args() {
		conn.Join(ch)
	}

	conn.HandleFunc(ResHandler)

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, os.Kill)

	go func() {
		for _, open := <-conn.Msg; open; _, open = <-conn.Msg {
		}
		quit <- os.Interrupt
	}()
	<-quit
}
