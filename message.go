package irc

import (
	"errors"
	"io"
	"log"
	"strings"
)

type Prefix struct {
	Nick, User, Host string
}

func (p Prefix) String() string {
	if p.Nick == "" && p.User == "" {
		return p.Host
	}
	return p.Nick + "!" + p.User + "@" + p.Host
}

func ParsePrefix(prefix string) (Prefix, error) {
	iu := strings.Index(prefix, "!")
	ih := strings.Index(prefix, "@")
	if iu < 0 || ih < 0 || (ih-iu) < 0 {
		return Prefix{Host: prefix}, nil
	}
	return Prefix{
		Nick: prefix[:iu-len("!")],
		User: prefix[iu+len("!") : ih-len("@")],
		Host: prefix[ih+len("@"):],
	}, nil
}

type Parms []string

func (p Parms) String() string {
	var str string
	for _, s := range p {
		str += s + " "
	}
	return str
}

type Message struct {
	Prefix   Prefix
	Command  string
	Parms    Parms
	Trailing string
}

func (m Message) String() string {
	var tail string
	if m.Trailing == "" {
		tail = ""
	} else {
		tail = ":" + m.Trailing
	}
	if m.Prefix.Host == "" {
		return m.Command + " " + m.Parms.String() + tail + "\r\n"
	}
	return ":" + m.Prefix.String() + " " + m.Command + " " + m.Parms.String() + tail + "\r\n"
}

var _ = io.EOF

func ParseMessage(b []byte) (m Message, err error) {
	str := string(b)
	if len(b) < 3 {
		return m, errors.New("message to short")
	}
	sep := strings.Index(str[1:], ":") + 1
	if sep > 0 {
		m.Trailing = str[sep+1:]
		str = str[:sep-1]
	}

	left := strings.Fields(str)
	if left[0][0] == ':' {
		m.Prefix, err = ParsePrefix(left[0][1:])
		if err != nil {
			return
		}
		left = left[1:]
	}
	m.Command = left[0]
	m.Parms = left[1:]
	return
}
