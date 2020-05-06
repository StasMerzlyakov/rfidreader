// Contain MIFARE PICC manipulate funcitons

package mfrc522

import (
	"bytes"
	"math/rand"
	"time"
)

type MFRC522Device interface {
	PCD_CalculateCRC(crcResetValue int, buffer []byte, duration time.Duration) ([]byte, error)
}

// LFSR16 Implementation
// byte order BIG-endian
type Lfsr16FN func() ([]byte, error)

// Mifare encrypt register
type Lfsr32FN func(input []byte) ([]byte, error)

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

func GenerateNR() []byte {
	// 0, 1
	value := []byte{uint8(rand.Uint32()), uint8(rand.Uint32())}
	tailFn := InitLfsr16FN(value)
	tail, _ := tailFn()
	return append(value, tail...)
}

/**
  MIFARE LFSR16
*/
func InitLfsr16FN(init []byte) Lfsr16FN {

	// BIT ORDER relative implementation
	// Little endian !!!

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

/**
  MIFARE LSFR32 encryptor
 	uid : uid0, uid1, uid2, uid3
	nt : nt0, nt1, nt2, nt3
	key: key0, key1, key2, key3, key4, key5
*/

func f4(rcode uint16, x3, x2, x1, x0 byte) byte {
	uid := uint16((x3&1)<<3) | uint16((x2&1)<<2) | uint16((x1&1)<<1) | uint16(x0&1)
	return byte(((rcode & (uint16(1) << uid)) >> uid) & 1)
}

func f5(rcode uint32, x4, x3, x2, x1, x0 byte) byte {
	uid := uint32((x4&1)<<4) | uint32((x3&1)<<3) | uint32((x2&1)<<2) | uint32((x1&1)<<1) | uint32(x0&1)
	return byte(((rcode & (uint32(1) << uid)) >> uid) & 1)
}

func Fc(x4, x3, x2, x1, x0 byte) byte {
	return f5(0xec57e80a, x4, x3, x2, x1, x0)
}

func Fa(x3, x2, x1, x0 byte) byte {
	return f4(0x9e98, x3, x2, x1, x0)
}

func Fb(x3, x2, x1, x0 byte) byte {
	return f4(0xb48e, x3, x2, x1, x0)
}

func InitLfsr32FN(key []byte, uid []byte, nt []byte) Lfsr32FN {

	// BIT ORDER relative implementation
	// Little endian !!!
	var state = uint64(key[0]) | uint64(key[1])<<8 | uint64(key[2])<<16 |
		uint64(key[3])<<24 | uint64(key[4])<<32 | uint64(key[5])<<40

	init := uint32(uid[0]^nt[0]) | uint32(uid[1]^nt[1])<<8 |
		uint32(uid[2]^nt[2])<<16 | uint32(uid[3]^nt[3]<<24)

	var lsfrN = func() uint64 {
		return state&0x1 ^ (state & 0x1 >> 5) ^ (state & 0x1 >> 9) ^ (state & 0x1 >> 10) ^
			(state & 0x1 >> 12) ^ (state & 0x1 >> 14) ^ (state & 0x1 >> 15) ^ (state & 0x1 >> 17) ^
			(state & 0x1 >> 19) ^ (state & 0x1 >> 24) ^ (state & 0x1 >> 27) ^ (state & 0x1 >> 29) ^
			(state & 0x1 >> 35) ^ (state & 0x1 >> 39) ^ (state & 0x1 >> 41) ^ (state & 0x1 >> 42) ^
			(state & 0x1 >> 43)
	}

	// Initialization cycle
	for i := 0; i < 32; i++ {
		bit := lsfrN()&0x1 ^ uint64(init&0x1)
		init = init >> 1
		state = (state>>1)&0x7FFFFF | (bit<<47)&0x800000
	}

	return func(input []byte) ([]byte, error) {
		val := uint16(0)
		for i := 0; i < 16; i++ {
			bit := uint64((state & 0x20 >> 5) ^ (state & 0x8 >> 3) ^ (state & 0x4 >> 2) ^ state&0x1)
			val = val | bit<<i
			//val = val<<1 | bit
			state = (state>>1)&0x7FFF | (bit<<15)&0x8000
		}
		return []byte{byte(val & 0x00ff), byte(val & 0xff00 >> 8)}, nil
	}
	return nil
}
