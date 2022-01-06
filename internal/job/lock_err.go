package job

type LockedError struct{}

func (lckd *LockedError) Error() string {
	return "job is locked"
}

type InvalidUnlockArgumentsError struct {
	msg string
}

func NewInvalidUnlockArgumentsErr(msg string) error {
	return &InvalidUnlockArgumentsError{msg: msg}
}

func (ivua *InvalidUnlockArgumentsError) Error() string {
	return ivua.msg
}
