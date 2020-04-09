package mfrc522

import (
	"errors"
)

type mfrc522Error struct{ error }

func UnexpectedIRqError(desc string) error {
	return mfrc522Error{errors.New(desc)}
}

func TimeoutIRqError(desc string) error {
	return mfrc522Error{errors.New(desc)}
}

func CRCIRqError(desc string) error {
	return mfrc522Error{errors.New(desc)}
}

func ErrIRqError(desc string) error {
	return mfrc522Error{errors.New(desc)}
}
