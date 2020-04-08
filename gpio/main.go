package main

import (
	"log"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host"
)

const (
	LED_GPIO_1 = "GPIO14"
	LED_GPIO_2 = "GPIO18"
	BTN_GPIO   = "GPIO17"
)

var LEAD_MAP = map[int]string{
	0: LED_GPIO_1,
	1: LED_GPIO_2}

func fireLED(gname string) {
	// Lookup a pin by its number:
	p := gpioreg.ByName(gname)
	if p == nil {
		log.Fatal("Failed to find " + gname)
	}

	log.Printf("%s: %s\n", p, p.Function())

	if err := p.Out(false); err != nil {
		log.Fatal(err)
	}

	time.Sleep(2000 * time.Millisecond)

	if err := p.Out(true); err != nil {
		log.Fatal(err)
	}

	time.Sleep(2000 * time.Millisecond)

	if err := p.Out(false); err != nil {
		log.Fatal(err)
	}

}

func main() {

	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}
	// Lookup a pin by its number:
	gname := BTN_GPIO
	p := gpioreg.ByName(gname)
	if p == nil {
		log.Fatal("Failed to find " + gname)
	}

	log.Printf("%s: %s\n", p, p.Function())

	// Set it as input, with an internal pull down resistor:
	if err := p.In(gpio.PullDown, gpio.BothEdges); err != nil {
		log.Fatal(err)
	}

	// Wait for edges as detected by the hardware, and print the value read:
	i := 0
	for {
		p.WaitForEdge(-1)
		i = (i + 1) % 2
		go fireLED(LEAD_MAP[i])
		log.Printf("-> %s\n", p.Read())
	}
}
