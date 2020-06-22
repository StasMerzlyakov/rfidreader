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
		//0x3bae: 0x03ed,
		0x0120: 0x0145,
		0x4ca3: 0xec7a,
		0x6876: 0x8c86,
		0x93a6: 0xd176,
		0x632e: 0x4481,
		0xe7a3: 0x7d92,
	}

	for k, v := range test_data {

		input := []byte{byte(k & 0xff00 >> 8), byte(k & 0xff)}
		output := []byte{byte(v & 0xff00 >> 8), byte(v & 0xff)}

		f16 := InitLfsr16FN(input)
		res1, _ := f16()
		is.True(bytes.Compare(res1, output) == 0)
	}
}

func TestAuthentificationAlg(t *testing.T) {
	is := is.New(t)

	key := []byte{0x09, 0x1e, 0x63, 0x9c, 0xb7, 0x15}

	/*
	  RDR<--->60 14 50 2d
	  TAG<--->ce 84 42 61
	  RDR<--->f8 04 9c cb 05 25 c8 4f
	  TAG<--->94 31 cc 40
	  RDR<--->70 93 df 99
	  TAG<--->99 72 42 8c e2 e8 52 3f 45 6b 99 c8 31 e7 69 dc ed 09
	  RDR<--->8c a6 82 7b
	  TAG<--->ab 79 7f d3 69 e8 b9 3a 86 77 6b 40 da e3 ef 68 6e fd
	  RDR<--->c3 c3 81 ba
	  TAG<--->49 e2 c9 de f4 86 8d 17 77 67 0e 58 4c 27 23 02 86 f4
	  RDR<--->fb dc d7 c1
	  TAG<--->4a bd 96 4b 07 d3 56 3a a0 66 ed 0a 2e ac 7f 63 12 bf
	  RDR<--->9f 91 49 ea
	*/

	// RDR<--->60 14 50 2d (auth(14) + crc)
	buffer := []byte{PICC_CMD_MF_AUTH_KEY_A, 0x14}
	crc := ISO14443aCRC(buffer)
	buffer = append(buffer, crc...)
	is.True(bytes.Compare(buffer, []byte{0x60, 0x14, 0x50, 0x2d}) == 0)

	// TAG<--->ce 84 42 61 (nT(tag nonce) )

	buffer = []byte{0xce, 0x84}
	f16 := InitLfsr16FN(key)
	res1, _ := f16()

	log.Printf("res1: [% x]\n", res1)
	
	
	
//	is.True(bytes.Compare(res1, []byte{0x42, 0x61}) == 0)

	// Инициализируем регистр линейного сдвига
	//lfsr32 := InitLfsr32FN(key)
	//ks1, _ := lfsr32(init)

}
