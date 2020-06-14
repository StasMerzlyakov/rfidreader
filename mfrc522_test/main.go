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
		if result, err := mfrc522dev.PCD_CalculateCRC(mfrc522.CRC_RESET_VALUE_FFFF, arr, mfrc522.INTERUPT_TIMEOUT); err != nil {
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

	//driver := iso14443.NewISO14443Driver(mfrc522dev)

	defer mfrc522dev.PCD_AntennaOff()

	for i := 0; i < 50; i++ {
		mfrc522dev.PCD_Reset()
		mfrc522dev.PCD_Init()
		if err := mfrc522dev.PCD_AntennaOn(); err != nil {
			log.Printf("mfrc522dev.PCD_AntennaOn error %s\n", err.Error())
		}
		val := mfrc522dev.PICC_IsNewCardPresent()
		log.Printf("IsNewCardPresent %t", val)
		if val {
			log.Printf("    Try select card\n")
			if uid, err := mfrc522dev.PICC_Select(); err != nil {
				log.Printf(err.Error())
			} else {
				log.Printf("Found card:\n")
				log.Printf("    uid: [% x]\n", uid.Uid)
				log.Printf("    sak: %08b\n", uid.Sak)
				log.Printf("    type: %d\n", uid.PicType)
				sector := 0
				if err1 := mfrc522dev.PICC_AuthentificateKeyA(*uid, []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, sector); err1 != nil {
					log.Printf("  Authentificate error\n")
				} else {
					log.Printf("  Authentificate success\n")
					block := 2
					/*if res, err1 := mfrc522dev.PICC_ReadBlock(sector, block); err1 != nil {
						log.Printf("  PCD_StopCrypto1 error\n")
					} else {
						log.Printf(" Block %d %d: [% x]\n", sector, block, res)
					}*/
					/*
						var blockVal [16]byte
						copy(blockVal[:], []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})

						copy(blockVal[6:], []byte{0xFF, 0x07, 0x80})

						copy(blockVal[10:], []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})
						//copy(blockVal[:], []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})

						if err1 := mfrc522dev.PICC_WriteBlock(sector, block, blockVal[:]); err1 != nil {
							log.Printf("  PICC_WriteBlock error\n")
						} else {
							log.Printf(" Write OK\n")
						}*/

					/*var blockVal [16]byte
					copy(blockVal[:], []byte("Hello world"))

					if err1 := mfrc522dev.PICC_WriteBlock(sector, 2, blockVal[:]); err1 != nil {
						log.Printf("  PICC_WriteBlock error\n")
					} else {
						log.Printf(" Write OK\n")
					}

					for block = 3; block >= 0; block-- {
						if res, err1 := mfrc522dev.PICC_ReadBlock(sector, block); err1 != nil {
							log.Printf("  Error %s\n", err1.Error())
						} else {
							if block != 2 {
								log.Printf(" read sector:%d  block: %d value: [% x]\n", sector, block, res)
							} else {
								log.Printf(" read sector:%d  block: %d value: %s\n", sector, block, string(res))
							}
						}
					} */

					// Только чтение для проверки
					if res, err1 := mfrc522dev.PICC_ReadBlock(sector, block); err1 != nil {
						log.Printf("  Error %s\n", err1.Error())
					} else {
						log.Printf(" read sector:%d  block: %d value: %s\n", sector, block, string(res))
					}

					if err1 := mfrc522dev.PICC_StopCrypto1(); err1 != nil {
						log.Printf("  PCD_StopCrypto1 error\n")
					}
				}

			}
		}
		time.Sleep(time.Millisecond * 500)
		if err := mfrc522dev.PCD_AntennaOff(); err != nil {
			log.Printf("mfrc522dev.PCD_AntennaOff error %s\n", err.Error())
		}

		time.Sleep(time.Second * 1)
	}

	return 0

}

func main() {
	os.Exit(run())

	/*var blockVal [16]byte
	copy(blockVal[:], []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})

	copy(blockVal[6:], []byte{0xFF, 0x07, 0x80})

	copy(blockVal[10:], []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})
	*/
	//val := []byte{0x1, 0x2, 0x3, 0x4}
	//	log.Printf("[% x]\n", blockVal)

	/*val := uint16(0x0145)
	fmt.Printf("%016b\n", val)

	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, &val)

	fmt.Printf("% x\n", buf.Bytes())
	*/
	/*uid := uint64(0)
	// /	uid1 := uint32(0)
	for u := 31; u >= 0; u-- {
		x0 := byte(u & 1)
		x1 := byte(u&(1<<1)) >> 1
		x2 := byte(u&(1<<2)) >> 2
		x3 := byte(u&(1<<3)) >> 3
		x4 := byte(u&(1<<4)) >> 4
		y := mfrc522.Fc(x0, x1, x2, x3, x4)

		// Fa ((y0 ∨y1)⊕(y0 ∧y3))⊕ (y2 ∧((y0 ⊕y1)∨y3))
		//y1 := ((x0 | x1) ^ (x0 & x3)) ^ (x2 & ((x0 ^ x1) | x3))

		// Fb ((y0 ∧y1)∨y2)⊕ ((y0 ⊕y1)∧(y2 ∨y3))

		// Fc (y0 ∨((y1 ∨y4)∧(y3 ⊕y4)))⊕((y0 ⊕(y1 ∧y3))∧((y2 ⊕y3)∨(y1 ∧y4)))
		//y1 := (x0 | ((x1 | x4) & (x3 ^ x4))) ^ ((x0 ^ (x1 & x3)) & ((x2 ^ x3) | (x1 & x4)))

		uid |= uint64((y & 1)) << u
		//uid1 |= uint64((y1 & 1)) << u
		log.Printf("%01b%01b%01b%01b%01b = %01b \n", x0, x1, x2, x3, x4, y)
	}

	log.Printf("%x\n", uid) */

	/*uidVal := []byte{0x2a, 0x69, 0x83, 0x43}
	uid := mfrc522.UID{Uid: uidVal}

	nt := []byte{0x3b, 0xae, 0x03, 0x2d}

	key := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}

	init := make([]byte, 4)
	init[0] = uid.Uid[0] ^ nt[0]
	init[1] = uid.Uid[1] ^ nt[1]
	init[2] = uid.Uid[2] ^ nt[2]
	init[3] = uid.Uid[3] ^ nt[3]

	// инициализируем регистр линейного сдвига
	lfsr32 := mfrc522.InitLfsr32FN(key)

	// генерируем ключ ks1
	ks1, _ := lfsr32(init)
	log.Printf("ks1: [% x]\n", ks1)

	// генерируем n_r
	n_r := mfrc522.GenerateNR()
	log.Printf("n_r: [% x]\n", n_r)

	// формируем n_r^ks1
	buffer := make([]byte, 8)
	buffer[0] = ks1[0] ^ n_r[0]
	buffer[1] = ks1[1] ^ n_r[1]
	buffer[2] = ks1[2] ^ n_r[2]
	buffer[3] = ks1[3] ^ n_r[3]

	// формируем вторую часть
	ks2, _ := lfsr32(n_r)
	log.Printf("ks2: [% x]\n", ks2)

	suc := mfrc522.InitSuc(nt)
	// нужен suc2()
	suc()
	suc2, _ := suc()
	log.Printf("ackR: [% x]\n", suc2)

	// формируем ackR^ks2
	buffer[4] = ks2[0] ^ suc2[0]
	buffer[5] = ks2[1] ^ suc2[1]
	buffer[6] = ks2[2] ^ suc2[2]
	buffer[7] = ks2[3] ^ suc2[3]

	log.Printf("buffer: [% x]\n", buffer)
	// Генерируем suc3^ks3
	suc3, _ := suc()

	ks3, _ := lfsr32(nil)
	expeted := make([]byte, 4)
	expeted[0] = suc3[0] ^ ks3[0]
	expeted[1] = suc3[1] ^ ks3[1]
	expeted[2] = suc3[2] ^ ks3[2]
	expeted[3] = suc3[3] ^ ks3[3]
	log.Printf("expected suc3^ks3: [% x]\n", expeted) */
}
