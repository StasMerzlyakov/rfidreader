// Contain MIFARE PICC manipulate funcitons

package mfrc522

import (
	"bytes"
	"math/bits"
	"time"
)

type MFRC522Device interface {
	PCD_CalculateCRC(crcResetValue int, buffer []byte, duration time.Duration) ([]byte, error)
}

// LFSR16 Implementation
type Lfsr16FN func(cnt int) uint16

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

/**
MIFARE LFSR16
*/
func InitLfsr16FN(init uint16) Lfsr16FN {
	var state = (init & 0xff00 >> 8) | init&0x00ff<<8
	return func(cnt int /* cricle count */) uint16 {
		val := uint16(0)
		for i := 0; i < cnt; i++ {
			bit := (state & 0x20 >> 5) ^ (state & 0x8 >> 3) ^ (state & 0x4 >> 2) ^ state&0x1
			val = val<<1 | bit
			state = (state>>1)&0x7FFF | (bit<<15)&0x8000
		}
		val = bits.Reverse16(val)
		return (val & 0xff00 >> 8) | val&0x00ff<<8
	}
}
