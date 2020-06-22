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
	_ "math/rand"
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

	CRC_RESET_VALUE_ZERO = 0
	CRC_RESET_VALUE_A671 = 0xa671
	CRC_RESET_VALUE_6363 = ISO_14443_CRC_RESET
	CRC_RESET_VALUE_FFFF = 0xffff

	PICC_CMD_MF_AUTH_KEY_A = 0x60 // Perform authentication with Key A
	PICC_CMD_MF_AUTH_KEY_B = 0x61 // Perform authentication with Key B
	PICC_CMD_MF_READ       = 0x30 // Reads one 16 byte block from the authenticated sector of the PICC. Also used for MIFARE Ultralight.
	PICC_CMD_MF_WRITE      = 0xA0 // Writes one 16 byte block to the authenticated sector of the PICC. Called "COMPATIBILITY WRITE" for MIFARE Ultralight.
	PICC_CMD_MF_DECREMENT  = 0xC0 // Decrements the contents of a block and stores the result in the internal data register.
	PICC_CMD_MF_INCREMENT  = 0xC1 // Increments the contents of a block and stores the result in the internal data register.
	PICC_CMD_MF_RESTORE    = 0xC2 // Reads the contents of a block into the internal data register.
	PICC_CMD_MF_TRANSFER   = 0xB0 // Writes the contents of the internal data register to a block.
)

type MFRC522 struct {
	spiDev spi.Conn
	//operationTimeout time.Duration
	//	beforeCall       func()
	//afterCall        func()
	resetPin gpio.PinOut
	irqPin   gpio.PinIn
	//antennaGain int
}

type IRQCallbackFn func()

func NewMFRC522(spiPort spi.Port, resetPin gpio.PinOut, irqPin gpio.PinIn) (*MFRC522, error) {

	if resetPin == nil {
		return nil, CommonError("Reset pin is not set")
	}

	if irqPin == nil {
		return nil, CommonError("IRQ pin is not set")
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
func (r *MFRC522) PCD_CommunicateWithPICC(command byte, dataToSend []byte,
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

	if err = r.PCD_WriteRegister(CommandReg, command); err != nil {
		return
	}
	if command == PCD_Transceive {
		if r.PCD_SetRegisterBitMask(BitFramingReg, 0x80); err != nil { // StartSend=1, transmission of data starts
			return
		}
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
func (r *MFRC522) PCD_CalculateCRC(crcResetValue int, buffer []byte, duration time.Duration) ([]byte, error) {

	switch crcResetValue {
	case CRC_RESET_VALUE_ZERO:
		r.PCD_WriteRegister(ModeReg, 0x00)
	case CRC_RESET_VALUE_6363:
		r.PCD_WriteRegister(ModeReg, 0x3d)
	case CRC_RESET_VALUE_A671:
		r.PCD_WriteRegister(ModeReg, 0x3e)
	case CRC_RESET_VALUE_FFFF:
		r.PCD_WriteRegister(ModeReg, 0x3f)
	default:
		return nil, CommonError(fmt.Sprintf("Unexpected crcResetValue: %x", crcResetValue))
	}

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
		time.Sleep(INTERUPT_TIMEOUT)
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
			time.Sleep(INTERUPT_TIMEOUT)
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
		return CommonError(fmt.Sprintf("Can't read version: %x", version))
	}

	if bytes.Compare(result, expected) != 0 {
		return errors.New(fmt.Sprintf("MFRC522 Self test [ERROR]:\nexpected: [% x]\n   "+
			"actual: [% x]", expected, result))
	}
	return nil
} // End PCD_PerformSelfTest()

/**
 * Returns true if a PICC responds to PICC_CMD_REQA.
 * Only "new" cards in state IDLE are invited. Sleeping cards in state HALT are ignored.
 */
func (r *MFRC522) PICC_IsNewCardPresent() bool {

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
func (r *MFRC522) selectLevel(clevel int /* Cascade level */, duration time.Duration) (uid []byte, sak byte, err error) {

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
	if result, err = r.PCD_CommunicateWithPICC(PCD_Transceive, dataToSend, &validBits, duration); err != nil {
		return
	}

	if len(result) != 5 {
		// UIDcl + BC
		err = CommonError(fmt.Sprintf("Unexpected result length: level %d, len(result): %d\n", clevel, len(result)))
		return
	}

	// Check collision
	var collOccr bool
	if collOccr, err = r.PCD_IsCollisionOccure(); err != nil {
		return
	} else {
		if !collOccr { // CollErr is 0!!
			log.Printf(" CollErr is 0\n")
			log.Printf("   validBits: %d\n", validBits)
			var crc_a []byte
			nvb = byte(0x70)
			// Calculate CRC
			dataToSend = append([]byte{selByte, nvb}, result...)
			if crc_a, err = r.PCD_CalculateCRC(ISO_14443_CRC_RESET, dataToSend, INTERUPT_TIMEOUT); err != nil {
				return
			}

			uid = result[:4]
			dataToSend = append(dataToSend, crc_a...)
			log.Printf("  send [% x]\n", dataToSend)
			if result, err = r.PCD_CommunicateWithPICC(PCD_Transceive, dataToSend, &validBits, duration); err != nil {
				return
			}
			if len(result) != 3 { // SAK must be exactly 24 bits (1 byte + CRC_A)
				err = CommonError(fmt.Sprintf("SAK must be exactly 24 bits (1 byte + CRC_A). Received %d\n", len(result)))
				return
			}

			var crcRes []byte
			if crcRes, err = r.PCD_CalculateCRC(ISO_14443_CRC_RESET, result[:1], duration); err != nil {
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
 * BIG ENDIAN !!!
 */
func (r *MFRC522) PICC_Select() (uid *UID, err error) {
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

/**
 */
func (r *MFRC522) PICC_RequestA() ([]byte, error) {
	validBits := byte(7)
	return r.PCD_CommunicateWithPICC(PCD_Transceive, []byte{PICC_CMD_REQA}, &validBits, INTERUPT_TIMEOUT)
}

/**
 */
func (r *MFRC522) PICC_RequestWUPA() ([]byte, error) {
	validBits := byte(7)
	return r.PCD_CommunicateWithPICC(PCD_Transceive, []byte{PICC_CMD_WUPA}, &validBits, INTERUPT_TIMEOUT)
}

/**
 */
func (r *MFRC522) PICC_Halt() error {

	buffer := []byte{PICC_CMD_HLTA, 0}

	// Calculate CRC_A
	if result, err := r.PCD_CalculateCRC(ISO_14443_CRC_RESET, buffer, INTERUPT_TIMEOUT); err != nil {
		return err
	} else {

		buffer = append(buffer, result...)
		validBits := byte(0)
		_, err := r.PCD_CommunicateWithPICC(PCD_Transceive, buffer, &validBits, INTERUPT_TIMEOUT)
		return err
	}
}

func (r *MFRC522) authentificateKey(keyCode byte, uid UID, key []byte, sector int) (err error) {
	buffer := []byte{keyCode, byte(sector)}
	crc := ISO14443aCRC(buffer)
	buffer = append(buffer, crc...)
	validBits := byte(0)
	var nt []byte
	if nt, err = r.PCD_CommunicateWithPICC(PCD_Transceive, buffer, &validBits, INTERUPT_TIMEOUT); err != nil {
		return
	}
	log.Printf("n_t: [% x]\n", nt)

	init := make([]byte, 4)
	init[0] = uid.Uid[0] ^ nt[0]
	init[1] = uid.Uid[1] ^ nt[1]
	init[2] = uid.Uid[2] ^ nt[2]
	init[3] = uid.Uid[3] ^ nt[3]

	// Инициализируем регистр линейного сдвига
	lfsr32 := InitLfsr32FN(key)

	// Генерируем ключ ks1
	ks1, _ := lfsr32(init)

	log.Printf("ks1 [% x]\n", ks1)

	// Генерируем nr
	nr := []byte{0x00, 0x00, 0x00, 0x00} //GenerateNR()

	// Формируем nr^ks1
	buffer = make([]byte, 8)
	buffer[0] = ks1[0] ^ nr[0]
	buffer[1] = ks1[1] ^ nr[1]
	buffer[2] = ks1[2] ^ nr[2]
	buffer[3] = ks1[3] ^ nr[3]

	// Формируем вторую часть
	ks2, _ := lfsr32(nr)

	suc := InitSuc(nt)
	// нужен suc2()
	suc()
	suc2, _ := suc()

	// Формируем suc2^ks2
	buffer[4] = ks2[0] ^ suc2[0]
	buffer[5] = ks2[1] ^ suc2[1]
	buffer[6] = ks2[2] ^ suc2[2]
	buffer[7] = ks2[3] ^ suc2[3]
	log.Printf("nr^ks1, suc2(nt)^ks2: [% x]\n", buffer)

	// Генерируем suc3^ks3
	suc3, _ := suc()

	ks3, _ := lfsr32(nil)
	expeted := make([]byte, 4)
	expeted[0] = suc3[0] ^ ks3[0]
	expeted[1] = suc3[1] ^ ks3[1]
	expeted[2] = suc3[2] ^ ks3[2]
	expeted[3] = suc3[3] ^ ks3[3]
	log.Printf("expected suc3^ks3: [% x]\n", expeted)

	// Отправляем
	var actual []byte
	if actual, err = r.PCD_CommunicateWithPICC(PCD_Transceive, buffer, &validBits, INTERUPT_TIMEOUT); err != nil {
		return
	}
	log.Printf("actual  suc3^ks3: [% x]\n", actual)
	if bytes.Compare(actual, expeted) != 0 {
		return AuthentificationError("Unexpected card result")
	}

	return nil
}

/*
func (r *MFRC522) authentificateKey(keyCode byte, uid UID, key []byte, sector int) (err error) {

	// Authentication command code (60h, 61h)
	// Block address
	// Sector key byte 0
	// Sector key byte 1
	// Sector key byte 2
	// Sector key byte 3
	// Sector key byte 4
	// Sector key byte 5
	// Card serial number byte 0
	// Card serial number byte 1
	// Card serial number byte 2
	// Card serial number byte 3

	buffer := []byte{keyCode, byte(sector)}
	buffer = append(buffer, key...)
	buffer = append(buffer, uid.Uid[:4]...)

	validBits := byte(0)
	if _, err := r.PCD_CommunicateWithPICC(PCD_MFAuthent, buffer, &validBits, INTERUPT_TIMEOUT); err != nil {
		return err
	}

	// check MFCrypto1On bit of Status2Reg
	if status2RegVal, err := r.PCD_ReadRegister(Status2Reg); err != nil {
		return err
	} else {
		log.Printf("Status2Reg: %08b\n", status2RegVal)
		if status2RegVal&0x08 == 0 {
			return AuthentificationError("Authentification error")
		}
	}
	return
}
*/
func (r *MFRC522) PICC_AuthentificateKeyA(uid UID, key []byte, sector int) (err error) {
	return r.authentificateKey(PICC_CMD_MF_AUTH_KEY_A, uid, key, sector)
}

func (r *MFRC522) PICC_AuthentificateKeyB(uid UID, key []byte, sector int) (err error) {
	return r.authentificateKey(PICC_CMD_MF_AUTH_KEY_B, uid, key, sector)
}

func (r *MFRC522) PICC_StopCrypto1() error {
	// Clear MFCrypto1On bit
	r.PCD_ClearRegisterBitMask(Status2Reg, 0x08) // Status2Reg[7..0] bits are: TempSensClear I2CForceHS reserved reserved MFCrypto1On ModemState[2:0]
	return r.PICC_Halt()
}

func calcBlockAddress(sector int, block int) byte {
	return byte(sector*4 + block)
}

func (r *MFRC522) PICC_ReadBlock(sector int, block int) (result []byte, err error) {
	// Authentification required
	var buffer [4]byte
	copy(buffer[:], []byte{PICC_CMD_MF_READ, calcBlockAddress(sector, block)})

	if crc, err := r.PCD_CalculateCRC(ISO_14443_CRC_RESET, buffer[:2], INTERUPT_TIMEOUT); err != nil {
		return nil, err
	} else {
		buffer[2] = crc[0]
		buffer[3] = crc[1]
		validBits := byte(0)
		if result, err = r.PCD_CommunicateWithPICC(PCD_Transceive, buffer[:], &validBits, INTERUPT_TIMEOUT); err != nil {
			return nil, err
		}
		// TODO check crc
		if len(result) > 16 {
			return result[:16], nil
		} else {
			return nil, AuthentificationError(fmt.Sprintf("PICC_ReadBlock [% x]", result))
		}
	}
}

func (r *MFRC522) PICC_WriteBlock(sector int, block int, data []byte) error {
	// Authentification required

	var buffer [4]byte
	copy(buffer[:], []byte{PICC_CMD_MF_WRITE, calcBlockAddress(sector, block)})

	if crc, err := r.PCD_CalculateCRC(ISO_14443_CRC_RESET, buffer[:2], INTERUPT_TIMEOUT); err != nil {
		return err
	} else {
		buffer[2] = crc[0]
		buffer[3] = crc[1]
	}

	validBits := byte(0)
	if res, err := r.PCD_CommunicateWithPICC(PCD_Transceive, buffer[:], &validBits, INTERUPT_TIMEOUT); err != nil {
		return err
	} else {
		if res[0]&0x0F != 0x0A {
			log.Printf("PICC_CMD_MF_WRITE result: [% x]\n", res)
			return AuthentificationError("Can't authorize write")
		}
	}

	var newData [18]byte

	copy(newData[:], data[:16])

	crc, err := r.PCD_CalculateCRC(ISO_14443_CRC_RESET, newData[:16], INTERUPT_TIMEOUT)
	if err != nil {
		return err
	}

	newData[16] = crc[0]
	newData[17] = crc[1]
	if _, err := r.PCD_CommunicateWithPICC(PCD_Transceive, newData[:], &validBits, INTERUPT_TIMEOUT); err != nil {
		return err
	}
	log.Printf("    PICC_WriteBlock end\n")
	return nil
}
