// Contain MIFARE PICC manipulate funcitons

package mfrc522

import (
	"bytes"
	_ "log"
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

// Suc function
type SucFn func() ([]byte, error)

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

func InitSuc(init []byte) SucFn {

	// BIT ORDER relative implementation
	// Little endian !!!

	var state = uint32(init[0]) | uint32(init[1])<<8 | uint32(init[2])<<16 | uint32(init[3])<<24

	return func() ([]byte, error) {
		for i := 0; i < 32; i++ {
			bit := (state & (0x1 << 21) >> 21) ^ (state & (0x1 << 19) >> 19) ^
				(state & (0x1 << 18) >> 18) ^ (state & (0x1 << 16) >> 16)
			state = (state>>1)&0x7FFFFFFF | (bit<<31)&0x80000000
		}

		return []byte{byte(state & 0xFF), byte(state & (0xFF << 8) >> 8),
			byte(state & (0xFF << 16) >> 16), byte(state & (0xFF << 24) >> 24)}, nil
	}
}

/**
  MIFARE LSFR32 encryptor
 	uid : uid0, uid1, uid2, uid3
	nt : nt0, nt1, nt2, nt3
	key: key0, key1, key2, key3, key4, key5
*/

func f4(rcode uint16, x0, x1, x2, x3 byte) byte {
	uid := uint16((x3&1)<<3) | uint16((x2&1)<<2) | uint16((x1&1)<<1) | uint16(x0&1)
	return byte(((rcode & (uint16(1) << uid)) >> uid) & 1)
}

func f5(rcode uint32, x0, x1, x2, x3, x4 byte) byte {
	uid := uint32((x4&1)<<4) | uint32((x3&1)<<3) | uint32((x2&1)<<2) | uint32((x1&1)<<1) | uint32(x0&1)
	return byte(((rcode & (uint32(1) << uid)) >> uid) & 1)
}

//
// Dismantling_mifare_classic.pdf
// MIFARE_CRYPTO.pdf
//
func Fc(y0, y1, y2, y3, y4 byte) byte {
	// Fc (y0∨((y1 ∨y4)∧(y3⊕y4)))⊕((y0⊕(y1∧y3))∧((y2⊕y3)∨(y1∧y4)))
	return f5(0xec57e80a, y0, y1, y2, y3, y4)
	//return f5(0x4457c3b3, y0, y1, y2, y3, y4)
}

func Fa(y0, y1, y2, y3 byte) byte {
	// Fa ((y0∨y1)⊕(y0 ∧y3))⊕(y2∧((y0⊕y1)∨y3))
	return f4(0xb48e, y0, y1, y2, y3)
	//return f4(0x26c7, y0, y1, y2, y3)
}

func Fb(y0, y1, y2, y3 byte) byte {
	// Fb ((y0∧y1)∨y2)⊕((y0⊕y1)∧(y2∨y3))
	return f4(0x9e98, y0, y1, y2, y3)
	//return f4(0x0dd3^0xFFFF, y0, y1, y2, y3)

}

func InitLfsr32FN(key []byte) Lfsr32FN {

	// BIT ORDER relative implementation
	// Little endian !!!
	var state = uint64(key[0]) | uint64(key[1])<<8 | uint64(key[2])<<16 |
		uint64(key[3])<<24 | uint64(key[4])<<32 | uint64(key[5])<<40

	var lsfrN = func() uint64 {
		return state&0x1 ^ (state & (0x1 << 5) >> 5) ^ (state & (0x1 << 9) >> 9) ^
			(state & (0x1 << 10) >> 10) ^ (state & (0x1 << 12) >> 12) ^ (state & (0x1 << 14) >> 14) ^
			(state & (0x1 << 15) >> 15) ^ (state & (0x1 << 17) >> 17) ^ (state & (0x1 << 19) >> 19) ^
			(state & (0x1 << 24) >> 24) ^ (state & (0x1 << 25) >> 25) ^ (state & (0x1 << 27) >> 27) ^
			(state & (0x1 << 29) >> 29) ^ (state & (0x1 << 35) >> 35) ^ (state & (0x1 << 39) >> 39) ^
			(state & (0x1 << 41) >> 41) ^ (state & (0x1 << 42) >> 42) ^ (state & (0x1 << 43) >> 43)
	}

	var f_0 = func() byte {
		y0 := byte(state & (0x1 << 9) >> 9)
		y1 := byte(state & (0x1 << 11) >> 11)
		y2 := byte(state & (0x1 << 13) >> 13)
		y3 := byte(state & (0x1 << 15) >> 15)
		return Fa(y0, y1, y2, y3)
	}

	var f_1 = func() byte {
		y0 := byte(state & (0x1 << 17) >> 17)
		y1 := byte(state & (0x1 << 19) >> 19)
		y2 := byte(state & (0x1 << 21) >> 21)
		y3 := byte(state & (0x1 << 23) >> 23)
		return Fb(y0, y1, y2, y3)
	}

	var f_2 = func() byte {
		y0 := byte(state & (0x1 << 25) >> 25)
		y1 := byte(state & (0x1 << 27) >> 27)
		y2 := byte(state & (0x1 << 29) >> 29)
		y3 := byte(state & (0x1 << 31) >> 31)
		return Fb(y0, y1, y2, y3)
	}

	var f_3 = func() byte {
		y0 := byte(state & (0x1 << 33) >> 33)
		y1 := byte(state & (0x1 << 35) >> 35)
		y2 := byte(state & (0x1 << 37) >> 37)
		y3 := byte(state & (0x1 << 39) >> 39)
		return Fa(y0, y1, y2, y3)
	}

	var f_4 = func() byte {
		y0 := byte(state & (0x1 << 41) >> 41)
		y1 := byte(state & (0x1 << 43) >> 43)
		y2 := byte(state & (0x1 << 45) >> 45)
		y3 := byte(state & (0x1 << 47) >> 47)
		return Fb(y0, y1, y2, y3)
	}

	var f = func() byte {
		return Fc(f_0(), f_1(), f_2(), f_3(), f_4())
	}

	var round = 0

	return func(input []byte) ([]byte, error) {

		var result []byte
		if input == nil {
			result = make([]byte, 4)
		} else {
			result = make([]byte, len(input))
		}

		// кол-во раундов
		rounds := 8 * len(result)

		if round <= 63 {
			if len(input) != 4 {
				return nil, UsageError("Unexpected length")
			}

			init := uint32(input[0]) | uint32(input[1])<<8 |
				uint32(input[2])<<16 | uint32(input[3]<<24)

			for i := 0; i < rounds; i++ {
				fn := f()
				bit := lsfrN()&0x1 ^ uint64(init&0x1)
				init = init >> 1
				state = state>>1 | bit<<47
				round += 1
				cell := i / 8
				pos := i % 8
				result[cell] = result[cell] | (fn & 0x1 << pos)
			}
		} else {
			// Столько раундов, сколько длина входного массива
			for i := 0; i < rounds; i++ {
				fn := f()
				bit := lsfrN() & 0x1
				state = state>>1 | bit<<47
				round += 1
				cell := i / 8
				pos := i % 8
				result[cell] = result[cell] | (fn & 0x1 << pos)
			}
		}
		return result, nil
	}

}
