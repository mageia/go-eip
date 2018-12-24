package go_eip

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

type ClientHandler interface {
	Packager
	Transporter
}

type client struct {
	packager    Packager
	transporter Transporter
}

type Tag struct {
	TagName  string
	Offset   uint16
	DataType uint8
}

func NewClient(handler ClientHandler) Client {
	c := &client{packager: handler, transporter: handler}

	resp, err := c.transporter.Send(c.packager.BuildRegisterSessionRequest())
	if err != nil {
		c.transporter.Close()
		return nil
	}
	var sessionHandle, networkCID uint32
	binary.Read(bytes.NewBuffer(resp[4:8]), binary.LittleEndian, &sessionHandle)
	c.packager.SetSessionHandle(sessionHandle)

	resp, err = c.transporter.Send(c.packager.BuildForwardOpenRequest())
	if err != nil {
		c.transporter.Close()
		return nil
	}
	binary.Read(bytes.NewBuffer(resp[44:48]), binary.LittleEndian, &networkCID)
	c.packager.SetNetworkConnectionID(networkCID)

	return c
}

func (c *client) Read(tag string, element int) (interface{}, error) {
	var status uint8
	tagSplit := strings.Split(tag, ".")

	t, b, i := c.packager.TagNameParser(tag, 0)
	r := c.packager.BuildPartialReadRequest(t, b)
	response, err := c.send(NewProtocolDataUnit(r))
	if err != nil {
		return nil, err
	}
	if e := binary.Read(bytes.NewBuffer(response.Data[48:49]), binary.LittleEndian, &status); e != nil {
		return nil, e
	}

	if status != 0 && status != 6 {
		return nil, errors.New(ErrorText(int(status)))
	}

	var dataType uint8
	binary.Read(bytes.NewBuffer(response.Data[50:51]), binary.LittleEndian, &dataType)

	var requestData []byte
	bitCount := c.packager.GetBitCount(dataType).BitCount * 8

	if dataType == 211 {
		words := (i + element) / 4
		if (i+element)%32 > 0 {
			words += 1
		}
		requestData = c.packager.BuildReadIOIRequest(tag, true, words)
	} else {
		if pos, err := strconv.Atoi(tagSplit[len(tagSplit)-1]); err == nil {
			words := (pos + element) / int(bitCount)
			if (pos+element)%32 > 0 {
				words += 1
			}
			requestData = c.packager.BuildReadIOIRequest(tag, false, words)
		} else {
			requestData = c.packager.BuildReadIOIRequest(tag, false, element)
		}
	}
	response, err = c.send(NewProtocolDataUnit(requestData))
	if err != nil {
		return nil, err
	}
	if e := binary.Read(bytes.NewBuffer(response.Data[48:49]), binary.LittleEndian, &status); e != nil {
		return nil, e
	}
	if status != 0 && status != 6 {
		return nil, errors.New(ErrorText(int(status)))
	}

	log.Printf("%02x", response.Data)

	return nil, nil
}
func (c *client) Write(tag string, value interface{}) error {
	var status uint8
	request := c.packager.BuildWriteIOIRequest(tag, value)
	response, err := c.send(NewProtocolDataUnit(request))
	if err != nil {
		return err
	}
	if e := binary.Read(bytes.NewBuffer(response.Data[48:49]), binary.LittleEndian, &status); e != nil {
		return e
	}

	if status == 0 {
		return nil
	}

	return errors.New(ErrorText(int(status)))
}
func (c *client) MultiRead() {
}
func (c *client) GetPLCTime() (time.Time, error) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, struct {
		AttributeService      uint8
		AttributeSize         uint8
		AttributeClassType    uint8
		AttributeClass        uint8
		AttributeInstanceType uint8
		AttributeInstance     uint8
		AttributeCount        uint16
		TimeAttribute         uint16
	}{
		0x03,
		0x02,
		0x20,
		0x8B,
		0x24,
		0x01,
		0x01,
		0x0B,
	})

	eipHeader := c.packager.BuildEIPHeader(buf.Bytes())
	response, err := c.send(NewProtocolDataUnit(eipHeader))
	if err != nil {
		return time.Time{}, err
	}
	var plcTimeMSOffset uint64
	binary.Read(bytes.NewBuffer(response.Data[56:64]), binary.LittleEndian, &plcTimeMSOffset)
	originTime := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	return originTime.Add(time.Microsecond * time.Duration(plcTimeMSOffset)), nil
}
func (c *client) SetPLCTime(t time.Time) error {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, struct {
		AttributeService      uint8
		AttributeSize         uint8
		AttributeClassType    uint8
		AttributeClass        uint8
		AttributeInstanceType uint8
		AttributeInstance     uint8
		AttributeCount        uint16
		TimeAttribute         uint16
		Time                  uint64
	}{
		0x04,
		0x02,
		0x20,
		0x8B,
		0x24,
		0x01,
		0x01,
		0x06,
		uint64(t.UnixNano()) / 1e3,
	})
	eipHeader := c.packager.BuildEIPHeader(buf.Bytes())
	_, err := c.send(NewProtocolDataUnit(eipHeader))
	if err != nil {
		return err
	}
	return nil
}
func (c *client) GetTagList() ([]Tag, error) {
	tagList := make([]Tag, 0)

	tList, e := c._getTagList("")
	if e != nil {
		log.Println(e)
		return tagList, e
	}
	tagList = append(tagList, tList...)

	for _, p := range c.packager.GetProgramName() {
		tList, e := c._getTagList(p)
		if e != nil {
			log.Println(e)
			return tagList, e
		}
		tagList = append(tagList, tList...)
	}

	for _, t := range tagList {
		c.packager.FillKnownTags(t.TagName, t.DataType)
	}

	return tagList, nil
}
func (c *client) Discover() {}

func (c *client) _getTagList(p string) ([]Tag, error) {
	var status uint8
	tagList := make([]Tag, 0)

	if p != "" {
		c.packager.SetOffset(0)
	}

	tagListRequest := c.packager.BuildTagListRequest(p)
	response, err := c.send(NewProtocolDataUnit(tagListRequest))
	if err != nil {
		return tagList, err
	}
	if e := binary.Read(bytes.NewBuffer(response.Data[48:49]), binary.LittleEndian, &status); e != nil {
		return tagList, e
	}
	tList, e := c.extractTagPacket(response.Data, p)
	if e != nil {
		return tagList, e
	}
	tagList = append(tagList, tList...)

	for status == 6 {
		c.packager.IncreaseOffset(1)
		tagListRequest := c.packager.BuildTagListRequest(p)
		response, err := c.send(NewProtocolDataUnit(tagListRequest))
		if err != nil {
			return tagList, err
		}
		if e := binary.Read(bytes.NewBuffer(response.Data[48:49]), binary.LittleEndian, &status); e != nil {
			return tagList, e
		}
		tList, e := c.extractTagPacket(response.Data, p)
		if e != nil {
			return tagList, e
		}
		tagList = append(tagList, tList...)
		time.Sleep(250 * time.Millisecond)
	}

	return tagList, nil
}
func (c *client) extractTagPacket(data []byte, programName string) ([]Tag, error) {
	var packetStart uint16 = 50
	var tagLen uint16
	tagList := make([]Tag, 0)

	for int(packetStart) < len(data) {
		if e := binary.Read(bytes.NewBuffer(data[packetStart+8:packetStart+10]), binary.LittleEndian, &tagLen); e != nil {
			return tagList, e
		}
		packet := data[packetStart : packetStart+tagLen+10]
		var offset uint16
		if e := binary.Read(bytes.NewBuffer(packet[:2]), binary.LittleEndian, &offset); e != nil {
			return tagList, e
		}
		c.packager.SetOffset(uint32(offset))
		tag, _ := c.parseTag(packet, programName)
		tagList = append(tagList, tag)
		if programName == "" {
			if strings.Contains(tag.TagName, "Program:") {
				c.packager.AddProgramName(tag.TagName)
			}
		}
		packetStart += tagLen + 10
	}
	return tagList, nil
}
func (c *client) parseTag(packet []byte, programName string) (Tag, error) {
	t := Tag{}
	var length uint16
	if e := binary.Read(bytes.NewBuffer(packet[8:10]), binary.LittleEndian, &length); e != nil {
		return t, e
	}

	t.TagName = string(packet[10 : length+10])
	if programName != "" {
		t.TagName = programName + "." + t.TagName
	}
	if e := binary.Read(bytes.NewBuffer(packet[:2]), binary.LittleEndian, &t.Offset); e != nil {
		return t, e
	}
	if e := binary.Read(bytes.NewBuffer(packet[4:5]), binary.LittleEndian, &t.DataType); e != nil {
		log.Println(e)
	}

	return t, nil
}
func (c *client) send(request *ProtocolDataUnit) (response *ProtocolDataUnit, err error) {
	dataResponse, err := c.transporter.Send(request.Data)

	if err = c.packager.Verify(request.Data, dataResponse); err != nil {
		return
	}
	if dataResponse == nil || len(dataResponse) == 0 {
		err = fmt.Errorf("eip: response data is empty")
		return
	}
	response = &ProtocolDataUnit{Data: dataResponse}
	err = responseError(response)
	return response, err
}

func responseError(response *ProtocolDataUnit) error {
	return nil
}
