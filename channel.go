package irc

import "errors"

type Mode map[byte]struct{}

func (m Mode) SetMode(s string) error {
	if len(s) < 2 {
		return errors.New("string to short")
	}
	if s[0] == '+' {
		for _, c := range []byte(s[1:]) {
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
				m[c] = struct{}{}
				continue
			}
			break
		}
	} else if s[0] == '-' {
		for _, c := range []byte(s[1:]) {
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
				delete(m, c)
				continue
			}
			break
		}
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
	nicks  []string
	topic  string
	cMode  Mode
	myMode Mode

	cl *Client
}

func (c *Channel) Name() string {
	return c.name
}

func (c *Channel) Topic() string {
	return c.topic
}

func (c *Channel) Mode() string {
	return c.cMode.String()
}

func (ch *Channel) Part() {
	ch.cl.Send <- Message{
		Command: "PART",
		Parms:   Parms{0: ch.name},
	}
}
