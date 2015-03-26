package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strings"
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

var responseMap = map[string]func(*irc.Message) irc.Message{
	"!hello": func(req *irc.Message) irc.Message {
		return irc.Msg(req.Parms[0], "hello "+req.Prefix.Nick)
	},
	"!time": func(req *irc.Message) irc.Message {
		return irc.Msg(req.Parms[0], time.Now().UTC().String())
	},
	"!backdoor": func(req *irc.Message) irc.Message {
		return irc.Op(req.Prefix.Nick, req.Parms[0])
	},
}

func ResHandler(req irc.Message, res chan<- irc.Message) bool {
	switch req.Command {
	case "MODE":
		log.Print(req)
	case "PRIVMSG":
		str := strings.Replace(req.Trailing, "\x02", "", -1)
		str = strings.Replace(str, "\x03", "", -1)
		fmt.Printf("%q<->%q: %q\n", req.Parms[0], req.Prefix.Nick, str)
		switch req.Parms[0] {
		// Privat message
		case *clNick:
			res <- irc.Msg(req.Prefix.Nick, "I'll not speak to you "+req.Prefix.Nick)
		default:
			if strings.HasPrefix(req.Trailing, "!") { // if somebody says something what starts with a '!'
				if mf, ok := responseMap[req.Trailing]; ok { // if we have a '!...' function
					msg := mf(&req) // run it
					fmt.Println("-=>", msg.Parms[0], "<=-", msg.Trailing)
					res <- msg // send the answer
				}
			}
		}
	case "JOIN":
		if req.Prefix.Nick == "JuggleTux" {
			res <- irc.Op(req.Prefix.Nick, req.Parms[0])
		}
		log.Print(req)
	case "PART", "QUIT":
		log.Print(req)
	}
	return false
}

func main() {
	// Profile
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

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
	cm := irc.NewCM(conn)
	cm.DefaultHandler = irc.HandlerFunc(ResHandler)
	conn.Handle(cm)
	for _, ch := range flag.Args() {
		conn.Send(irc.Join(ch))
	}

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, os.Kill)

	go func() {
		for _, open := <-conn.Msg; open; _, open = <-conn.Msg {
		}
		quit <- os.Interrupt
	}()

	<-quit
}
