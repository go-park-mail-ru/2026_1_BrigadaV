package testutil

import "time"

func PtrString(s string) *string {
	return &s
}

func PtrBool(b bool) *bool {
	return &b
}

func PtrTime(t time.Time) *time.Time {
	return &t
}
