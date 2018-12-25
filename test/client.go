package test

import (
	"go-eip"
	"log"
	"runtime"
	"strings"
	"testing"
	"time"
)

func ClientTestAll(t *testing.T, client go_eip.Client) {
	//ClientTestGetPLCTime(t, client)
	//ClientTestReadWriteString(t, client)
	ClientTestReadWriteSINT(t, client)
	ClientTestReadWriteBit(t, client)
	ClientTestMultiRead(t, client)
}

func ClientTestGetPLCTime(t *testing.T, client go_eip.Client) {
	r, _ := client.GetPLCTime()
	log.Println(r.Format(time.RFC3339))
	client.SetPLCTime(time.Now())
	r, _ = client.GetPLCTime()
	log.Println(r.Format(time.RFC3339))
}
func ClientTestReadWriteString(t *testing.T, client go_eip.Client) {
	client.Write("Program:MainProgram.string", "abcd")
	r, e := client.Read("Program:MainProgram.string")
	AssertEquals(t, e, nil)
	AssertEquals(t, r, "abcd")
}
func ClientTestReadWriteSINT(t *testing.T, client go_eip.Client) {
	client.Write("Program:MainProgram.sint", 12)
	r, e := client.Read("Program:MainProgram.sint")
	AssertEquals(t, e, nil)
	AssertEquals(t, r, uint8(12))
}
func ClientTestReadWriteBit(t *testing.T, client go_eip.Client) {
	client.Write("Program:MainProgram.first.1", false)
	r1, e1 := client.Read("Program:MainProgram.first.1")
	AssertEquals(t, e1, nil)
	AssertEquals(t, r1, false)

	client.Write("Program:MainProgram.third.15", true)

	r1, e1 = client.Read("Program:MainProgram.third.15")
	AssertEquals(t, r1 == true, true)
}
func ClientTestMultiRead(t *testing.T, client go_eip.Client) {
	log.Println(client.MultiRead(
		"Program:MainProgram.first",
		"Program:MainProgram.first.1",
		"Program:MainProgram.sint",
		"Program:MainProgram.sint",
		"Program:MainProgram.sint",
		"Program:MainProgram.sint",
		"Program:MainProgram.sint",
		"Program:MainProgram.sint",
		"Program:MainProgram.string",
		"Program:MainProgram.string",
		"Program:MainProgram.string",
		"Program:MainProgram.string",
		"Program:MainProgram.sint.2",
		))
}


func AssertEquals(t *testing.T, expected interface{}, actual interface{}) {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "???"
		line = 0
	} else {
		idx := strings.LastIndex(file, "/")
		if idx >= 0 {
			file = file[idx+1:]
		}
	}

	if expected != actual {
		t.Logf("%s:%d: Expected: %+v (%T), actual: %+v (%T)", file, line,
			expected, expected, actual, actual)
		t.FailNow()
	}
}
