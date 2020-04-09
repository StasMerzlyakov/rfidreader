package iso14443

import (
	"bytes"
	"fmt"
	"log"
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

	// interupt timeout
	INTERUPT_TIMEOUT = 5 * time.Millisecond
)

type PICC_TYPE = int

type UID struct {
	Uid     []byte // type A  (Unique Identifier)
	Sak     byte   // The SAK (Select acknowledge) byte returned from the PICC after successful selection
	PicType PICC_TYPE
}

// ISO 14443 methods
type PCDDevice interface {

	/**
	 * Communicate with PICC.
	 */
	PCD_CommunicateWithPICC(dataToSend []byte,
		validBits *byte,
		duration time.Duration) (
		result []byte,
		err error)

	/**
	 * Check collison
	 */
	PCD_IsCollisionOccure() (bool, error)

	/**
	 * Calculate a CRC_A
	 */
	PCD_CalculateCRC(buffer []byte, duration time.Duration) ([]byte, error)

	/**
	 * Initializes chip.
	 */
	PCD_Init() error

	/**
	 * Performs a device reset
	 */
	PCD_Reset() error

	/**
	 * Performs a self-test if exists
	 */
	PCD_PerformSelfTest() error

	/**
	 *  Turns the antenna
	 */
	PCD_AntennaOn() error

	/**
	 *  Turns the antenna off
	 */
	PCD_AntenaOff() error
}

type ISO14443Driver struct {
	device PCDDevice
}

func NewISO14443Driver(device PCDDevice) *ISO14443Driver {
	return &ISO14443Driver{device: device}
}

/**
 * Initialize physical device for scanning
 */
func (r *ISO14443Driver) PCD_Init() error {
	return r.device.PCD_Init()

}

/**
 * Initialize physical device for scanning
 */
func (r *ISO14443Driver) StartScan() error {
	return r.device.PCD_AntenaOff()
}

/**
 * Deinitialize physical scan
 */
func (r *ISO14443Driver) StopScan() error {
	return r.device.PCD_AntennaOn()
}

/**
 */
func (r *ISO14443Driver) PICC_RequestA() ([]byte, error) {
	validBits := byte(7)
	return r.device.PCD_CommunicateWithPICC([]byte{PICC_CMD_REQA}, &validBits, INTERUPT_TIMEOUT)
}

/**
 */
func (r *ISO14443Driver) PICC_RequestWUPA() ([]byte, error) {
	validBits := byte(7)
	return r.device.PCD_CommunicateWithPICC([]byte{PICC_CMD_WUPA}, &validBits, INTERUPT_TIMEOUT)
}

/**
 * Returns true if a PICC responds to PICC_CMD_REQA.
 * Only "new" cards in state IDLE are invited. Sleeping cards in state HALT are ignored.
 */
func (r *ISO14443Driver) PICC_IsNewCardPresent() bool {

	// Reset baud rates
	res, err := r.PICC_RequestA()
	log.Printf("PICC_RequestA:  len(%d)\n", len(res))
	if err != nil {
		log.Printf("PICC_RequestA: %s\n", err.Error())
	}
	return len(res) == 2
}

/**
 * Anticollision cycle ISO/IEC 14443-3:2011
 */
func (r *ISO14443Driver) selectLevel(clevel int /* Cascade level */, duration time.Duration) (uid []byte, sak byte, err error) {

	log.Printf(" selectLevel %d\n", clevel)

	var selByte byte

	switch clevel {
	case 1:
		selByte = PICC_CMD_SEL_CL1
	case 2:
		selByte = PICC_CMD_SEL_CL2
	case 3:
		selByte = PICC_CMD_SEL_CL3
	default:
		err = CommonError(fmt.Sprintf("Wrong cascade level %d\n", clevel))
	}

	nvb := byte(0x20)

	dataToSend := []byte{selByte, nvb}

	validBits := byte(0)

	var result []byte
	log.Printf("  send [% x]\n", dataToSend)
	if result, err = r.device.PCD_CommunicateWithPICC(dataToSend, &validBits, duration); err != nil {
		return
	}

	if len(result) != 5 {
		// UIDcl + BC
		err = CommonError(fmt.Sprintf("Unexpected result length: level %d, len(result): %d\n", clevel, len(result)))
		return
	}

	// Check collision
	var collOccr bool
	if collOccr, err = r.device.PCD_IsCollisionOccure(); err != nil {
		return
	} else {
		if !collOccr { // CollErr is 0!!
			log.Printf(" CollErr is 0\n")
			log.Printf("   validBits: %d\n", validBits)
			var crc_a []byte
			nvb = byte(0x70)
			// Calculate CRC
			dataToSend = append([]byte{selByte, nvb}, result...)
			if crc_a, err = r.device.PCD_CalculateCRC(dataToSend, INTERUPT_TIMEOUT); err != nil {
				return
			}

			uid = result[:4]
			dataToSend = append(dataToSend, crc_a...)
			log.Printf("  send [% x]\n", dataToSend)
			if result, err = r.device.PCD_CommunicateWithPICC(dataToSend, &validBits, duration); err != nil {
				return
			}
			if len(result) != 3 { // SAK must be exactly 24 bits (1 byte + CRC_A)
				err = CommonError(fmt.Sprintf("SAK must be exactly 24 bits (1 byte + CRC_A). Received %d\n", len(result)))
				return
			}

			var crcRes []byte
			if crcRes, err = r.device.PCD_CalculateCRC(result[:1], duration); err != nil {
				return
			}
			if bytes.Compare(crcRes, result[1:]) != 0 {
				err = CommonError(fmt.Sprintf("CRC check SAK CRC_A error: \n"+
					"calucated: [% x]\n received [% x]\n", crcRes, result[:2]))
				return
			}

			sak = result[0]
			return
		} else {
			err = CommonError("COLLISION CYCLE NOT SUPPORTED YET\n")
		}
	}

	return
}

/**
 * Initialization and anticollision cycle ISO/IEC 14443-3:2011
 */
func (r *ISO14443Driver) PICC_Select() (uid *UID, err error) {
	// Expected that RequestA sended by method PICC_IsNewCardPresent

	level := 1
	var sak byte
	var buffer []byte
	uidVal := []byte{}

	for {
		if buffer, sak, err = r.selectLevel(level, INTERUPT_TIMEOUT); err != nil {
			return nil, err
		}

		if buffer[0] == PICC_CMD_CT { // skip Cascade Tag
			uidVal = append(uidVal, buffer[1:]...)
		} else {
			uidVal = append(uidVal, buffer[:]...)
		}
		// Check SAK
		if sak&0x04 == 0 { // UID complete
			uid = &UID{Uid: uidVal, Sak: sak}

			switch sak & 0x7F {
			case 0x04:
				uid.PicType = PICC_TYPE_NOT_COMPLETE // UID not complete
			case 0x09:
				uid.PicType = PICC_TYPE_MIFARE_MINI
			case 0x08:
				uid.PicType = PICC_TYPE_MIFARE_1K
			case 0x18:
				uid.PicType = PICC_TYPE_MIFARE_4K
			case 0x00:
				uid.PicType = PICC_TYPE_MIFARE_UL
			case 0x10:
			case 0x11:
				uid.PicType = PICC_TYPE_MIFARE_PLUS
			case 0x01:
				uid.PicType = PICC_TYPE_TNP3XXX
			case 0x20:
				uid.PicType = PICC_TYPE_ISO_14443_4
			case 0x40:
				uid.PicType = PICC_TYPE_ISO_18092
			default:
				uid.PicType = PICC_TYPE_UNKNOWN
			}

			break
		}
		log.Printf(" ------ Level: %d. UID is not complete. SAK: %08b\n", level, sak)

		level++
	}
	return
}
