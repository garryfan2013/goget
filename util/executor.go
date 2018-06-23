package util

import (
	"context"
	"errors"
	"reflect"

	_ "fmt"
)

const (
	DefaultChannelSize = 10
)

var (
	ErrIteratorEOF = errors.New("Iterator reaches EOF")
)

func init() {

}

type Executor interface {
	Run(ctx context.Context, in <-chan interface{}) (<-chan interface{}, <-chan error, error)
}

type FanOutExecutor interface {
	Run(ctx context.Context, in <-chan interface{}, executors []Executor) ([]<-chan interface{}, []<-chan error, error)
}

type FanInExecutor interface {
	Run(ctx context.Context, in []<-chan interface{}, errs []<-chan error) (<-chan interface{}, <-chan error, error)
}

/*
	This could be the result of a handler if th handler intends to generate
	a collection of data.
	When the iteration done, return EOF error
*/
type Iterator interface {
	Next(ctx context.Context) (interface{}, error)
}

type Closer interface {
	Close() error
}

/*
	The interface for handling request
*/
type Handler interface {
	Init() error

	Handle(ctx context.Context, d interface{}) (interface{}, error)

	Finish() error
}

/*
	If the iterator needs some cleanning work when done
	this could be the choice
*/
type IterateCloser interface {
	Iterator
	Closer
}

// FanInExecutor is a converge executor ,
// In charge of merging multiple channels' output into one single channel
type FanInAsyncExecutor struct {
	h Handler
}

func NewFanInAsyncExecutor(h Handler) *FanInAsyncExecutor {
	return &FanInAsyncExecutor{h}
}

func (f *FanInAsyncExecutor) Run(ctx context.Context,
	in []<-chan interface{}, errs []<-chan error) (<-chan interface{}, <-chan error, error) {

	cancelCtx, _ := context.WithCancel(ctx)
	outErrCh := make(chan error)
	outDataCh := make(chan interface{}, DefaultChannelSize)

	cases := make([]reflect.SelectCase, len(in)*2+1)
	for i, ch := range in {
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
	}

	for i, err := range errs {
		cases[len(in)+i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(err)}
	}

	// This is the case for cancel signal
	cases[len(in)*2] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(cancelCtx.Done())}

	err := f.h.Init()
	if err != nil {
		return nil, nil, err
	}

	go func() {
		defer f.h.Finish()
		for {
			chosen, value, _ := reflect.Select(cases)

			// case cancel
			if chosen == len(in)*2 {
				return
			}

			// case err
			if chosen > len(in)-1 {
				e, ok := value.Interface().(error)
				if !ok {
					panic("Send something else other than error on error channel")
				}
				select {
				case outErrCh <- e:
				case <-cancelCtx.Done():
					return
				}
			}

			// case data
			if chosen < len(in) {
				select {
				case outDataCh <- value.Interface():
				case <-cancelCtx.Done():
					return
				}
			}
		}
	}()

	return outDataCh, outErrCh, nil
}

// FanOutExecutor expand the data processing of one intput channel to multiple executor
// Notice that the caller should handle the output channel slice
type FanOutAsyncExecutor struct {
}

func NewFanOutAsyncExecutor() *FanOutAsyncExecutor {
	return &FanOutAsyncExecutor{}
}

func (f *FanOutAsyncExecutor) Run(ctx context.Context,
	in <-chan interface{},
	executors []Executor) ([]<-chan interface{}, []<-chan error, error) {

	dataChs := make([]<-chan interface{}, 0, len(executors))
	errChs := make([]<-chan error, 0, len(executors))

	cancelCtx, cancel := context.WithCancel(ctx)
	for _, e := range executors {
		ch1, ch2, err := e.Run(cancelCtx, in)
		if err != nil {
			cancel()
			return nil, nil, err
		}

		dataChs = append(dataChs, ch1)
		errChs = append(errChs, ch2)
	}

	return dataChs, errChs, nil
}

// AsyncExecutor deal with a bunch of input data
// Buf make sure that those input data belongs to the same Request
// Because only one context associated with the Executor
// And there's supposed to be only one context bound to one Request
type AsyncExecutor struct {
	h Handler
}

func NewAsyncExecutor(h Handler) *AsyncExecutor {
	return &AsyncExecutor{h}
}

func (e *AsyncExecutor) Run(ctx context.Context, in <-chan interface{}) (<-chan interface{}, <-chan error, error) {
	cancelCtx, _ := context.WithCancel(ctx)
	outErrCh := make(chan error)
	outDataCh := make(chan interface{}) //, DefaultChannelSize)

	err := e.h.Init()
	if err != nil {
		return nil, nil, err
	}

	go func() {
		defer e.h.Finish()
		for {
			select {
			// In channel deal with input request
			case args := <-in:
				result, err := e.h.Handle(cancelCtx, args)
				if err != nil {
					outErrCh <- err
					break
				}

				func() {
					// If support Closer Interface, we need to close it sooner or later
					if closer, ok := result.(Closer); ok {
						defer closer.Close()
					}

					var iter Iterator
					// If this result is not a iterator type, just simply forwards it
					iter, ok := result.(Iterator)
					if !ok {
						// Normal single result
						select {
						case outDataCh <- result:
						case <-cancelCtx.Done():
						}
						return
					}

					// Here deal with the Iterator type
					for {
						// Make sure that the Next() call could be canceled by cancelCtx
						obj, err := iter.Next(cancelCtx)
						if err != nil {
							if err != ErrIteratorEOF {
								select {
								case outErrCh <- err:
								case <-cancelCtx.Done():
								}
								return
							}
						}

						// Normal round, continue
						if obj != nil {
							select {
							case outDataCh <- obj:
							case <-cancelCtx.Done():
								return
							}
						}

						if err == ErrIteratorEOF {
							return
						}
					}
				}()

			/*
				If we got cancel signals ,most likely from the parent functions,
				Then it's time we quit
			*/
			case <-cancelCtx.Done():
				return
			}
		}
	}()

	return outDataCh, outErrCh, nil
}
