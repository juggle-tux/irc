package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/juggle-tux/irc"
)

// flags
var (
	clServer = flag.String("s", "irc.freenode.org:6667", "irc server")
	clNick   = flag.String("n", "", "nickname")
	clUser   = flag.String("u", "", "username if emty use same as nickname")
	clHelp   = flag.Bool("h", false, "show this")
)

type myResponse struct{}

func (myResponse) ServeIRC(req irc.Message, res chan<- irc.Message) bool {
	switch req.Command {
	case "PRIVMSG":
		fmt.Println("<", req.Parms[0], "|", req.Prefix.Nick, ">", req.Trailing)
		switch req.Parms[0] {
		// Privat message
		case *clNick:
			if req.Trailing == "!hello" {
				msg := irc.Message{
					Command:  "PRIVMSG",
					Parms:    irc.Parms{req.Prefix.Nick},
					Trailing: "hello " + req.Prefix.Nick,
				}
				fmt.Println("-=>", msg.Parms[0], "<=-", msg.Trailing)
				res <- msg
			}
		// Channel
		default:
			if strings.HasPrefix(req.Trailing, "!hello") {
				msg := irc.Message{
					Command:  "PRIVMSG",
					Parms:    req.Parms,
					Trailing: "hello " + req.Prefix.Nick,
				}
				fmt.Println("-=>", msg.Parms[0], "<=-", msg.Trailing)
				res <- msg
			}
		}
	}
	return false
}

var response = new(myResponse)

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

	//time.Sleep(time.Minute)
	conn.AutoResponse(response)

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, os.Kill)

	for {
		select {
		case <-conn.Msg:
			continue
		case <-quit:
			return
		}
	}
}
