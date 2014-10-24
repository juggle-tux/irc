package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
//	"time"

	"github.com/juggle-tux/irc"
)

// flags
var (
	clServer = flag.String("s", "irc.freenode.org:6667", "irc server")
	clNick = flag.String("n", "", "nickname")
	clUser = flag.String("u", "", "username if emty use same as nickname")
	clHelp = flag.Bool("h", false, "show this")
)
var response = irc.ResponseFunc(func(req irc.Message, res chan<- irc.Message) bool {
	if req.Command == "PRIVMSG" && req.Trailing == "!hello" {
		res <- irc.Message{
			Command: "PRIVMSG",
			Parms: req.Parms,
			Trailing: "hello " + req.Prefix.Nick,
		}
	}
	return true
})

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
		case <- quit:
			return
		case m := <- conn.Msg:
			fmt.Print(m)
		}
	}
}