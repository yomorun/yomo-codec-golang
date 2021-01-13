package y3

import (
	"errors"
	"fmt"
	"testing"
)

func TestObservable(t *testing.T) {
	buf := []byte{0x81, 0x16, 0xb0, 0x14, 0x10, 0x4, 0x79, 0x6f, 0x6d, 0x6f, 0x11, 0x2, 0x43, 0xe4, 0x92, 0x8, 0x13, 0x2, 0x41, 0xf0, 0x14, 0x2, 0x42, 0x20}
	sourceChannel := make(chan interface{})

	go func() {
		sourceChannel <- buf
	}()

	callback := func(v []byte) (interface{}, error) {
		if (v[0] == 17) && (v[1] == 2) && (v[2] == 67) && (v[3] == 228) {
			return "ok", nil
		} else {
			return nil, errors.New("fail")
		}

	}

	source := &ObservableImpl{iterable: &IterableImpl{channel: sourceChannel}}

	consumer := source.Subscribe(0x11).OnObserve(callback)

	for c := range consumer {
		fmt.Println(c)
	}
}
