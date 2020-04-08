#!/bin/bash

source ./environment.sh

rm -rf bin
mkdir bin
#GOARCH=arm go build -o bin/gpio gpio/main.go
#GOARCH=arm go build -o bin/spi spi/main.go
#GOARCH=arm go build -o bin/selftest2 selftest2/main.go
GOARCH=arm go build -o bin/selftest3 mfrc522_test/main.go

