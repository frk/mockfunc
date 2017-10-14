package mockfunc

import (
	"fmt"
	"testing"
	//"github.com/frk/mockfunc/xyz"
)

func TestF1(t *testing.T) {
	fmt.Println(Foobar(""))

	fm := Mock(&Foobar)
	fm.Want(Call("hello").Out(777, fmt.Errorf("sdfsd")))

	fmt.Println(Foobar(""))

	fm.Done()

	fmt.Println(Foobar(""))
}
