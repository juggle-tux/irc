package irc

import (
	"errors"
	"fmt"
	"log"
	"strings"
)

type Mode map[byte]struct{}

func (m Mode) SetMode(s string) error {
	if len(s) < 2 {
		return errors.New("string to short")
	}
	op, list := s[0], []byte(s[1:])
	for _, c := range list {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')) {
			return fmt.Errorf("mode must be between a-Z got %q", c)
		}

	}

	switch op {
	case '+':
		for _, c := range list {
			m[c] = struct{}{}

		}
	case '-':
		for _, c := range list {
			delete(m, c)
		}
	default:
		return fmt.Errorf("op most be +/- got %q", op)
	}
	return nil
}

func (m Mode) String() string {
	if len(m) <= 0 {
		return ""
	}
	str := "+"
	for s := range m {
		str += string(s)
	}
	return str
}

type Channel struct {
	name   string
	nicks  map[string]Mode
	topic  string
	cMode  Mode
	myMode Mode

	cl *Client
}

func (c *Channel) Name() string {
	return c.name
}

func (c *Channel) NamesMap() map[string]Mode {
	return c.nicks
}

func (c *Channel) Names() []string {
	str := []string{}
	for ni, m := range c.nicks {
		if _, ok := m['o']; ok {
			ni = "@" + ni
		} else if _, ok := m['v']; ok {
			ni = "+" + ni
		}
		str = append(str, ni)
	}
	return str
}

func (c *Channel) Topic() string {
	return c.topic
}

func (c *Channel) Mode() string {
	return c.cMode.String()
}

func (c *Channel) Part() {
	c.cl.send <- Message{
		Command: "PART",
		Parms:   Parms{0: c.name},
	}
}

type ChannelManager struct {
	channels       map[string]*Channel
	send           chan<- Message
	nick           string
	DefaultHandler Handler
}

func NewCM(cl *Client) *ChannelManager {
	cm := &ChannelManager{
		channels:       make(map[string]*Channel),
		send:           cl.send,
		nick:           cl.nick,
		DefaultHandler: defaultHandler,
	}
	return cm
}

func (cm *ChannelManager) chControl(req Message, res chan<- Message) bool {
	switch req.Command {
	case "JOIN":
		if req.Prefix.Nick == cm.nick {
			if _, ok := cm.channels[req.Parms[0]]; !ok {
				log.Print("join channel ", req.Parms[0])
				cm.channels[req.Parms[0]] = &Channel{
					name:   req.Parms[0],
					nicks:  make(map[string]Mode),
					cMode:  make(Mode),
					myMode: make(Mode),
				}
			}

			return true
		}
		if ch, ok := cm.channels[req.Parms[0]]; ok {
			log.Printf("%q joins %q", req.Prefix.Nick, req.Parms[0])
			ch.nicks[req.Prefix.Nick] = Mode{}
			return true
		}

	case "PART":
		if req.Prefix.Nick == cm.nick {
			log.Print("left channel ", req.Parms[0])
			delete(cm.channels, req.Parms[0])
			return true
		}
		if ch, ok := cm.channels[req.Parms[0]]; ok {
			log.Printf("%q left %q", req.Prefix.Nick, req.Parms[0])
			delete(ch.nicks, req.Prefix.Nick)
			return true
		}

	case "QUIT":
		log.Printf("%q QUIT %q", req.Prefix.Nick, req.Trailing)
		for _, ch := range cm.channels {
			if _, ok := ch.nicks[req.Prefix.Nick]; ok {
				delete(ch.nicks, req.Prefix.Nick)
			}
		}
		return true

	case "MODE":
		if ch, ok := cm.channels[req.Parms[0]]; ok {
			switch {
			case len(req.Parms) < 2:
				return false
			case len(req.Parms) == 2:
				log.Printf("%q sets %q %q", req.Prefix.String(), req.Parms[0], req.Parms[1])
				ch.cMode.SetMode(req.Parms[1])
				return true
			default:
				if us, ok := ch.nicks[req.Parms[2]]; ok {
					log.Printf("%q sets %q %q in %q", req.Prefix.String(), req.Parms[2], req.Parms[1], req.Parms[0])
					us.SetMode(req.Parms[1])
					return true
				}
			}
		}
		return false

	case "NICK":
		if nick := req.Prefix.Nick; nick != "" {
			for _, ch := range cm.channels {
				if m, ok := ch.nicks[nick]; ok {
					ch.nicks[req.Trailing] = m
					delete(ch.nicks, nick)
				}
			}
		}

	case RPL_TOPIC:
		if ch, ok := cm.channels[req.Parms[0]]; ok {
			ch.topic = req.Trailing
			return true
		}

	case RPL_NAMREPLY:
		if ch, ok := cm.channels[req.Parms[2]]; ok {
			for _, ni := range strings.Fields(req.Trailing) {
				m := Mode{}
				switch ni[0] {
				case '@':
					m.SetMode("+o")
					ni = ni[1:]
				case '+':
					m.SetMode("+v")
					ni = ni[1:]
				default:
				}
				ch.nicks[ni] = m
			}
			return true
		}

	case ERR_BANNEDFROMCHAN:
		log.Print(req.Parms[0] + ": " + req.Trailing)

	default:
	}

	return false
}

func (cm *ChannelManager) List() []Channel {
	cl := []Channel{}
	for _, ch := range cm.channels {
		cl = append(cl, *ch)
	}
	return cl
}

func (cm *ChannelManager) ServeIRC(req Message, res chan<- Message) bool {
	if cm.chControl(req, res) {
		return true
	}
	return cm.DefaultHandler.ServeIRC(req, res)
}
