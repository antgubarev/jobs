package job

type Locked struct{}

func (lckd *Locked) Error() string {
	return "job is locked"
}

type InvalidUnlockArguments struct {
	msg string
}

func NewInvalidUnlockArgumentsErr(msg string) error {
	return &InvalidUnlockArguments{msg: msg}
}

func (ivua *InvalidUnlockArguments) Error() string {
	return ivua.msg
}
