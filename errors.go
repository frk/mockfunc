package mockfunc

import (
	"fmt"
	"reflect"
)

const (
	// ANSI color values used to colorize terminal output.
	redColor    = "\033[91m"
	yellowColor = "\033[93m"
	purpleColor = "\033[95m"
	cyanColor   = "\033[96m"
	greenColor  = "\033[92m"
	stopColor   = "\033[0m"
)

type BadArgNumError struct {
	fname     string
	want, got int
	out       bool
}

func (e *BadArgNumError) Error() string {
	const mesg = "mockfunc: %q %s specified with wrong number of arguments;" +
		" want %s, got %s."

	mf := "Call"
	if e.out {
		mf = "Out"
	}

	want := fmt.Sprintf("%s%d%s", yellowColor, e.want, stopColor)
	got := fmt.Sprintf("%s%d%s", yellowColor, e.got, stopColor)

	return fmt.Sprintf(mesg, e.fname, mf, want, got)
}

type BadArgTypeError struct {
	fname     string
	i         int
	want, got reflect.Type
	out       bool
}

func (e *BadArgTypeError) Error() string {
	const mesg = "mockfunc: %q %s %s argument has wrong type;" +
		" want %s, got %s."

	mf := "Call"
	if e.out {
		mf = "Out"
	}

	ith := fmt.Sprintf("%s#%d%s", yellowColor, e.i, stopColor)
	want := fmt.Sprintf("%s%v%s", cyanColor, e.want, stopColor)
	got := fmt.Sprintf("%s%v%s", cyanColor, e.got, stopColor)

	return fmt.Sprintf(mesg, e.fname, mf, ith, want, got)
}
