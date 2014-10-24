package irc

import (
	"errors"
	"log"
	"strings"
)

var _ = log.Lshortfile

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
		Nick: prefix[:iu],
		User: prefix[iu+len("!") : ih],
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

func ParseMessage(b []byte) (m Message, err error) {
	str := string(b)
	var tmp []string

	if strings.HasPrefix(str, ":") {
		tmp = strings.SplitN(str, " ", 2)
		m.Prefix, err = ParsePrefix(tmp[0][1:])
		if err != nil {
			return m, err
		}
		str = tmp[1]
	}

	tmp = strings.SplitN(str, ":", 2)
	if len(tmp) > 1 {
		m.Trailing = strings.TrimSpace(tmp[1])
	}

	tmp = strings.Fields(tmp[0])
	l := len(tmp)
	if l < 1 {
		return m, errors.New("massage has no command")
	}
	m.Command = tmp[0]

	if l == 1 {
		return m, nil

	}
	m.Parms = tmp[1:]
	return
}
