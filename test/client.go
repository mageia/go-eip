package test

import (
	"go-eip"
	"log"
	"runtime"
	"strings"
	"testing"
)

func ClientTestAll(t *testing.T, client go_eip.Client) {
	ClientTestGetPLCTime(t, client)
}

func ClientTestGetPLCTime(t *testing.T, client go_eip.Client) {
	//r, _ := client.GetPLCTime()
	//log.Println(r.Format(time.RFC3339))
	////
	//client.SetPLCTime(time.Now())
	////
	//r, _ = client.GetPLCTime()
	//log.Println(r.Format(time.RFC3339))
	//AssertEquals(t, 5, 5)
	log.Println(client.GetTagList())
	//
	//log.Println(client.Read("Program:MainProgram.first.1", 1))
	//log.Println(client.Read("Program:MainProgramTest.first.1"))
	client.Write("Program:MainProgram.first.7", false)
	client.Write("Program:MainProgram.first.0", true)
	client.Write("Program:MainProgram.first.1", true)
	client.Write("Program:MainProgram.first.2", true)
	client.Write("Program:MainProgram.first.3", true)
}

func AssertEquals(t *testing.T, expected, actual interface{}) {
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
