package go_eip

import "time"

type Client interface {
	Read(string, int) (interface{}, error)
	Write(string, interface{}) error
	MultiRead()
	GetPLCTime() (time.Time, error)
	SetPLCTime(time.Time) error
	GetTagList() ([]Tag, error)
	Discover()
}
