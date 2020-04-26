package mfrc522

import (
	"bytes"
	"log"
	"testing"
	"time"

	"github.com/matryer/is"
)

type Fn_PCD_CalculateCRC func(crcResetValue int, buffer []byte, duration time.Duration) ([]byte, error)

type MockISO14443Device struct {
	calculateCRCHandler Fn_PCD_CalculateCRC
}

func (m *MockISO14443Device) PCD_CalculateCRC(crcResetValue int, buffer []byte, duration time.Duration) ([]byte, error) {
	return m.calculateCRCHandler(crcResetValue, buffer, duration)
}
func TestGenerateNUID(t *testing.T) {
	is := is.New(t)

	device := &MockISO14443Device{
		calculateCRCHandler: func(crcResetValue int, buffer []byte, duration time.Duration) ([]byte, error) {
			result := ISO14443aCRC(buffer)
			return result, nil
		},
	}

	uid := UID{
		Uid: []byte{0xf0, 0xf0, 0xf0, 0xf0},
	}

	nuid, err := GenerateNUID(uid, device)
	is.NoErr(err)
	is.True(bytes.Compare(nuid, []byte{0xff, 0xf0, 0xf0, 0xf0}) == 0)

	uid = UID{
		Uid: []byte{0x10, 0x20, 0x30, 0x40, 0x50, 0x60, 0x70},
	}

	nuid, err = GenerateNUID(uid, device)
	is.NoErr(err)
	is.True(bytes.Compare(nuid, []byte{0x3f, 0x32, 0x86, 0xd5}) == 0)

}

func TestLfsr16FN(t *testing.T) {
	is := is.New(t)

	test_data := map[uint16]uint16{
		0x4297: 0xc0a4,
		0x0120: 0x0145,
		0x4ca3: 0xec7a,
		0x6876: 0x8c86,
		0x93a6: 0xd176,
		0x632e: 0x4481,
		0xe7a3: 0x7d92,
	}

	for k, v := range test_data {
		f := InitLfsr16FN(k)
		res := f(16)
		if res != v {
			log.Printf("k=%x v=%x res=%x", k, v, res)
		}
		is.True(res == v)
	}

}
