package main

import (
	"log"
	"os"
	"rfidreader/iso14443"
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

	mfrc522dev, err := mfrc522.NewMFRC522(spiPort, rstPin, irqPin)

	//if err := reader.PCD_HardReset(); err != nil {
	//	log.Fatal(err.Error())
	//}
	//reader.PCD_Init()

	if err = mfrc522dev.PCD_PerformSelfTest(); err != nil {
		mfrc522dev.PCD_Reset()
		if err = mfrc522dev.PCD_PerformSelfTest(); err != nil {
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
		// crcValue 0xffff - default value
		if result, err := mfrc522dev.PCD_CalculateCRC(0xffff, arr, iso14443.INTERUPT_TIMEOUT); err != nil {
			log.Printf(err.Error())
			return 1
		} else {
			log.Printf("CRC result: [% x]", result)
			log.Printf("Test 2 success")
		}
	}

	if err := mfrc522dev.PCD_Reset(); err != nil {
		log.Printf(err.Error())
		return 1
	}

	driver := iso14443.NewISO14443Driver(mfrc522dev)

	mfrc522dev.PCD_Init()
	defer mfrc522dev.PCD_AntennaOff()

	for i := 0; i < 50; i++ {
		if err := mfrc522dev.PCD_AntennaOn(); err != nil {
			log.Printf("mfrc522dev.PCD_AntennaOn error %s\n", err.Error())
		}
		val := driver.PICC_IsNewCardPresent()
		log.Printf("IsNewCardPresent %t", val)
		if val {
			log.Printf("    Try select card\n")
			if uid, err := driver.PICC_Select(); err != nil {
				log.Printf(err.Error())
			} else {
				log.Printf("Found card:\n")
				log.Printf("    uid: [% x]\n", uid.Uid)
				log.Printf("    sak: %08b\n", uid.Sak)
				log.Printf("    type: %d\n", uid.PicType)
			}
		}
		time.Sleep(time.Millisecond * 500)
		if err := mfrc522dev.PCD_AntennaOff(); err != nil {
			log.Printf("mfrc522dev.PCD_AntennaOff error %s\n", err.Error())
		}

		//driver.ScanPrepare()
	}

	return 0

}

func main() {
	os.Exit(run())
}
