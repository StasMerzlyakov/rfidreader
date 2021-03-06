package mfrc522

import (
	"time"
)

// A struct used for passing the UID of a PIC
const (
	PICC_TYPE_UNKNOWN = iota

	PICC_TYPE_ISO_14443_4           // PICC compliant with ISO/IEC 14443-4
	PICC_TYPE_ISO_18092             // PICC compliant with ISO/IEC 18092 (NFC)
	PICC_TYPE_MIFARE_MINI           // MIFARE Classic protocol, 320 bytes
	PICC_TYPE_MIFARE_1K             // MIFARE Classic protocol, 1KB
	PICC_TYPE_MIFARE_4K             // MIFARE Classic protocol, 4KB
	PICC_TYPE_MIFARE_UL             // MIFARE Ultralight or Ultralight C
	PICC_TYPE_MIFARE_PLUS           // MIFARE Plus
	PICC_TYPE_MIFARE_DESFIRE        // MIFARE DESFire
	PICC_TYPE_TNP3XXX               // Only mentioned in NXP AN 10833 MIFARE Type Identification Procedure
	PICC_TYPE_NOT_COMPLETE   = 0xff // SAK indicates UID is not complete.

	// ISO14443-3 Commands
	PICC_CMD_REQA    = 0x26 // REQuest command, Type A. Invites PICCs in state IDLE to go to READY and prepare for anticollision or selection. 7 bit frame.
	PICC_CMD_WUPA    = 0x52 // Wake-UP command, Type A. Invites PICCs in state IDLE and HALT to go to READY(*) and prepare for anticollision or selection. 7 bit frame.
	PICC_CMD_HLTA    = 0x50 // HaLT command, Type A. Instructs an ACTIVE PICC to go to state HALT.
	PICC_CMD_CT      = 0x88 // Cascade Tag. Not really a command, but used during anti collision.
	PICC_CMD_SEL_CL1 = 0x93 // Anti collision/Select, Cascade Level 1
	PICC_CMD_SEL_CL2 = 0x95 // Anti collision/Select, Cascade Level 2
	PICC_CMD_SEL_CL3 = 0x97 // Anti collision/Select, Cascade Level 3

	ISO_14443_CRC_RESET = 0x6363

	// interupt timeout
	INTERUPT_TIMEOUT = 5 * time.Millisecond
)

type PICC_TYPE = int

type UID struct {
	Uid     []byte // type A  (Unique Identifier)
	Sak     byte   // The SAK (Select acknowledge) byte returned from the PICC after successful selection
	PicType PICC_TYPE
}

// Calculate an ISO 14443a CRC. Code translated from the code in
// iso14443a_crc().
func ISO14443aCRC(data []byte) []byte {
	crc := uint32(0x6363)
	for _, bt := range data {
		bt ^= uint8(crc & 0xff)
		bt ^= bt << 4
		bt32 := uint32(bt)
		crc = (crc >> 8) ^ (bt32 << 8) ^ (bt32 << 3) ^ (bt32 >> 4)
	}
	return []byte{byte(crc & 0xff), byte((crc >> 8) & 0xff)}
}
