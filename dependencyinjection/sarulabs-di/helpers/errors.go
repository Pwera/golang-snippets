package helpers

type ErrValidation struct {
	msg string
}

func NewErrValidation(msg string) *ErrValidation {
	return &ErrValidation{msg: msg}
}

func (err *ErrValidation) Error() string {
	return err.msg
}

type ErrNotFound struct {
	msg string
}

func NewErrNotFound(msg string) *ErrNotFound {
	return &ErrNotFound{msg: msg}
}

func (err *ErrNotFound) Error() string {
	return err.msg
}
