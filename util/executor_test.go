package util

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

type Accumulator struct {
	Sum int
	Cnt int
}

func (a *Accumulator) Accum(n int) {
	a.Sum += n
	a.Cnt += 1
}

func (a *Accumulator) Init() error {
	return nil
}

func (a *Accumulator) Finish() error {
	return nil
}

func (a *Accumulator) Handle(ctx context.Context, args interface{}) (interface{}, error) {
	n, ok := args.(int)
	if !ok {
		return nil, errors.New("Bad argument")
	}

	a.Accum(n)
	return *a, nil
}

func TestAsyncExecutor(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	acc := new(Accumulator)
	ex := NewAsyncExecutor(acc)

	argCh := make(chan interface{})
	resCh, errCh, _ := ex.Run(ctx, argCh)

	go func() {
		for i := 0; i < 10; i++ {
			argCh <- i
		}

		cancel()
	}()

	var sum int = 0
	var cnt int = 0

	for {
		select {
		case ret := <-resCh:
			a, ok := ret.(Accumulator)
			if !ok {
				t.Errorf("Bad type: %t", ret)
				return
			}

			sum += cnt
			cnt += 1
			// Output: (0 + 9)*10/2 = 45
			if a.Sum != sum {
				t.Errorf("Sum not equal %d, got %d\n", sum, a.Sum)
				return
			}
			if a.Cnt != cnt {
				t.Errorf("Cnt not equal %d, got %d\n", cnt, a.Cnt)
				return
			}

		case err := <-errCh:
			if err == context.Canceled {
				//fmt.Println("Cancel signal recv")
				return
			}

			t.Errorf("Channel err: %s\n", err.Error())
			return

		case <-ctx.Done():
			return
		}
	}
}

type IterHandler struct {
}

type IterType int

func (i *IterType) Next(ctx context.Context) (interface{}, error) {
	if *i > 0 {
		ret := int(*i)
		*i = *i - 1
		return ret, nil
	}

	return nil, ErrIteratorEOF
}

func (m *IterHandler) Init() error {
	return nil
}

func (m *IterHandler) Finish() error {
	return nil
}

func (m *IterHandler) Handle(ctx context.Context, args interface{}) (interface{}, error) {
	in, ok := args.(int)
	if !ok {
		return nil, errors.New("Unsupported format for args")
	}

	var i IterType = IterType(in)
	return &i, nil
}

func TestAsyncExecutorWithIterator(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	handler := new(IterHandler)
	ex := NewAsyncExecutor(handler)

	in := make(chan interface{})
	out, err, _ := ex.Run(ctx, in)

	go func() {
		in <- 18
		time.Sleep(time.Second * 2)
		cancel()
	}()

	for {
		select {
		case result := <-out:
			n, _ := result.(int)
			fmt.Printf("%d\n", n)
		case e := <-err:
			t.Error(e)
			return
		case <-ctx.Done():
			return
		}
	}
}

type MultiType struct{}

type Args struct {
	N int
	F int
}

func (m *MultiType) Init() error {
	return nil
}

func (m *MultiType) Finish() error {
	return nil
}

func (m *MultiType) Handle(ctx context.Context, args interface{}) (interface{}, error) {
	in, ok := args.(Args)
	if !ok {
		return 0, errors.New("Unsupported format for args")
	}

	if in.F == 0 {
		return 0, errors.New("Ilegal argument for division")
	}
	return in.N / in.F, nil
}

func printResultByExecutorId(id int, result interface{}) {
	n, _ := result.(int)
	fmt.Printf("executor[%d]: %d\n", id, n)
}

func printErrByExecutorId(t *testing.T, id int, err error) {
	t.Errorf("executor[%d]: %s\n", id, err.Error())
}

func TestFanOutAsyncExecutor(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	ex := NewFanOutAsyncExecutor()

	var fanOutFactor = 5
	in := make(chan interface{}, fanOutFactor)
	exSlice := make([]Executor, fanOutFactor)

	// Create a bunch of executors
	for i := 0; i < fanOutFactor; i++ {
		exSlice[i] = NewAsyncExecutor(new(MultiType))
	}

	out, err, _ := ex.Run(ctx, in, exSlice)

	go func() {
		in <- Args{100, 10}
		in <- Args{200, 5}
		in <- Args{300, 3}
		in <- Args{99, 11}

		time.Sleep(time.Second * 1)
		cancel()
	}()

	var result interface{}
	var e error
	for {
		select {
		case result = <-out[0]:
			printResultByExecutorId(0, result)
		case result = <-out[1]:
			printResultByExecutorId(1, result)
		case result = <-out[2]:
			printResultByExecutorId(2, result)
		case result = <-out[3]:
			printResultByExecutorId(3, result)
		case result = <-out[4]:
			printResultByExecutorId(4, result)
		case e = <-err[0]:
			printErrByExecutorId(t, 0, e)
			return
		case e = <-err[1]:
			printErrByExecutorId(t, 0, e)
			return
		case e = <-err[2]:
			printErrByExecutorId(t, 0, e)
			return
		case e = <-err[3]:
			printErrByExecutorId(t, 0, e)
			return
		case e = <-err[4]:
			printErrByExecutorId(t, 0, e)
			return
		case <-ctx.Done():
			return
		}
	}
}

var fanInCnt int = 5

func TestFanInAsyncExecutor(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	ex := NewFanInAsyncExecutor(nil)

	in := make([]chan interface{}, fanInCnt)
	errs := make([]chan error, fanInCnt)
	for i := 0; i < fanInCnt; i++ {
		in[i] = make(chan interface{})
		errs[i] = make(chan error)
	}

	inRead := make([]<-chan interface{}, fanInCnt)
	errsRead := make([]<-chan error, fanInCnt)
	for i := 0; i < fanInCnt; i++ {
		inRead[i] = in[i]
		errsRead[i] = errs[i]
	}

	out, err, _ := ex.Run(ctx, inRead, errsRead)

	go func() {
		for i := 0; i < 100; i++ {
			in[i%len(in)] <- i
		}

		time.Sleep(time.Second * 3)
		cancel()
	}()

	for {
		select {
		case result := <-out:
			n, _ := result.(int)
			fmt.Printf("%d\n", n)
		case e := <-err:
			t.Error(e)
			return
		case <-ctx.Done():
			return
		}
	}
}
