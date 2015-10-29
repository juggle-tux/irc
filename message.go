package irc

import (
	"errors"
	"strings"
)

// Prefix represents an IRC prefix "Nick!User@Host"
type Prefix struct {
	Nick, User, Host string
}

// String outputs the Prefix in the form "Nick!User@Host" or only "Host" if Nick and
// User is empty
func (p Prefix) String() string {
	if p.Nick == "" && p.User == "" {
		return p.Host
	}
	return p.Nick + "!" + p.User + "@" + p.Host
}

// ParsePrefix parses the string into a new Prefix. Name and User will be empty if
// the "!" and "@" are not in the string or at the wrong positon. In that case the
// entry input string will be in Host
func ParsePrefix(prefix string) Prefix {
	iu := strings.Index(prefix, "!")
	ih := strings.Index(prefix, "@")
	if iu < 0 || ih < 0 || (ih-iu) < 0 {
		return Prefix{Host: prefix}
	}
	return Prefix{
		Nick: prefix[:iu],
		User: prefix[iu+len("!") : ih],
		Host: prefix[ih+len("@"):],
	}
}

// Parms parameters
type Parms []string

func (p Parms) String() string {
	var str string
	for _, s := range p {
		str += s + " "
	}
	return str
}

// Message represents a IRC Message like it gets send over the TCP stream
type Message struct {
	Prefix   Prefix
	Command  string
	Parms    Parms
	Trailing string
}

// String returns a string represantation of the Message and contains a terminating \r\n
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

// ParseMessage parses the raw Message
func ParseMessage(b []byte) (Message, error) {
	var m Message
	var tmp []string
	str := string(b)

	if strings.HasPrefix(str, ":") {
		tmp = strings.SplitN(str, " ", 2)
		m.Prefix = ParsePrefix(tmp[0][1:])
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
	return m, nil
}

// Op creates a MODE message to set "+o" to nick in channel
func Op(nick, channel string) Message {
	return Message{
		Command: "MODE",
		Parms: Parms{
			0: channel,
			1: "+o",
			2: nick,
		},
	}
}

// Msg creates a PRIVMSG so recv (channel/nick) with the conntent of str
func Msg(recv, str string) Message {
	return Message{
		Command:  "PRIVMSG",
		Parms:    Parms{0: recv},
		Trailing: str,
	}
}

// Join creates a JOIN message to join channel
func Join(channel string) Message {
	return Message{
		Command: "JOIN",
		Parms:   Parms{0: channel},
	}
}
