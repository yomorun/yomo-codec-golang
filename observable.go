package y3

import (
	"github.com/yomorun/y3-codec-golang/pkg/spec/encoding"
)

type Iterable interface {
	Observe() <-chan interface{}
}

type Observable interface {
	Iterable
	Subscribe(key byte) Observable
	OnObserve(function func(v []byte) (interface{}, error)) chan interface{}
}

type ObservableImpl struct {
	iterable Iterable
}

type IterableImpl struct {
	channel chan interface{}
}

func (i *IterableImpl) Observe() <-chan interface{} {
	return i.channel
}

func (o *ObservableImpl) Observe() <-chan interface{} {
	return o.iterable.Observe()
}

func (o *ObservableImpl) OnObserve(function func(v []byte) (interface{}, error)) chan interface{} {
	_next := make(chan interface{})

	f := func(next chan interface{}) {
		defer close(next)

		observe := o.Observe()

		for {
			select {
			case item, ok := <-observe:
				if !ok {
					return
				}
				buf := item.([]byte)
				value, err := function(buf)
				if err != nil {
					return
				}

				next <- value
			}
		}
	}

	go f(_next)

	return _next
}

func (o *ObservableImpl) Subscribe(key byte) Observable {

	f := func(next chan interface{}) {
		defer close(next)

		resultBuffer := make([]byte, 0)
		var (
			flow   int32 = 0 //0 未监听到，1 判断长度，2 判断v
			length int32 = 1
			value  int32 = 0
		)

		observe := o.Observe()

		for {
			select {
			case item, ok := <-observe:
				if !ok {
					return
				}

				buf := item.([]byte)
				for _, b := range buf { //T
					if flow == 0 && b == key {
						flow = 1
						resultBuffer = append(resultBuffer, b)
						continue
					}

					if flow == 1 && b != key { // L
						resultBuffer = append(resultBuffer, b)
						l, e := decodeLength(resultBuffer[1 : length+1]) //l 是value占字节，s是l占字节

						if e != nil {
							length++
						} else {
							value = l
							flow = 2
						}
						continue
					}

					if flow == 2 && b != key {
						l := len(resultBuffer)
						if int32(l) == (length + value) {
							resultBuffer = append(resultBuffer, b)
							next <- resultBuffer
							flow = 0
							length = 1
							value = 0
							resultBuffer = make([]byte, 0)
						} else if int32(l) < (length + value) {
							resultBuffer = append(resultBuffer, b)
						}
						continue
					}
				}
			}
		}
	}

	return createObservable(f)

}

func decodeLength(buf []byte) (length int32, err error) {
	varCodec := encoding.VarCodec{}
	err = varCodec.DecodePVarInt32(buf, &length)
	return
}

func createObservable(f func(next chan interface{})) Observable {
	next := make(chan interface{})
	go f(next)
	return &ObservableImpl{iterable: &IterableImpl{channel: next}}
}
