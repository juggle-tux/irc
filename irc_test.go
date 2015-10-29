package irc

import (
	"fmt"
	"testing"
)

type testMsg struct {
	raw []byte
	msg *Message
	err error
}

var tests = map[string]testMsg{
	"server": {
		raw: []byte(":wilhelm.freenode.net 001 nc-test :Welcome to the freenode Internet Relay Chat Network nc-test\r\n"),
		msg: &Message{
			Prefix:   Prefix{Host: "wilhelm.freenode.net"},
			Command:  "001",
			Parms:    Parms{"nc-test"},
			Trailing: "Welcome to the freenode Internet Relay Chat Network nc-test",
		},
		err: nil,
	},
	"user": {
		raw: []byte(":schaeffer!~simcity20@unaffiliated/simcity2000 PRIVMSG #go-nuts :whyrusleeping: noted!\r\n"),
		msg: &Message{
			Prefix:   Prefix{Nick: "schaeffer", User: "~simcity20", Host: "unaffiliated/simcity2000"},
			Command:  "PRIVMSG",
			Parms:    Parms{"#go-nuts"},
			Trailing: "whyrusleeping: noted!",
		},
		err: nil,
	},
}

func TestMode(t *testing.T) {
	tM := Mode{}

	if err := tM.SetMode("+a"); err != nil {
		t.Fatal(err)
	}
	if err := tM.SetMode("+bc"); err != nil {
		t.Fatal(err)
	}
	if err := tM.SetMode("-ca"); err != nil {
		t.Fatal(err)
	}
	if err := tM.SetMode("-b"); err != nil {
		t.Fatal(err)
	}
	if str := tM.String(); str != "" {
		t.Fatalf("mode shoud be empty got %q", str)
	}
	if err := tM.SetMode("+o"); err != nil {
		t.Fatal(err)
	}
	if str := tM.String(); str != "+o" {
		t.Fatalf("mode shoud be \"+o\" got %q", str)
	}
}

func TestServerMessage(t *testing.T) {
	test := tests["server"]
	msg, err := ParseMessage(test.raw)
	if err != nil {
		t.Fatal("parse server msg fail: ", msg, err)
	}
	if err := test.eqMsg(msg); err != nil {
		t.Fatal(err)
	}
}

func TestUserMessage(t *testing.T) {
	test := tests["user"]
	msg, err := ParseMessage(test.raw)
	if err != nil {
		t.Fatal("parse server msg fail: ", msg, err)
	}
	if err := test.eqMsg(msg); err != nil {
		t.Fatal(err)
	}
}

func BenchmarkServerMessageParse(b *testing.B) {
	test := tests["server"].raw
	b.SetBytes(int64(len(test)))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseMessage(test)
	}
}

func BenchmarkServerMessageString(b *testing.B) {
	test := tests["server"]
	b.SetBytes(int64(len(test.raw)))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = test.msg.String()
	}
}

func BenchmarkServerMessageParseString(b *testing.B) {
	test := tests["server"]
	b.SetBytes(int64(len(test.raw)))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m, _ := ParseMessage(test.raw)
		_ = m.String()
	}
}

func BenchmarkUserMessageParse(b *testing.B) {
	test := tests["user"].raw
	b.SetBytes(int64(len(test)))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseMessage(test)
	}
}

func BenchmarkUserMessageString(b *testing.B) {
	test := tests["user"]
	b.SetBytes(int64(len(test.raw)))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = test.msg.String()
	}
}

func BenchmarkUserMessageParseString(b *testing.B) {
	test := tests["user"]
	b.SetBytes(int64(len(test.raw)))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m, _ := ParseMessage(test.raw)
		_ = m.String()
	}
}

func BenchmarkUserMessageParseStringPipe(b *testing.B) {
	test := tests["user"]
	b.SetBytes(int64(len(test.raw)))
	res := make(chan *Message, 64)

	b.ResetTimer()
	go func() {
		for i := 0; i < b.N; i++ {
			m, _ := ParseMessage(test.raw)

			res <- &m
		}
		close(res)
	}()
	for m := range res {
		_ = m.String()
	}
}

func (tm testMsg) eqMsg(m Message) error {
	switch {
	case m.Prefix.String() != tm.msg.Prefix.String():
		return fmt.Errorf("prefix not the same got %s want %s", m.Prefix, tm.msg.Prefix)
	case m.Command != tm.msg.Command:
		return fmt.Errorf("command not the same got %s want %s", m.Command, tm.msg.Command)
	case m.Parms.String() != tm.msg.Parms.String():
		return fmt.Errorf("parms not the same got %s want %s", m.Parms, tm.msg.Parms)
	case m.Trailing != tm.msg.Trailing:
		return fmt.Errorf("tail not the same got %s want %s", m.Trailing, tm.msg.Trailing)
	case m.String() != string(tm.raw):
		return fmt.Errorf("message not the same got %s want %s", m, string(tm.raw))
	default:
		return nil
	}
}
