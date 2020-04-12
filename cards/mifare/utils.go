// Contain MIFARE PICC manipulate funcitons

package mifare

import (
	"bytes"
	"rfidreader/iso14443"
)

// NUID generation
// https://www.nxp.com/docs/en/application-note/AN10927.pdf
// Reset the CRC calculator with the standard ISO/IEC 14443 type A preset values: 6363hex first
func GenerateNUID(uid iso14443.UID, device iso14443.PCDDevice) (nuid []byte, err error) {

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
		if crc, err = device.PCD_CalculateCRC(iso14443.ISO_14443_CRC_RESET, uid.Uid[:3], iso14443.INTERUPT_TIMEOUT); err != nil {
			return
		}
		buff.WriteByte((crc[0] | 0x0f) & 0xff)
		buff.WriteByte(crc[1])

		if crc, err = device.PCD_CalculateCRC(iso14443.ISO_14443_CRC_RESET, uid.Uid[3:], iso14443.INTERUPT_TIMEOUT); err != nil {
			return
		}
		buff.Write(crc)
		nuid = buff.Bytes()
	default:
		err = iso14443.CommonError("Wrong uid lenght: %d. Expected 4 or 7")
	}
	return
}
