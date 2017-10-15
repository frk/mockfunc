package mockfunc

import (
	"fmt"
	"testing"
)

func TestMock(t *testing.T) {
	fmt.Println(Foobar(""))

	fm := Mock(&Foobar)
	defer fm.Done()

	tests := []struct {
		args *Args
		in   string
	}{{
		args: In("hello").Out(123332, nil),
		in:   "hello",
	}, {
		args: In("calipso").Out(-888, fmt.Errorf("sdfsd")),
		in:   "calipso 2",
	}, {
		args: Out(-888, fmt.Errorf("sdfsd")),
		in:   "calipso 3",
	}}

	for _, tt := range tests {
		if err := fm.Want(tt.args); err != nil {
			t.Error(err)
		}

		Foobar(tt.in)

		if err := fm.Check(); err != nil {
			t.Error(err)
		}
	}
}
