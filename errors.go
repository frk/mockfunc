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

type BadArgsNumError struct {
	fname     string
	want, got int
	out       bool
}

func (e *BadArgsNumError) Error() string {
	const text = "frk/mockfunc: %q %s (Args)" +
		" specified with wrong number of arguments; want %s, got %s."

	call := "In"
	if e.out {
		call = "Out"
	}
	want := fmt.Sprintf("%s%d%s", yellowColor, e.want, stopColor)
	got := fmt.Sprintf("%s%d%s", yellowColor, e.got, stopColor)
	return fmt.Sprintf(text, e.fname, call, want, got)
}

type BadArgsTypeError struct {
	fname     string
	i         int
	want, got reflect.Type
	out       bool
}

func (e *BadArgsTypeError) Error() string {
	const text = "frk/mockfunc: %q %s (Args)" +
		" %s argument has wrong type; want %s, got %s."

	call := "In"
	if e.out {
		call = "Out"
	}
	ith := fmt.Sprintf("%s#%d%s", yellowColor, e.i, stopColor)
	want := fmt.Sprintf("%s%v%s", cyanColor, e.want, stopColor)
	got := fmt.Sprintf("%s%v%s", cyanColor, e.got, stopColor)
	return fmt.Sprintf(text, e.fname, call, ith, want, got)
}

type BadFuncCallNumError struct {
	fname     string
	want, got int
}

func (e *BadFuncCallNumError) Error() string {
	const text = "frk/mockfunc: Inconsistent number of calls to %q;" +
		" want %s, got %s."

	want := fmt.Sprintf("%s%d%s", yellowColor, e.want, stopColor)
	got := fmt.Sprintf("%s%d%s", yellowColor, e.got, stopColor)
	return fmt.Sprintf(text, e.fname, want, got)
}

type BadFuncArgNumError struct {
	fname        string
	i, want, got int
}

func (e *BadFuncArgNumError) Error() string {
	const text = "frk/mockfunc: Inconsistent number of arguments passed" +
		" to the %s call of %q; want %s, got %s."

	calln := fmt.Sprintf("%s#%d%s", yellowColor, e.i, stopColor)
	want := fmt.Sprintf("%s%d%s", yellowColor, e.want, stopColor)
	got := fmt.Sprintf("%s%d%s", yellowColor, e.got, stopColor)
	return fmt.Sprintf(text, calln, e.fname, want, got)
}

type BadFuncArgError struct {
	fname     string
	i, j      int
	want, got interface{}
}

func (e *BadFuncArgError) Error() string {
	const text = "frk/mockfunc: Unexpected argument value passed" +
		" to the call of %q (%s call, %s argument); want %s, got %s."

	calln := fmt.Sprintf("%s#%d%s", yellowColor, e.i, stopColor)
	argn := fmt.Sprintf("%s#%d%s", yellowColor, e.j, stopColor)
	want := fmt.Sprintf("%s%v%s", cyanColor, e.want, stopColor)
	got := fmt.Sprintf("%s%v%s", cyanColor, e.got, stopColor)
	return fmt.Sprintf(text, e.fname, calln, argn, want, got)
}

type ErrorList struct {
	list []error
}

func (el *ErrorList) Len() int {
	return len(el.list)
}

func (el *ErrorList) Error() string {
	var text string
	for i, e := range el.list {
		text += fmt.Sprintf("#%d: %s\n", i, e)
	}
	return text
}

func (el *ErrorList) Add(errs ...error) {
	el.list = append(el.list, errs...)
}
