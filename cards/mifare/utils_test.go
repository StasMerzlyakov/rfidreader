package mifare

import (
	"bytes"
	"log"
	"rfidreader/iso14443"
	"testing"
	"time"

	"github.com/matryer/is"
)

type Fn_PCD_CommunicateWithPICC func(dataToSend []byte,
	validBits *byte,
	duration time.Duration) (
	result []byte,
	err error)

type Fn_PCD_IsCollisionOccure func() (bool, error)

type Fn_PCD_CalculateCRC func(buffer []byte, duration time.Duration) ([]byte, error)

type Fn_PCD_Init func() error

type Fn_PCD_Reset func() error

type Fn_PCD_PerformSelfTest func() error

type Fn_PCD_AntennaOn func() error

type Fn_PCD_AntennaOff func() error

type MockISO14443Device struct {
	communicateWithPICCHandler Fn_PCD_CommunicateWithPICC
	isCollisionOccureHandler   Fn_PCD_IsCollisionOccure
	calculateCRCHandler        Fn_PCD_CalculateCRC
	initHandler                Fn_PCD_Init
	resetHandler               Fn_PCD_Reset
	performSelfTestHandler     Fn_PCD_PerformSelfTest
	antennaOnHandler           Fn_PCD_AntennaOn
	antennaOffHandler          Fn_PCD_AntennaOff
}

func (m *MockISO14443Device) PCD_CommunicateWithPICC(dataToSend []byte,
	validBits *byte,
	duration time.Duration) (
	result []byte,
	err error) {
	return m.communicateWithPICCHandler(dataToSend, validBits, duration)
}

func (m *MockISO14443Device) PCD_IsCollisionOccure() (bool, error) {
	return m.isCollisionOccureHandler()
}

func (m *MockISO14443Device) PCD_CalculateCRC(buffer []byte, duration time.Duration) ([]byte, error) {
	return m.calculateCRCHandler(buffer, duration)
}

func (m *MockISO14443Device) PCD_Init() error {
	return m.initHandler()
}

func (m *MockISO14443Device) PCD_Reset() error {
	return m.resetHandler()
}

func (m *MockISO14443Device) PCD_PerformSelfTest() error {
	return m.performSelfTestHandler()
}

func (m *MockISO14443Device) PCD_AntennaOn() error {
	return m.antennaOnHandler()
}

func (m *MockISO14443Device) PCD_AntennaOff() error {
	return m.antennaOffHandler()
}

func TestGenerateNUID(t *testing.T) {
	is := is.New(t)
	device := &MockISO14443Device{
		calculateCRCHandler: func(buffer []byte, duration time.Duration) ([]byte, error) {

			if len(buffer) < 2 {
				return []byte{0x0, 0x0}, nil
			}
			log.Printf("[% x]", buffer[:2])
			return buffer[:2], nil
		},
	}

	uid := iso14443.UID{
		Uid: []byte{0xf0, 0xf0, 0xf0, 0xf0},
	}

	nuid, err := GenerateNUID(uid, device)
	is.NoErr(err)
	is.True(bytes.Compare(nuid, []byte{0xff, 0xf0, 0xf0, 0xf0}) == 0)

	uid = iso14443.UID{
		Uid: []byte{0x10, 0x20, 0x30, 0x40, 0x50, 0x60, 0x70},
	}

	nuid, err = GenerateNUID(uid, device)
	is.NoErr(err)
	is.True(bytes.Compare(nuid, []byte{0x1f, 0x20, 0x40, 0x50}) == 0)

}
