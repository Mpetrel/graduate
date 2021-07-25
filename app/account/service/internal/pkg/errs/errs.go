package errs

import "fmt"

type Error struct {
	Msg string
}

func (e Error) Error() string {
	return fmt.Sprintf("error: %s", e.Msg)
}

var (
	ErrNotFound = Error{Msg: "Record Not Found"}
	EmailAlreadyUsed = Error{Msg: "Email already used"}
)


func IsNotFound(err error) bool {
	return err == ErrNotFound
}

func IsEmailAlreadyUsed(err error) bool {
	return err == EmailAlreadyUsed
}