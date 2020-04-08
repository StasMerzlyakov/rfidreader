package main

import (
	"bytes"
	"log"
	"time"

	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/conn/spi/spireg"
	_ "periph.io/x/periph/experimental/devices/mfrc522"
	"periph.io/x/periph/experimental/devices/mfrc522/commands"
	"periph.io/x/periph/host"
)

const (
	PCD_MEM = 0x01
	ZERO    = 0x00
	VER_1_0 = 0x91
	VER_2_0 = 0x92
)

var MFRC522_VER_1_0 = []byte{0x00, 0xC6, 0x37, 0xD5, 0x32, 0xB7, 0x57, 0x5C,
	0xC2, 0xD8, 0x7C, 0x4D, 0xD9, 0x70, 0xC7, 0x73,
	0x10, 0xE6, 0xD2, 0xAA, 0x5E, 0xA1, 0x3E, 0x5A,
	0x14, 0xAF, 0x30, 0x61, 0xC9, 0x70, 0xDB, 0x2E,
	0x64, 0x22, 0x72, 0xB5, 0xBD, 0x65, 0xF4, 0xEC,
	0x22, 0xBC, 0xD3, 0x72, 0x35, 0xCD, 0xAA, 0x41,
	0x1F, 0xA7, 0xF3, 0x53, 0x14, 0xDE, 0x7E, 0x02,
	0xD9, 0x0F, 0xB5, 0x5E, 0x25, 0x1D, 0x29, 0x79}

var MFRC522_VER_2_0 = []byte{0x00, 0xEB, 0x66, 0xBA, 0x57, 0xBF, 0x23, 0x95,
	0xD0, 0xE3, 0x0D, 0x3D, 0x27, 0x89, 0x5C, 0xDE,
	0x9D, 0x3B, 0xA7, 0x00, 0x21, 0x5B, 0x89, 0x82,
	0x51, 0x3A, 0xEB, 0x02, 0x0C, 0xA5, 0x00, 0x49,
	0x7C, 0x84, 0x4D, 0xB3, 0xCC, 0xD2, 0x1B, 0x81,
	0x5D, 0x48, 0x76, 0xD5, 0x71, 0x61, 0x21, 0xA9,
	0x86, 0x96, 0x83, 0x38, 0xCF, 0x9D, 0x5B, 0x6D,
	0xDC, 0x15, 0xBA, 0x3E, 0x7D, 0x95, 0x3B, 0x2F}

func main() {

	// SELF TEST
	// 1. Perform a soft reset.
	// 2. Clear the internal buffer by writing 25 bytes of 00h and implement the Config command.
	// 3. Enable the self test by writing 09h to the AutoTestReg register.
	// 4. Write 00h to the FIFO buffer
	// 5. Start the self test with the CalcCRC command
	// 6. The self test is initiated.
	// 7. When the self test has completed, the FIFO buffer contains the following 64 bytes:

	// FIFO buffer byte values for MFRC522 version 1.0:
	// 00h, C6h, 37h, D5h, 32h, B7h, 57h, 5Ch,
	// C2h, D8h, 7Ch, 4Dh, D9h, 70h, C7h, 73h,
	// 10h, E6h, D2h, AAh, 5Eh, A1h, 3Eh, 5Ah,
	// 14h, AFh, 30h, 61h, C9h, 70h, DBh, 2Eh,
	// 64h, 22h, 72h, B5h, BDh, 65h, F4h, ECh,
	// 22h, BCh, D3h, 72h, 35h, CDh, AAh, 41h,
	// 1Fh, A7h, F3h, 53h, 14h, DEh, 7Eh, 02h,
	// D9h, 0Fh, B5h, 5Eh, 25h, 1Dh, 29h, 79h

	// FIFO buffer byte values for MFRC522 version 2.0:
	// 00h, EBh, 66h, BAh, 57h, BFh, 23h, 95h,
	// D0h, E3h, 0Dh, 3Dh, 27h, 89h, 5Ch, DEh,
	// 9Dh, 3Bh, A7h, 00h, 21h, 5Bh, 89h, 82h,
	// 51h, 3Ah, EBh, 02h, 0Ch, A5h, 00h, 49h,
	// 7Ch, 84h, 4Dh, B3h, CCh, D2h, 1Bh, 81h,
	// 5Dh, 48h, 76h, D5h, 71h, 61h, 21h, A9h,
	// 86h, 96h, 83h, 38h, CFh, 9Dh, 5Bh, 6Dh,
	// DCh, 15h, BAh, 3Eh, 7Dh, 95h, 3Bh, 2Fh
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}
	// Use spireg SPI port registry to find the first available SPI bus.
	spiPort, err := spireg.Open("")
	if err != nil {
		log.Fatal(err)
	}
	defer spiPort.Close()

	// Convert the spi.Port into a spi.Conn so it can be used for communication.
	// commands.LowLevel
	spiDev, err := spiPort.Connect(physic.MegaHertz, spi.Mode3, 8)
	if err != nil {
		log.Fatal(err)
	}

	// 1. Perform a soft reset.
	newData := []byte{(byte(commands.CommandReg) << 1) & 0x7E, commands.PCD_RESETPHASE}
	spiDev.Tx(newData, nil)
	if err := spiDev.Tx(newData, nil); err != nil {
		log.Fatal(err)
	}

	// The datasheet does not mention how long the SoftRest command takes to complete.
	// But the MFRC522 might have been in soft power-down mode (triggered by bit 4 of CommandReg)
	// Section 8.8.2 in the datasheet says the oscillator start-up time is the start up time of
	// the crystal + 37,74μs. Let us be generous: 50ms.

	count := 0
	for {
		time.Sleep(50 * time.Millisecond)
		count++
		data := []byte{((byte(commands.CommandReg) << 1) & 0x7E) | 0x80, ZERO}
		out := make([]byte, len(data))
		if err := spiDev.Tx(data, out); err != nil {
			log.Fatal(err)
		}
		if out[1]&(1<<4) == 0 {
			break
		}
		if count == 3 {
			log.Fatal("MFRC522 not returned")
		}
	}
	log.Printf("MFRC522 is ready")

	// 2. Clear the internal buffer by writing 25 bytes of 00h and implement the Config command.

	// Stop calculating CRC for new content in the FIFO
	newData = []byte{(byte(commands.CommandReg) << 1) & 0x7E, commands.PCD_IDLE}
	if err := spiDev.Tx(newData, nil); err != nil {
		log.Fatal(err)
	}

	newData = []byte{(byte(commands.FIFOLevelReg) << 1) & 0x7E, 0x80} // flush the FIFO buffer
	if err := spiDev.Tx(newData, nil); err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 25; i++ {
		newData = []byte{(byte(commands.FIFODataReg) << 1) & 0x7E, ZERO}
		if err := spiDev.Tx(newData, nil); err != nil {
			log.Fatal(err)
		}

	}

	newData = []byte{(byte(commands.CommandReg) << 1) & 0x7E, PCD_MEM}
	if err := spiDev.Tx(newData, nil); err != nil {
		log.Fatal(err)
	}
	log.Printf("MFRC522 25 bytes of 00h are writen to the internal buffer")

	// 3. Enable the self test by writing 09h to the AutoTestReg register.
	newData = []byte{(byte(commands.AutoTestReg) << 1) & 0x7E, 0x09}
	if err := spiDev.Tx(newData, nil); err != nil {
		log.Fatal(err)
	}

	// 4. Write 00h to the FIFO buffer
	newData = []byte{(byte(commands.FIFODataReg) << 1) & 0x7E, ZERO}
	if err := spiDev.Tx(newData, nil); err != nil {
		log.Fatal(err)
	}
	log.Printf("MFRC522 enable self test")

	// 5. Start the self test with the CalcCRC command.
	newData = []byte{(byte(commands.CommandReg) << 1) & 0x7E, commands.PCD_CALCCRC}
	if err := spiDev.Tx(newData, nil); err != nil {
		log.Fatal(err)
	}
	log.Printf("MFRC522 self test started")

	// 6. Wait for self-test to complete
	i := 0
	for {
		// The datasheet does not specify exact completion condition except
		// that FIFO buffer should contain 64 bytes.
		// While selftest is initiated by CalcCRC command
		// it behaves differently from normal CRC computation,
		// so one can't reliably use DivIrqReg to check for completion.
		// It is reported that some devices does not trigger CRCIRq flag
		// during selftest.

		data := []byte{((byte(commands.FIFOLevelReg) << 1) & 0x7E) | 0x80, ZERO}
		out := make([]byte, len(data))
		if err := spiDev.Tx(data, out); err != nil {
			log.Fatal(err)
		}

		if out[1] >= 64 {
			break
		}
		i++
		if i >= 0xFF {
			log.Fatalf("MFRC522 self test error %d", out[1])
		}
	}
	log.Printf("MFRC522 self test complete")

	// Stop calculating CRC for new content in the FIFO
	newData = []byte{(byte(commands.CommandReg) << 1) & 0x7E, commands.PCD_IDLE}
	if err := spiDev.Tx(newData, nil); err != nil {
		log.Fatal(err)
	}

	// READ BYTES HERE !!
	backData := make([]byte, 64)
	for i := 0; i < 64; i++ {
		data := []byte{((byte(commands.FIFODataReg) << 1) & 0x7E) | 0x80, ZERO}
		out := make([]byte, len(data))
		if err := spiDev.Tx(data, out); err != nil {
			log.Fatal(err)
		}

		backData[i] = out[1]
	}

	// Auto self-test done
	// Reset AutoTestReg register to be 0 again. Required for normal operation.
	newData = []byte{(byte(commands.AutoTestReg) << 1) & 0x7E, ZERO}
	if err := spiDev.Tx(newData, nil); err != nil {
		log.Fatal(err)
	}
	log.Printf("MFRC522 returned to normal state")
	log.Printf("read: [% x]\n", backData)

	data := []byte{((byte(commands.VersionReg) << 1) & 0x7E) | 0x80, ZERO}
	out := make([]byte, len(data))
	if err := spiDev.Tx(data, out); err != nil {
		log.Fatal(err)
	}
	var expected []byte
	var ver string
	switch version := out[1]; version {
	case VER_1_0:
		expected = MFRC522_VER_1_0
		ver = "1.0"
	case VER_2_0:
		expected = MFRC522_VER_2_0
		ver = "2.0"
	}

	if bytes.Compare(backData, expected) != 0 {
		log.Fatal("MFRC522 Self test [ERROR]")
	}

	log.Printf("MFRC522 ver %s Self test [OK]", ver)

}
