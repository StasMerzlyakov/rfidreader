package mfrc522

import (
	"bytes"
	"testing"

	"github.com/matryer/is"
)

func TestISO14443aCRC(t *testing.T) {
	is := is.New(t)
	is.True(bytes.Compare(ISO14443aCRC([]byte{0x60, 0x30}), []byte{0x76, 0x4a}) == 0)
}
