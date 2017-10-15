package mockfunc

import (
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"sync"
	"unsafe"

	"github.com/frk/mockfunc/xyz"
)

var Foobar = xyz.Foobar

// registry keeps track of already mocked functions to avoid
// mocking a function more than once.
var registry = struct {
	sync.Mutex
	m map[unsafe.Pointer]*mock
}{m: make(map[unsafe.Pointer]*mock)}

type mock struct {
	name string
	orig interface{}
	rtyp reflect.Type
	addr reflect.Value
	swap reflect.Value

	sync.Mutex
	funcs map[string]*Func
}

func getmock(rv reflect.Value) *mock {
	key := unsafe.Pointer(rv.Pointer())

	registry.Lock()
	m := registry.m[key]
	if m == nil {
		rv = rv.Elem()

		m = &mock{funcs: make(map[string]*Func)}
		m.name = runtime.FuncForPC(rv.Pointer()).Name()
		m.rtyp = rv.Type()
		m.orig = rv.Interface()
		m.addr = rv.Addr()
		m.swap = reflect.MakeFunc(m.rtyp, m.call)

		rv.Set(m.swap)
		registry.m[key] = m
	}
	registry.Unlock()
	return m
}

func delmock(rv reflect.Value) {
	key := unsafe.Pointer(rv.Pointer())
	registry.Lock()
	delete(registry.m, key)
	registry.Unlock()
}

func (m *mock) call(args []reflect.Value) (result []reflect.Value) {
	orig := reflect.ValueOf(m.orig)

	f, err := m.getFunc()
	if err != nil {
		return orig.Call(args)
	}

	// safe the arguments passed to the call
	f.actual = append(f.actual, args)

	if out := f.nextout(); len(out) > 0 {
		return out
	}
	return orig.Call(args)
}

func (m *mock) getFunc() (*Func, error) {
	test, err := testname()
	if err != nil {
		return nil, err
	}

	m.Lock()
	f := m.funcs[test]
	m.Unlock()

	if f == nil {
		return nil, fmt.Errorf("no Func found for test: ", test)
	}
	return f, nil
}

func (m *mock) setFunc(f *Func) error {
	test, err := testname()
	if err != nil {
		return err
	}

	m.Lock()
	if _, ok := m.funcs[test]; ok {
		return fmt.Errorf("already registered %s", test)
	}
	m.funcs[test] = f
	m.Unlock()
	return nil
}

func (m *mock) delFunc() error {
	test, err := testname()
	if err != nil {
		return err
	}

	m.Lock()
	delete(m.funcs, test)
	m.Unlock()
	return nil
}

func (m *mock) numFunc() int {
	return len(m.funcs)
}

type Func struct {
	*mock
	num      int
	fakeout  [][]reflect.Value
	expected [][]reflect.Value
	actual   [][]reflect.Value
}

func Mock(fvar interface{}) *Func {
	rv := reflect.ValueOf(fvar)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Func {
		panic("mockfunc: Info argument must be a pointer to a func variable.")
	}
	m := getmock(rv)

	F := &Func{mock: m}
	if err := m.setFunc(F); err != nil {
		panic(err)
	}
	return F
}

func (f *Func) Want(fc *CallArgs) error {
	if err := f.validargs(fc); err != nil {
		return err
	}
	f.expected = append(f.expected, fc.in)
	f.fakeout = append(f.fakeout, fc.out)
	return nil
}

func (f *Func) MustWant(fc *CallArgs) {
	if err := f.Want(fc); err != nil {
		panic(err)
	}
}

// Check compares the expected input against the actual input, returns an error
// if they're unequal, and then resets the receiver to its initial state.
//
// Check is intended to be called at the end of each test case.
func (f *Func) Check() error {
	defer f.reset()

	if elen, alen := len(f.expected), len(f.actual); elen != alen {
		return &BadFuncCallNumError{fname: f.name, want: elen, got: alen}
	}

	errlist := &ErrorList{}
	for i, expected := range f.expected {
		if expected == nil {
			// skip comparison if no expected input was specified
			continue
		}

		actual := f.actual[i]
		if elen, alen := len(expected), len(actual); elen != alen {
			errlist.Add(&BadFuncArgNumError{fname: f.name, i: i, want: elen, got: alen})
		}

		for j, exp := range expected {
			act := actual[j]

			want := exp.Interface()
			got := act.Interface()
			if !reflect.DeepEqual(want, got) {
				errlist.Add(&BadFuncArgValueError{
					fname: f.name, i: i, j: j, want: want, got: got})
			}

		}
	}

	if errlist.Len() > 0 {
		return errlist
	}
	return nil
}

// Done removes the receiver Func from the registry and if the mock has no
// other Func instances associated with it, it resets the mocked function
// back to the original.
//
// Done should be called at the end of the TestXxx function in which the
// receiver was created, utilizing defer is recommended.
func (f *Func) Done() error {
	if err := f.mock.delFunc(); err != nil {
		return err
	}
	if f.mock.numFunc() == 0 {
		f.addr.Elem().Set(reflect.ValueOf(f.orig))
		delmock(f.addr)
	}
	return nil
}

// The validargs method checks whether the CallArgs' types match the types
// of the mocked function's actual arguments and return values.
func (f *Func) validargs(fc *CallArgs) error {
	if fc.in != nil { // check only if expected input was specified
		if len1, len2 := f.rtyp.NumIn(), len(fc.in); len1 != len2 {
			return &BadCallArgsNumError{
				fname: f.name, want: len1, got: len2}
		}
	}
	if fc.out != nil { // check only if fake output was specified
		if len1, len2 := f.rtyp.NumOut(), len(fc.out); len1 != len2 {
			return &BadCallArgsNumError{
				fname: f.name, want: len1, got: len2, out: true}
		}
	}

	if err := f.validargtypes(fc.in, f.rtyp.In, false); err != nil {
		return err
	}
	if err := f.validargtypes(fc.out, f.rtyp.Out, true); err != nil {
		return err
	}
	return nil
}

// validargtypes is a helper that compares CallArgs contents against the actual func's input/output types.
func (f *Func) validargtypes(args []reflect.Value, argtype func(int) reflect.Type, out bool) error {
	for i, v := range args {
		typ := argtype(i)
		if !v.IsValid() {
			v = reflect.New(typ).Elem()
			args[i] = v
		}
		if w, g := typ, v.Type(); w != g {
			if !g.ConvertibleTo(w) {
				return &BadCallArgsTypeError{
					fname: f.name,
					i:     i,
					want:  w,
					got:   g,
					out:   out,
				}
			}
			v = v.Convert(w)
			args[i] = v
		}
	}
	return nil
}

// nextout returns the next-in-order fake output.
func (f *Func) nextout() (out []reflect.Value) {
	if len(f.fakeout) > f.num {
		out = f.fakeout[f.num]
	}
	f.num += 1
	return out
}

// reset resets the receiver's state to zero.
func (f *Func) reset() {
	f.num = 0
	f.fakeout = nil
	f.expected = nil
	f.actual = nil
}

// The CallArgs type is used to specify the expected input to a function call
// as well as a fake output from the same function call.
//
// CallArgs without fake output, that is, those that are created with func In
// without calling the Out method afterwards will cause the mock to call the
// original function to retrieve and return the output.
//
// CallArgs wihtout expected input, that is, those that are created with
// func Out will cause the mock to skip comparing the actual input.
type CallArgs struct {
	in  []reflect.Value
	out []reflect.Value
}

// In allocates and returns a new CallArgs instance with the given values
// set as the expected input to a function call.
func In(vs ...interface{}) *CallArgs {
	fc := &CallArgs{}
	for _, v := range vs {
		fc.in = append(fc.in, reflect.ValueOf(v))
	}
	return fc
}

// Out allocates and returns a new CallArgs instance with the given values
// set as the fake output from a function call.
func Out(vs ...interface{}) *CallArgs {
	return (&CallArgs{}).Out(vs...)
}

// Out appends the given values to the receiver's fake output.
func (fc *CallArgs) Out(vs ...interface{}) *CallArgs {
	for _, v := range vs {
		fc.out = append(fc.out, reflect.ValueOf(v))
	}
	return fc
}

// testname returns the package path-qualified function name of
// the test in which the caller of testname was executed.
func testname() (name string, err error) {
	callers := reverse(collectcallers())
	frames := runtime.CallersFrames(callers)
	for {
		f, more := frames.Next()
		if testframe(f) {
			name = f.Func.Name()
		}
		if len(name) > 0 || !more {
			break
		}
	}
	if len(name) == 0 {
		return "", fmt.Errorf("unable to resolve test name...")
	}
	return name, nil
}

// collectcallers returns the PCs of all callers of the function.
func collectcallers() (callers []uintptr) {
	for {
		pcs := make([]uintptr, 10)
		num := runtime.Callers(len(callers), pcs)
		callers = append(callers, pcs[:num]...)
		if num < len(pcs) {
			break
		}
	}
	return callers
}

var reTestName = regexp.MustCompile(`\.Test[0-9A-Za-z_]*$`)
var reTestFile = regexp.MustCompile(`_test\.go$`)

// testframe reports whether the given runtime.Frame matches
// a TestXxx function and is located in a test file.
func testframe(f runtime.Frame) bool {
	return reTestFile.MatchString(f.File) &&
		reTestName.MatchString(f.Func.Name())
}

// The reverse func reverses the order of the callers PCs
// making the farthest caller be at the 0th index.
func reverse(callers []uintptr) []uintptr {
	// https://github.com/golang/go/wiki/SliceTricks#reversing
	for i := len(callers)/2 - 1; i >= 0; i-- {
		opp := len(callers) - 1 - i
		callers[i], callers[opp] = callers[opp], callers[i]
	}
	return callers
}