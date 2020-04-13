package mifare

import (
	"bytes"
	"log"
	"rfidreader/mfrc522"
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

			if len(buffer) < 2 {
				return []byte{0x0, 0x0}, nil
			}
			log.Printf("[% x]", buffer[:2])
			return buffer[:2], nil
		},
	}

	uid := mfrc522.UID{
		Uid: []byte{0xf0, 0xf0, 0xf0, 0xf0},
	}

	nuid, err := GenerateNUID(uid, device)
	is.NoErr(err)
	is.True(bytes.Compare(nuid, []byte{0xff, 0xf0, 0xf0, 0xf0}) == 0)

	uid = mfrc522.UID{
		Uid: []byte{0x10, 0x20, 0x30, 0x40, 0x50, 0x60, 0x70},
	}

	nuid, err = GenerateNUID(uid, device)
	is.NoErr(err)
	is.True(bytes.Compare(nuid, []byte{0x1f, 0x20, 0x40, 0x50}) == 0)

}
