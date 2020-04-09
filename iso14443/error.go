package iso14443

import "errors"

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
