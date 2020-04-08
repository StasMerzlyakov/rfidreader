package main

import (
	"log"
	"os"
	"rfidreader/mfrc522"
	"time"

	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/spi/spireg"
	"periph.io/x/periph/host"
)

const (
	IRQ = "GPIO4"
	RST = "GPIO25"
)

func run() int {

	if _, err := host.Init(); err != nil {
		log.Printf(err.Error())
		return 1
	}
	// Use spireg SPI port registry to find the first available SPI bus.
	spiPort, err := spireg.Open("")
	if err != nil {
		log.Printf(err.Error())
		return 1
	}
	defer spiPort.Close()

	rstPin := gpioreg.ByName(RST)
	irqPin := gpioreg.ByName(IRQ)

	reader, err := mfrc522.NewMFRC522(spiPort, rstPin, irqPin)

	//if err := reader.PCD_HardReset(); err != nil {
	//	log.Fatal(err.Error())
	//}
	//reader.PCD_Init()
	if err = reader.PCD_PerformSelfTest(); err != nil {
		reader.PCD_Reset()
		if err = reader.PCD_PerformSelfTest(); err != nil {
			log.Printf(err.Error())
			return 1
		}
	}

	log.Printf("MFRC522 Self test complete")
	/*
			reader.PCD_Init()
			for i := 0; i < 20; i++ {

				result, err := reader.PICC_IsNewCardPresent()
				log.Printf("MFRC isNewCardPresent: %t", result)
				if result {

				}
				time.Sleep(1000 * time.Millisecond)
			}

		// Reset FIFO buffer
		if err := reader.PCD_SetRegisterBitMask(mfrc522.FIFOLevelReg, 0x80); err != nil {
			log.Fatalf(err.Error())
		}
		log.Printf("Reset FIFO buffer")

		if hiAlert, err := reader.PCD_ReadRegister(mfrc522.Status1Reg); err != nil {
			log.Fatalf(err.Error())
		} else {
			log.Printf("before %b", hiAlert)

		}

			callbackInvoke := false

			callback := func() {
				callbackInvoke = true
				log.Printf("IRQ Called")
			}

			// Уровень в 8
			if err := reader.PCD_WriteRegister(mfrc522.WaterLevelReg, 0x08); err != nil {
				log.Fatalf(err.Error())
			}
			log.Printf("WaterLevelReg set")

			// 9.3.1.3 ComIEnReg register
			// Устанавливаем бит, отвечающий за блокировку/разблокировку  interrupt requests
			// Bit 7 6 5 4 3 2 1 0
			// Symbol IRqInv TxIEn RxIEn IdleIEn HiAlertIEn LoAlertIEn ErrIEn TimerIEn
			// Access R/W R/W R/W R/W R/W R/W R/W R/W

			// собираемся отслеживать HiAlertIEn
			reader.PCD_SetRegisterBitMask(mfrc522.ComIEnReg, 0x08)

			reader.PCD_IRQ_Callback(callback, 10*time.Second)
			log.Printf("Callback set")

			// Буфер 64 байта
			// Пишем 60 байт
			arr := make([]byte, 60)
			if err := reader.PCD_WriteRegisterArr(mfrc522.FIFODataReg, arr); err != nil {
				log.Fatalf(err.Error())
			}
			log.Printf("60 byte written")

			if statusErr, err := reader.PCD_ReadRegister(mfrc522.ErrorReg); err != nil {
				log.Fatalf(err.Error())
			} else {
				log.Printf("ErrorReg: %b", statusErr)
			}

			if hiAlert, err := reader.PCD_ReadRegister(mfrc522.Status1Reg); err != nil {
				log.Fatalf(err.Error())
			} else {
				log.Printf("%b", hiAlert)
				if hiAlert&0x02>>1 != 1 {
					log.Fatalf("hiAlert bit not set")
				}
			}

			if err := reader.PCD_WriteRegisterArr(mfrc522.FIFODataReg, arr); err != nil {
				log.Fatalf(err.Error())
			}
			log.Printf("additional 60 byte written")

			if statusErr, err := reader.PCD_ReadRegister(mfrc522.ErrorReg); err != nil {
				log.Fatalf(err.Error())
			} else {
				log.Printf("ErrorReg: %b", statusErr)
				if statusErr&0x10>>4 != 1 {
					log.Fatalf("BufferOvfl bit not set")
				}
			}*/

	/*arr := make([]byte, 63)
	result, err := reader.PCD_CalculateCRC(arr, mfrc522.MAX_TIMEOUT_MILLS)
	if err == nil {
		log.Printf("CRC result: [% x]", result)
		log.Printf("Test 1 success")
	} else {
		log.Fatalf(err.Error())
	}

	result, err = reader.PCD_CalculateCRC(arr, mfrc522.MAX_TIMEOUT_MILLS)
	if err == nil {
		log.Printf("CRC result: [% x]", result)
		log.Printf("Test 2 success")
	} else {
		log.Fatalf(err.Error())
	}*/
	arr := make([]byte, 63)
	for i := 0; i < 10; i++ {
		if result, err := reader.PCD_CalculateCRC(arr, mfrc522.INTERUPT_TIMEOUT); err != nil {
			log.Printf(err.Error())
			return 1
		} else {
			log.Printf("CRC result: [% x]", result)
			log.Printf("Test 2 success")
		}
	}

	if err := reader.PCD_Reset(); err != nil {
		log.Printf(err.Error())
		return 1
	}
	reader.PCD_Init()

	for i := 0; i < 50; i++ {
		val := reader.PICC_IsNewCardPresent()
		log.Printf("IsNewCardPresent %t", val)
		if val {
			log.Printf("    Try select card\n")
			if uid, err := reader.PICC_Select(); err != nil {
				log.Printf(err.Error())
			} else {
				log.Printf("Found card:\n")
				log.Printf("    uid: [% x]\n", uid.Uid)
				log.Printf("    sak: %08b\n", uid.Sak)
				log.Printf("    type: %d\n", uid.PicType)
				reader.PCD_AntennaOff()
				reader.PCD_AntennaOn()
			}
		}
		time.Sleep(time.Millisecond * 500)
	}
	return 0

}

func main() {
	os.Exit(run())
}
