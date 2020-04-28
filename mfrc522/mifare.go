// Contain MIFARE PICC manipulate funcitons

package mfrc522

import (
	"bytes"
	"time"
)

type MFRC522Device interface {
	PCD_CalculateCRC(crcResetValue int, buffer []byte, duration time.Duration) ([]byte, error)
}

// LFSR16 Implementation
// byte order BIG-endian
type Lfsr16FN func() ([]byte, error)

// NUID generation
// https://www.nxp.com/docs/en/application-note/AN10927.pdf
// Reset the CRC calculator with the standard ISO/IEC 14443 type A preset values: 6363hex first
func GenerateNUID(uid UID, device MFRC522Device) (nuid []byte, err error) {

	var buff bytes.Buffer

	switch ln := len(uid.Uid); ln {
	case 4: // Single Size UID
		if err = buff.WriteByte((uid.Uid[0] | 0x0f) & 0xff); err != nil {
			return
		}
		buff.Write(uid.Uid[1:])
		nuid = buff.Bytes()
	case 7: // 	Double Size UID
		var crc []byte
		if crc, err = device.PCD_CalculateCRC(ISO_14443_CRC_RESET, uid.Uid[:3], INTERUPT_TIMEOUT); err != nil {
			return
		}
		buff.WriteByte((crc[0] | 0x0f) & 0xff)
		buff.WriteByte(crc[1])

		if crc, err = device.PCD_CalculateCRC(ISO_14443_CRC_RESET, uid.Uid[3:], INTERUPT_TIMEOUT); err != nil {
			return
		}
		buff.Write(crc)
		nuid = buff.Bytes()
	default:
		err = CommonError("Wrong uid lenght: %d. Expected 4 or 7")
	}
	return
}

var LFSR16Polinom = map[string]byte{
	"0000": '0',
	"0001": '1',
	"0010": '1',
	"0011": '0',
	"0100": '1',
	"0101": '0',
	"0110": '0',
	"0111": '1',
	"1000": '1',
	"1001": '0',
	"1010": '0',
	"1011": '1',
	"1100": '0',
	"1101": '1',
	"1110": '1',
	"1111": '0',
}

/**
MIFARE LFSR16
*/
func InitLfsr16FN(init []byte) Lfsr16FN {
	var state = uint16(init[0]) | uint16(init[1])<<8

	return func() ([]byte, error) {
		val := uint16(0)
		for i := 0; i < 16; i++ {
			bit := (state & 0x20 >> 5) ^ (state & 0x8 >> 3) ^ (state & 0x4 >> 2) ^ state&0x1
			val = val | bit<<i
			//val = val<<1 | bit
			state = (state>>1)&0x7FFF | (bit<<15)&0x8000
		}
		return []byte{byte(val & 0x00ff), byte(val & 0xff00 >> 8)}, nil
	}

}
