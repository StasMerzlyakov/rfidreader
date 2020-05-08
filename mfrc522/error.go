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

type iso14443Error struct{ error }

func SelectionError(desc string) error {
	return iso14443Error{errors.New(desc)}
}

func CollErrError(desc string) error {
	return iso14443Error{errors.New(desc)}
}

func UnexpectedResponse(desc string) error {
	return iso14443Error{errors.New(desc)}
}

func CommonError(desc string) error {
	return iso14443Error{errors.New(desc)}
}

func CRCCheckError(desc string) error {
	return iso14443Error{errors.New(desc)}
}

func UsageError(desc string) error {
	return mfrc522Error{errors.New(desc)}
}

func AuthentificationError(desc string) error {
	return mfrc522Error{errors.New(desc)}
}
