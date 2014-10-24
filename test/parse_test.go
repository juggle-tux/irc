package main

import (
	"github.com/juggle-tux/irc"
	"testing"
)

var (
	testServerRaw     = []byte(":wilhelm.freenode.net 001 nc-test :Welcome to the freenode Internet Relay Chat Network nc-test\r\n")
	testServerMessage = irc.Message{
		Prefix:   irc.Prefix{Host: "wilhelm.freenode.net"},
		Command:  "001",
		Parms:    irc.Parms{"nc-test"},
		Trailing: "Welcome to the freenode Internet Relay Chat Network nc-test",
	}

	testUserRaw     = []byte(":schaeffer!~simcity20@unaffiliated/simcity2000 PRIVMSG #go-nuts :whyrusleeping: noted!\r\n")
	testUserMessage = irc.Message{
		Prefix:   irc.Prefix{Nick: "schaeffer", User: "~simcity20", Host: "unaffiliated/simcity2000"},
		Command:  "PRIVMSG",
		Parms:    irc.Parms{"#go-nuts"},
		Trailing: "whyrusleeping: noted!",
	}
)

func TestServerMessage(t *testing.T) {
	msg, err := irc.ParseMessage(testServerRaw)
	if err != nil {
		t.Fatal("parse server msg fail: ", msg, err)
	}
	if msg.Prefix.String() != testServerMessage.Prefix.String() {
		t.Fatalf("prefix not the same got %s want %s", msg.Prefix, testServerMessage.Prefix)
	}
	if msg.Command != testServerMessage.Command {
		t.Fatalf("command not the same got %s want %s", msg.Command, testServerMessage.Command)
	}
	if msg.Parms.String() != testServerMessage.Parms.String() {
		t.Fatalf("parms not the same got %s want %s", msg.Parms, testServerMessage.Parms)
	}
	if msg.Trailing != testServerMessage.Trailing {
		t.Fatalf("tail not the same got %s want %s", msg.Trailing, testServerMessage.Trailing)
	}
	if msg.String() != testServerMessage.String() {
		t.Fatalf("message not the same got %s want %s", msg, testServerMessage)
	}
}

func BenchmarkServerMessageParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		irc.ParseMessage(testServerRaw)
	}
}

func BenchmarkServerMessageString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		testServerMessage.String()
	}
}

func TestUserMessage(t *testing.T) {
	msg, err := irc.ParseMessage(testUserRaw)
	if err != nil {
		t.Fatal("parse server msg fail: ", msg, err)
	}
	if msg.Prefix.String() != testUserMessage.Prefix.String() {
		t.Fatalf("prefix not the same got %s want %s", msg.Prefix, testUserMessage.Prefix)
	}
	if msg.Command != testUserMessage.Command {
		t.Fatalf("command not the same got %s want %s", msg.Command, testUserMessage.Command)
	}
	if msg.Parms.String() != testUserMessage.Parms.String() {
		t.Fatalf("parms not the same got %s want %s", msg.Parms, testUserMessage.Parms)
	}
	if msg.Trailing != testUserMessage.Trailing {
		t.Fatalf("tail not the same got %s want %s", msg.Trailing, testUserMessage.Trailing)
	}
	if msg.String() != testUserMessage.String() {
		t.Fatalf("message not the same got %s want %s", msg, testUserMessage)
	}
}

func BenchmarkUserMessageParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		irc.ParseMessage(testUserRaw)
	}
}

func BenchmarkUserMessageString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		testUserMessage.String()
	}
}
