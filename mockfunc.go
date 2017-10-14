package mockfunc

import (
	"reflect"
	"runtime"

	"github.com/frk/mockfunc/xyz"
)

//func X() {
//	pc := make([]uintptr, 0, 20)
//	fmt.Println(runtime.Callers(0, pc))
//	fmt.Println(pc)
//}

var Foobar = xyz.Foobar

type Func struct {
	name string
	orig interface{}
	rtyp reflect.Type
	mock reflect.Value
	addr reflect.Value

	want []*FuncCall
	got  []*FuncCall
}

func Mock(fnvar interface{}) *Func {
	rv := reflect.ValueOf(fnvar)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Func {
		panic("mockfunc: Info argument must be a pointer to a func variable.")
	}
	rv = rv.Elem()

	F := &Func{}
	F.rtyp = rv.Type()
	F.orig = rv.Interface()
	F.addr = rv.Addr()
	F.name = runtime.FuncForPC(rv.Pointer()).Name()
	F.mock = reflect.MakeFunc(F.rtyp, func(args []reflect.Value) (result []reflect.Value) {
		return F.call(args)
	})
	rv.Set(F.mock)

	return F
}

func (f *Func) call(args []reflect.Value) (result []reflect.Value) {
	return f.want[0].out
}

func (f *Func) Want(fc *FuncCall) {
	if err := f.validCall(fc); err != nil {
		panic(err)
	}
	f.want = append(f.want, fc)
}

func (f *Func) validCall(fc *FuncCall) error {
	if w, g := f.rtyp.NumIn(), len(fc.in); w != g {
		return &BadArgNumError{fname: f.name, want: w, got: g}
	}
	if w, g := f.rtyp.NumOut(), len(fc.out); w != g {
		return &BadArgNumError{fname: f.name, want: w, got: g, out: true}
	}
	for i, v := range fc.in {
		intyp := f.rtyp.In(i)
		if !v.IsValid() {
			v = reflect.New(intyp).Elem()
			fc.in[i] = v
		}
		if w, g := intyp, v.Type(); w != g {
			if !g.ConvertibleTo(w) {
				return &BadArgTypeError{fname: f.name, i: i, want: w, got: g}
			}
			v = v.Convert(w)
			fc.in[i] = v
		}
	}

	for i, v := range fc.out {
		outyp := f.rtyp.Out(i)
		if !v.IsValid() {
			v = reflect.New(outyp).Elem()
			fc.out[i] = v
		}
		if w, g := outyp, v.Type(); w != g {
			if !g.ConvertibleTo(w) {
				return &BadArgTypeError{fname: f.name, i: i, want: w, got: g, out: true}
			}
			v = v.Convert(w)
			fc.out[i] = v
		}
	}
	return nil
}

func (f *Func) Err() error {
	return nil
}

func (f *Func) Done() {
	f.addr.Elem().Set(reflect.ValueOf(f.orig))
}

type FuncCall struct {
	in  []reflect.Value
	out []reflect.Value
}

func Call(vs ...interface{}) *FuncCall {
	fc := &FuncCall{}
	for _, v := range vs {
		fc.in = append(fc.in, reflect.ValueOf(v))
	}
	return fc
}

func (fc *FuncCall) Out(vs ...interface{}) *FuncCall {
	for _, v := range vs {
		fc.out = append(fc.out, reflect.ValueOf(v))
	}
	return fc
}