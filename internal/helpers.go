package internal

import "time"

func NewPointerOfInt(val int) *int {
	return &val
}

func NewPointerOfString(val string) *string {
	return &val
}

func NewPointerOfTime(t time.Time) *time.Time {
	return &t
}
