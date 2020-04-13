// Contain MIFARE PICC manipulate funcitons

package mifare

import (
	"bytes"
	"rfidreader/mfrc522"
	"time"
)

type MFRC522Device interface {
	PCD_CalculateCRC(crcResetValue int, buffer []byte, duration time.Duration) ([]byte, error)
}

// NUID generation
// https://www.nxp.com/docs/en/application-note/AN10927.pdf
// Reset the CRC calculator with the standard ISO/IEC 14443 type A preset values: 6363hex first
func GenerateNUID(uid mfrc522.UID, device MFRC522Device) (nuid []byte, err error) {

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
		if crc, err = device.PCD_CalculateCRC(mfrc522.ISO_14443_CRC_RESET, uid.Uid[:3], mfrc522.INTERUPT_TIMEOUT); err != nil {
			return
		}
		buff.WriteByte((crc[0] | 0x0f) & 0xff)
		buff.WriteByte(crc[1])

		if crc, err = device.PCD_CalculateCRC(mfrc522.ISO_14443_CRC_RESET, uid.Uid[3:], mfrc522.INTERUPT_TIMEOUT); err != nil {
			return
		}
		buff.Write(crc)
		nuid = buff.Bytes()
	default:
		err = mfrc522.CommonError("Wrong uid lenght: %d. Expected 4 or 7")
	}
	return
}
