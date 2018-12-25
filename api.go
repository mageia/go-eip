package go_eip

import "time"

type Client interface {
	Read(string) (interface{}, error)
	Write(string, interface{}) error
	MultiRead(...string) (map[string]interface{}, error)
	GetPLCTime() (time.Time, error)
	SetPLCTime(time.Time) error
	GetTagList() ([]Tag, error)
	Discover()
	Stop()
}
