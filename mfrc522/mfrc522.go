// Dev is an handle to an MFRC522 RFID reader.
// Rewrite periph.io/x/periph/experimental/devices/mfrc522
// based on git@github.com:paguz/RPi-RFID.git
// Package mfrc522 controls a Mifare RFID card reader.
// Datasheet
// https://www.nxp.com/docs/en/data-sheet/MFRC522.pdf
//
package mfrc522

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"rfidreader/iso14443"
	"time"

	"periph.io/x/periph/conn/gpio"
	_ "periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/spi"
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

const (
	// MFRC522 v1
	VER_1_0 = 0x91
	// MFRC522 v2
	VER_2_0 = 0x92

	// MFRC522 commands
	PCD_Idle             = 0x00 // no action, cancels current command execution
	PCD_Mem              = 0x01 // stores 25 bytes into the internal buffer
	PCD_GenerateRandomID = 0x02 // generates a 10-byte random ID number
	PCD_CalcCRC          = 0x03 // activates the CRC coprocessor or performs a self test
	PCD_Transmit         = 0x04 // transmits data from the FIFO buffer
	PCD_NoCmdChange      = 0x07 // no command change, can be used to modify the CommandReg register bits without affecting the command, for example, the PowerDown bit
	PCD_Receive          = 0x08 // activates the receiver circuits
	PCD_Transceive       = 0x0C // transmits data from FIFO buffer to antenna and automatically activates the receiver after transmission
	PCD_MFAuthent        = 0x0E // performs the MIFARE standard authentication as a reader
	PCD_SoftReset        = 0x0F // resets the MFRC522

	// Page 0: Command and status
	//						  0x00	// reserved for future use
	CommandReg    = 0x01 // starts and stops command execution
	ComIEnReg     = 0x02 // enable and disable interrupt request control bits
	DivIEnReg     = 0x03 // enable and disable interrupt request control bits
	ComIrqReg     = 0x04 // interrupt request bits
	DivIrqReg     = 0x05 // interrupt request bits
	ErrorReg      = 0x06 // error bits showing the error status of the last command executed
	Status1Reg    = 0x07 // communication status bits
	Status2Reg    = 0x08 // receiver and transmitter status bits
	FIFODataReg   = 0x09 // input and output of 64 byte FIFO buffer
	FIFOLevelReg  = 0x0A // number of bytes stored in the FIFO buffer
	WaterLevelReg = 0x0B // level for FIFO underflow and overflow warning
	ControlReg    = 0x0C // miscellaneous control registers
	BitFramingReg = 0x0D // adjustments for bit-oriented frames
	CollReg       = 0x0E // bit position of the first bit-collision detected on the RF interface
	//						  0x0F		// reserved for future use

	// Page 1: Command
	// 						  0x10	// reserved for future use
	ModeReg        = 0x11 // defines general modes for transmitting and receiving
	TxModeReg      = 0x12 // defines transmission data rate and framing
	RxModeReg      = 0x13 // defines reception data rate and framing
	TxControlReg   = 0x14 // controls the logical behavior of the antenna driver pins TX1 and TX2
	TxASKReg       = 0x15 // controls the setting of the transmission modulation
	TxSelReg       = 0x16 // selects the internal sources for the antenna driver
	RxSelReg       = 0x17 // selects internal receiver settings
	RxThresholdReg = 0x18 // selects thresholds for the bit decoder
	DemodReg       = 0x19 // defines demodulator settings
	// 						  0x1A	// reserved for future use
	// 						  0x1B	// reserved for future use
	MfTxReg = 0x1C // controls some MIFARE communication transmit parameters
	MfRxReg = 0x1D // controls some MIFARE communication receive parameters
	// 						  0x1E	// reserved for future use
	SerialSpeedReg = 0x1F // selects the speed of the serial UART interface

	// Page 2: Configuration
	// 						  0x20	// reserved for future use
	CRCResultRegH = 0x21 // shows the MSB and LSB values of the CRC calculation
	CRCResultRegL = 0x22
	// 						  0x23	// reserved for future use
	ModWidthReg = 0x24 // controls the ModWidth setting?
	// 						  0x25	// reserved for future use
	RFCfgReg          = 0x26 // configures the receiver gain
	GsNReg            = 0x27 // selects the conductance of the antenna driver pins TX1 and TX2 for modulation
	CWGsPReg          = 0x28 // defines the conductance of the p-driver output during periods of no modulation
	ModGsPReg         = 0x29 // defines the conductance of the p-driver output during periods of modulation
	TModeReg          = 0x2A // defines settings for the internal timer
	TPrescalerReg     = 0x2B // the lower 8 bits of the TPrescaler value. The 4 high bits are in TModeReg.
	TReloadRegH       = 0x2C // defines the 16-bit timer reload value
	TReloadRegL       = 0x2D
	TCounterValueRegH = 0x2E // shows the 16-bit timer value
	TCounterValueRegL = 0x2F

	// Page 3: Test Registers
	// 						  0x30			// reserved for future use
	TestSel1Reg     = 0x31 // general test signal configuration
	TestSel2Reg     = 0x32 // general test signal configuration
	TestPinEnReg    = 0x33 // enables pin output driver on pins D1 to D7
	TestPinValueReg = 0x34 // defines the values for D1 to D7 when it is used as an I/O bus
	TestBusReg      = 0x35 // shows the status of the internal test bus
	AutoTestReg     = 0x36 // controls the digital self test
	VersionReg      = 0x37 // shows the software version
	AnalogTestReg   = 0x38 // controls the pins AUX1 and AUX2
	TestDAC1Reg     = 0x39 // defines the test value for TestDAC1
	TestDAC2Reg     = 0x3A // defines the test value for TestDAC2
	TestADCReg      = 0x3B // shows the value of ADC I and Q channels
	// 						  0x3C			// reserved for production tests
	// 						  0x3D			// reserved for production tests
	// 						  0x3E			// reserved for production tests
	// 						  0x3F			// reserved for production tests

)

type MFRC522 struct {
	iso14443.ISO14443Driver
	spiDev spi.Conn
	//operationTimeout time.Duration
	//	beforeCall       func()
	//afterCall        func()
	resetPin gpio.PinOut
	irqPin   gpio.PinIn
	//antennaGain int
}

type IRQCallbackFn func()

func NewMFRC522(spiPort spi.Port, resetPin gpio.PinOut, irqPin gpio.PinIn) (iso14443.PCDDevice, error) {

	if resetPin == nil {
		return nil, iso14443.CommonError("Reset pin is not set")
	}

	if irqPin == nil {
		return nil, iso14443.CommonError("IRQ pin is not set")
	}

	spiDev, err := spiPort.Connect(10*physic.MegaHertz, spi.Mode0, 8)
	if err != nil {
		return nil, err
	}
	if err := resetPin.Out(gpio.High); err != nil {
		return nil, err
	}

	if err := irqPin.In(gpio.PullUp, gpio.NoEdge); err != nil {
		return nil, err
	}

	reader := &MFRC522{
		spiDev:   spiDev,
		resetPin: resetPin,
		irqPin:   irqPin,
	}

	reader.PCD_Reset()
	return reader, nil
}

/////////////////////////////////////////////////////////////////////////////////////
// Basic interface functions for communicating with the MFRC522
/////////////////////////////////////////////////////////////////////////////////////

/**
 * Writes a byte to the specified register in the MFRC522 chip.
 * The interface is described in the datasheet section 8.1.2.
 */
func (r *MFRC522) PCD_WriteRegister(address, value byte) error {
	newData := []byte{(byte(address) << 1) & 0x7E, value}
	err := r.spiDev.Tx(newData, nil)
	return err
}

/**
 * Writes a number of bytes to the specified register in the MFRC522 chip.
 * The interface is described in the datasheet section 8.1.2.
 */
func (r *MFRC522) PCD_WriteFIFOBuffer(value []byte) error {
	for _, val := range value {
		if err := r.PCD_WriteRegister(FIFODataReg, val); err != nil {
			return err
		}
	}
	return nil
}

/**
 * Reads a number of bytes from the specified register in the MFRC522 chip.
 * The interface is described in the datasheet section 8.1.2.
 */
func (r *MFRC522) PCD_ReadRegister(address byte) (byte, error) {
	data := []byte{((byte(address) << 1) & 0x7E) | 0x80, 0}
	out := make([]byte, len(data))
	if err := r.spiDev.Tx(data, out); err != nil {
		return 0, err
	}
	return out[1], nil
}

/**
 * Reads a number of bytes from the specified register in the MFRC522 chip.
 * The interface is described in the datasheet section 8.1.2.
 */

func (r *MFRC522) PCD_ReadFIFOBuffer(count int) ([]byte, error) {
	result := make([]byte, count)
	for ind := range result {
		if val, err := r.PCD_ReadRegister(FIFODataReg); err != nil {
			return nil, err
		} else {
			result[ind] = val
		}
	}
	return result, nil
}

/**
 * Clears the bits given in mask from register reg.
 */
func (r *MFRC522) PCD_ClearRegisterBitMask(reg, mask byte) error {
	if current, err := r.PCD_ReadRegister(reg); err != nil {
		return err
	} else {
		return r.PCD_WriteRegister(reg, current&^mask) // clear bit mask
	}
} // End PCD_ClearRegisterBitMask()

/**
 * Sets the bits given in mask in register reg.
 */
func (r *MFRC522) PCD_SetRegisterBitMask(reg, mask byte) error {
	if tmp, err := r.PCD_ReadRegister(reg); err != nil {
		return err
	} else {
		return r.PCD_WriteRegister(reg, tmp|mask) // set bit mask
	}

} // End PCD_SetRegisterBitMask()

/**
 */
func (r *MFRC522) PCD_IsCollisionOccure() (bool, error) {
	if anyByte, err := r.PCD_ReadRegister(ErrorReg); err != nil {
		return false, err
	} else {
		if anyByte&0x08 == 0 { // CollErr is 0!!
			return false, nil
		}
	}
	return true, nil
}

/**
 * Communicate with PICC.
 */
func (r *MFRC522) PCD_CommunicateWithPICC(dataToSend []byte,
	validBits *byte,
	duration time.Duration) (
	result []byte,
	err error) {

	// Clear collision registr
	r.PCD_ClearRegisterBitMask(CollReg, 0x80)

	// Stop all operations
	if err = r.PCD_WriteRegister(CommandReg, PCD_Idle); err != nil {
		return
	}

	///////////////////////////////////////////////
	//// Write data
	///////////////////////////////////////////////
	// Clear FIFO biffer
	if err = r.PCD_SetRegisterBitMask(FIFOLevelReg, 0x80); err != nil {
		return
	}

	// Write data
	if err = r.PCD_WriteFIFOBuffer(dataToSend); err != nil {
		return
	}

	// Prepare values for BitFramingReg
	bitFraming := *validBits & 0x07
	r.PCD_WriteRegister(BitFramingReg, bitFraming)

	///////////////////////////////////////////////
	//// Transmite data
	///////////////////////////////////////////////

	if err = r.PCD_WriteRegister(CommandReg, PCD_Transceive); err != nil {
		return
	}
	if err = r.PCD_SetRegisterBitMask(BitFramingReg, 0x80); err != nil {
		return
	}

	// Whait PICC
	time.Sleep(duration)

	// check Irq flag
	var irqFlag byte
	if irqFlag, err = r.PCD_ReadRegister(ComIrqReg); err != nil {
		return
	} else {
		log.Printf("ComIrqReg: %08b\n", irqFlag)
		if irqFlag&0x30 == 0 {
			switch {
			case irqFlag&0x01 > 0:
				err = TimeoutIRqError("Response not completed\n")
				return

			case irqFlag&0x02 > 0:
				// Ercontactless UART an error is detected
				var errBit byte
				if errBit, err = r.PCD_ReadRegister(ErrorReg); err != nil {
					return
				} else {
					if errBit&0x13 > 0 { // BufferOvfl ParityErr ProtocolErr
						err = ErrIRqError(fmt.Sprintf("Contactless UART error is detected. ErrReg: %08b\n", errBit))
						return
					}
				}
			default:
				err = UnexpectedIRqError(fmt.Sprintf("Unexpected Irq Value. ComIrqReg: %x\n", irqFlag))
				return
			}
		}
	}

	// A received data stream ends
	var count byte
	if count, err = r.PCD_ReadRegister(FIFOLevelReg); err != nil {
		return
	}

	log.Printf("FIFOLevelReg : %08b\n", count)

	// TODO add rxAlign
	if result, err = r.PCD_ReadFIFOBuffer(int(count)); err != nil {
		return
	}

	if *validBits > 0 {
		var rValidBits byte
		if rValidBits, err = r.PCD_ReadRegister(ControlReg); err != nil {
			return
		}
		*validBits = rValidBits & 0x7
	}

	return

}

/**
 * Use the CRC coprocessor in the MFRC522 to calculate a CRC_A.
 * @return Result is written to result[0..1], low byte first.
 */
func (r *MFRC522) PCD_CalculateCRC(buffer []byte, duration time.Duration) ([]byte, error) {

	// Stop any active command.
	if err := r.PCD_WriteRegister(CommandReg, PCD_Idle); err != nil {
		return nil, err
	}

	// Clear FIFO biffer
	if err := r.PCD_SetRegisterBitMask(FIFOLevelReg, 0x80); err != nil {
		return nil, err
	}

	// Start the calculation
	if err := r.PCD_WriteRegister(CommandReg, PCD_CalcCRC); err != nil {
		return nil, err
	}

	// Write only first 64 byte
	var arr []byte
	if 64 <= len(buffer) {
		arr = buffer[:64]
	} else {
		arr = buffer[:len(buffer)]
	}

	if err := r.PCD_WriteFIFOBuffer(arr); err != nil {
		return nil, err
	}

	time.Sleep(duration)

	if bit, err := r.PCD_ReadRegister(DivIrqReg); err != nil {
		return nil, err
	} else {
		if bit&0x04>>2 != 1 {
			return nil, UnexpectedIRqError(fmt.Sprintf(" CalcCRC command notended. DivIrqReg: %x\n", bit))
		}
	}

	// CRC completed
	// Stop any active command.
	if err := r.PCD_WriteRegister(CommandReg, PCD_Idle); err != nil {
		return nil, err
	}
	// Transfer the result from the registers to the result buffer
	result := make([]byte, 2)
	var err error
	result[0], err = r.PCD_ReadRegister(CRCResultRegL)
	if err != nil {
		return nil, err
	}
	result[1], err = r.PCD_ReadRegister(CRCResultRegH)
	if err != nil {
		return nil, err
	}
	return result, nil
} // End PCD_CalculateCRC()

/////////////////////////////////////////////////////////////////////////////////////
// Functions for manipulating the MFRC522
/////////////////////////////////////////////////////////////////////////////////////

/**
 * Performs a soft reset on the MFRC522 chip and waits for it to be ready again.
 */
func (r *MFRC522) PCD_Reset() error {

	r.resetPin.Out(gpio.Low)
	time.Sleep(50 * time.Microsecond)
	r.resetPin.Out(gpio.High)

	// Wait for the PowerDown bit in CommandReg to be cleared
	for i := 0; i < 3; i++ {
		// Section 8.8.2 in the datasheet says the oscillator start-up time is the
		// start up time of the crystal + 37,74ms. Let us be generous: 50ms.
		time.Sleep(50 * time.Millisecond)
		if val, err := r.PCD_ReadRegister(CommandReg); err != nil {
			return err
		} else {
			if val&(1<<4) == 0 {
				return nil
			}
		}

	}
	return errors.New("PowerDown bit not cleared")
} // End PCD_Reset()

/**
 * Turns the antenna on by enabling pins TX1 and TX2.
 * After a reset these pins are disabled.
 */
func (r *MFRC522) PCD_AntennaOn() error {
	value, err := r.PCD_ReadRegister(TxControlReg)
	if err != nil {
		return err
	}
	if (value & 0x03) != 0x03 {
		err = r.PCD_SetRegisterBitMask(TxControlReg, 0x03)
		if err != nil {
			return err
		}
		time.Sleep(iso14443.INTERUPT_TIMEOUT)
	}
	return nil
} // End PCD_AntennaOn()

/**
 * Turns the antenna off by disabling pins TX1 and TX2.
 */
func (r *MFRC522) PCD_AntennaOff() error {
	if value, err := r.PCD_ReadRegister(TxControlReg); err != nil {
		return err
	} else {
		if (value & 0x03) == 0x03 {
			if err := r.PCD_ClearRegisterBitMask(TxControlReg, 0x03); err != nil {
				return err
			}
			time.Sleep(iso14443.INTERUPT_TIMEOUT)
		}
	}
	return nil
} // End PCD_AntennaOff()

/**
 * Initializes the MFRC522 chip.
 */
func (r *MFRC522) PCD_Init() error { // TODO error processing

	// When communicating with a PICC we need a timeout if something goes wrong.
	// f_timer = 13.56 MHz / (2*TPreScaler+1) where TPreScaler = [TPrescaler_Hi:TPrescaler_Lo].
	// TPrescaler_Hi are the four low bits in TModeReg. TPrescaler_Lo is TPrescalerReg.
	r.PCD_WriteRegister(TModeReg, 0x80)      // TAuto=1; timer starts automatically at the end of the transmission in all communication modes at all speeds
	r.PCD_WriteRegister(TPrescalerReg, 0xA9) // TPreScaler = TModeReg[3..0]:TPrescalerReg, ie 0x0A9 = 169 => f_timer=40kHz, ie a timer period of 25ms.
	r.PCD_WriteRegister(TReloadRegH, 0x03)   // Reload timer with 0x3E8 = 1000, ie 25ms before timeout.
	r.PCD_WriteRegister(TReloadRegL, 0xE8)

	// Reset baud rates
	r.PCD_WriteRegister(TxModeReg, 0x00)
	r.PCD_WriteRegister(RxModeReg, 0x00)
	// Reset ModWidthReg
	r.PCD_WriteRegister(ModWidthReg, 0x26)

	r.PCD_WriteRegister(TxASKReg, 0x40) // Default 0x00. Force a 100 % ASK modulation independent of the ModGsPReg register setting
	r.PCD_WriteRegister(ModeReg, 0x3D)  // Default 0x3F. Set the preset value for the CRC coprocessor for the CalcCRC command to 0x6363 (ISO 14443-3 part 6.2.4)
	//r.PCD_AntennaOn()                   // Enable the antenna driver pins TX1 and TX2 (they were disabled by the reset)

	return nil
} // End PCD_Init()

/**
 * Get the current MFRC522 Receiver Gain (RxGain[2:0]) value.
 * See 9.3.3.6 / table 98 in http://www.nxp.com/documents/data_sheet/MFRC522.pdf
 * NOTE: Return value scrubbed with (0x07<<4)=01110000b as RCFfgReg may use reserved bits.
 *
 * @return Value of the RxGain, scrubbed to the 3 bits used.
 */
func (r *MFRC522) PCD_GetAntennaGain() (byte, error) {
	val, err := r.PCD_ReadRegister(RFCfgReg)
	if err != nil {
		return 0, err
	}
	return val & (0x07 << 4), nil
} // End PCD_GetAntennaGain()

/**
 * Set the MFRC522 Receiver Gain (RxGain) to value specified by given mask.
 * See 9.3.3.6 / table 98 in http://www.nxp.com/documents/data_sheet/MFRC522.pdf
 * NOTE: Given mask is scrubbed with (0x07<<4)=01110000b as RCFfgReg may use reserved bits.
 */
func (r *MFRC522) PCD_SetAntennaGain(mask byte) error {
	if val, err := r.PCD_GetAntennaGain(); err != nil {
		return err
	} else {
		if val != mask {
			// only bother if there is a change
			// clear needed to allow 000 pattern
			if er := r.PCD_ClearRegisterBitMask(RFCfgReg, (0x07 << 4)); er != nil {
				return er
			}
			if er := r.PCD_SetRegisterBitMask(RFCfgReg, mask&(0x07<<4)); er != nil {
				return er
			} // only set RxGain[2:0] bits
		}
	}
	return nil
} // End PCD_SetAntennaGain()

/**
 * Performs a self-test of the MFRC522
 * See 16.1.1 in http://www.nxp.com/documents/data_sheet/MFRC522.pdf
 *
 * @return Whether or not the test passed.
 */
func (r *MFRC522) PCD_PerformSelfTest() error {
	// This follows directly the steps outlined in 16.1.1
	// 1. Perform a soft reset.

	if err := r.PCD_Reset(); err != nil {
		return err
	}

	// 2. Clear the internal buffer by writing 25 bytes of 00h
	emptyBuf := make([]byte, 25)
	if err := r.PCD_SetRegisterBitMask(FIFOLevelReg, 0x80); err != nil { // flush the FIFO buffer
		return err
	}
	if err := r.PCD_WriteFIFOBuffer(emptyBuf); err != nil { // write 25 bytes of 00h to FIFO
		return err
	}
	if err := r.PCD_WriteRegister(CommandReg, PCD_Mem); err != nil { // transfer to internal buffer
		return err
	}
	// 3. Enable self-test
	if err := r.PCD_WriteRegister(AutoTestReg, 0x09); err != nil {
		return err
	}

	// 4. Write 00h to FIFO buffer
	if err := r.PCD_WriteRegister(FIFODataReg, 0x00); err != nil {
		return err
	}

	// 5. Start self-test by issuing the CalcCRC command
	if err := r.PCD_WriteRegister(CommandReg, PCD_CalcCRC); err != nil {
		return err
	}

	// 6. Wait for self-test to complete
	i := 0
	for i := 0; i < 0xFF; i++ {
		if n, err := r.PCD_ReadRegister(DivIrqReg); err != nil { // DivIrqReg[7..0] bits are: Set2 reserved reserved MfinActIRq reserved CRCIRq reserved reserved
			return err
		} else {
			if n&0x04 > 0 { // CRCIRq bit set - calculation done
				break
			}
		}
	}
	if i == 0xFF {
		return errors.New("MFRC522 self test error")
	}

	if err := r.PCD_WriteRegister(CommandReg, PCD_Idle); err != nil { // Stop calculating CRC for new content in the FIFO.
		return err
	}
	// 7. Read out resulting 64 bytes from the FIFO buffer.
	result, err := r.PCD_ReadFIFOBuffer(64)
	if err != nil {
		return err
	}

	// Auto self-test done
	// Reset AutoTestReg register to be 0 again. Required for normal operation.
	if err := r.PCD_WriteRegister(AutoTestReg, 0x00); err != nil {
		return err
	}

	// Determine firmware version (see section 9.3.4.8 in spec)
	version, err := r.PCD_ReadRegister(VersionReg)
	if err != nil {
		return err
	}
	var expected []byte
	switch version {
	case VER_1_0:
		expected = MFRC522_VER_1_0
	case VER_2_0:
		expected = MFRC522_VER_2_0
	default:
		return iso14443.CommonError(fmt.Sprintf("Can't read version: %x", version))
	}

	if bytes.Compare(result, expected) != 0 {
		return errors.New(fmt.Sprintf("MFRC522 Self test [ERROR]:\nexpected: [% x]\n   "+
			"actual: [% x]", expected, result))
	}
	return nil
} // End PCD_PerformSelfTest()