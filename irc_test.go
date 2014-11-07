package irc

import (
	"testing"
)

var (
	testServerRaw     = []byte(":wilhelm.freenode.net 001 nc-test :Welcome to the freenode Internet Relay Chat Network nc-test\r\n")
	testServerMessage = Message{
		Prefix:   Prefix{Host: "wilhelm.freenode.net"},
		Command:  "001",
		Parms:    Parms{"nc-test"},
		Trailing: "Welcome to the freenode Internet Relay Chat Network nc-test",
	}

	testUserRaw     = []byte(":schaeffer!~simcity20@unaffiliated/simcity2000 PRIVMSG #go-nuts :whyrusleeping: noted!\r\n")
	testUserMessage = Message{
		Prefix:   Prefix{Nick: "schaeffer", User: "~simcity20", Host: "unaffiliated/simcity2000"},
		Command:  "PRIVMSG",
		Parms:    Parms{"#go-nuts"},
		Trailing: "whyrusleeping: noted!",
	}
)

func TestMode(t *testing.T) {
	tM := Mode{}
	if err := tM.SetMode("+vn"); err != nil {
		t.Fatal(err)
	}
	if err := tM.SetMode("+i"); err != nil {
		t.Fatal(err)
	}
	if err := tM.SetMode("-v"); err != nil {
		t.Fatal(err)
	}
	if err := tM.SetMode("+o"); err != nil {
		t.Fatal(err)
	}
	for i := range tM {
		if i == 'o' || i == 'i' || i == 'n' {
			continue
		}
		t.Fail()
	}
	t.Logf("%s", tM)
}
func TestServerMessage(t *testing.T) {
	msg, err := ParseMessage(testServerRaw)
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
func TestUserMessage(t *testing.T) {
	msg, err := ParseMessage(testUserRaw)
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

func BenchmarkServerMessageParse(b *testing.B) {
	b.SetBytes(int64(len(testServerRaw)))
	for i := 0; i < b.N; i++ {
		ParseMessage(testServerRaw)
	}
}
func BenchmarkServerMessageString(b *testing.B) {
	b.SetBytes(int64(len(testServerRaw)))
	for i := 0; i < b.N; i++ {
		testServerMessage.String()
	}
}
func BenchmarkUserMessageParse(b *testing.B) {
	b.SetBytes(int64(len(testUserRaw)))
	for i := 0; i < b.N; i++ {
		ParseMessage(testUserRaw)
	}
}
func BenchmarkUserMessageString(b *testing.B) {
	b.SetBytes(int64(len(testUserRaw)))
	for i := 0; i < b.N; i++ {
		testUserMessage.String()
	}
}
