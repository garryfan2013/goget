package util

import (
	"context"
	"errors"
	"reflect"
)

const (
	DefaultChannelSize = 10
)

var (
	ErrIteratorEOF = errors.New("Iterator reaches EOF")
)

func init() {

}

type Handler interface {
	Handle(ctx context.Context, d interface{}) (interface{}, error)
}

type Executor interface {
	Run(ctx context.Context, in <-chan interface{}) (<-chan interface{}, <-chan error)
}

type FanOutExecutor interface {
	Run(ctx context.Context, in <-chan interface{}, executors []Executor) ([]<-chan interface{}, []<-chan error)
}

type FanInExecutor interface {
	Run(ctx context.Context, in []<-chan interface{}, errs []<-chan error) (<-chan interface{}, <-chan error)
}

// When the iteration done, return EOF error
type Iterator interface {
	Next(ctx context.Context) (interface{}, error)
}

type Closer interface {
	Close() error
}

type IterateCloser interface {
	Iterator
	Closer
}

type AsyncExecutor struct {
	H Handler
}

type FanOutAsyncExecutor struct {
}

type FanInAsyncExecutor struct {
}

// FanInExecutor is a converge executor ,
// In charge of merging multiple channels' output into one single channel
func NewFanInAsyncExecutor() *FanInAsyncExecutor {
	return &FanInAsyncExecutor{}
}

func (f *FanInAsyncExecutor) Run(ctx context.Context,
	in []<-chan interface{},
	errs []<-chan error) (<-chan interface{}, <-chan error) {

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

	go func() {
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
				outErrCh <- e
			}

			// case data
			if chosen < len(in) {
				outDataCh <- value.Interface()
			}
		}
	}()

	return outDataCh, outErrCh
}

// FanOutExecutor expand the data processing of one intput channel to multiple executor
// Notice that the caller should handle the output channel slice
func NewFanOutAsyncExecutor() *FanOutAsyncExecutor {
	return &FanOutAsyncExecutor{}
}

func (f *FanOutAsyncExecutor) Run(ctx context.Context,
	in <-chan interface{},
	executors []Executor) ([]<-chan interface{}, []<-chan error) {

	dataChs := make([]<-chan interface{}, 0, len(executors))
	errChs := make([]<-chan error, 0, len(executors))

	for _, e := range executors {
		ch1, ch2 := e.Run(ctx, in)
		dataChs = append(dataChs, ch1)
		errChs = append(errChs, ch2)
	}

	return dataChs, errChs
}

func NewAsyncExecutor(h Handler) *AsyncExecutor {
	return &AsyncExecutor{h}
}

// SyncJobExecutor deal with a bunch of input data
// Buf make sure that those input data belongs to the same Request
// Because only one context associated with the Executor
// And there's supposed to be only one context bound to one Request
func (e *AsyncExecutor) Run(ctx context.Context, in <-chan interface{}) (<-chan interface{}, <-chan error) {
	cancelCtx, _ := context.WithCancel(ctx)
	outErrCh := make(chan error)
	outDataCh := make(chan interface{}) //, DefaultChannelSize)

	go func() {
		for {
			select {
			// In channel deal with input request
			case args := <-in:
				result, err := e.H.Handle(cancelCtx, args)
				if err != nil {
					outErrCh <- err
					break
				}

				func() {
					// If support Closer Interface, we need to close it sooner or later
					if closer, ok := result.(Closer); ok {
						defer closer.Close()
					}

					// If this result is a iterator type , we need to handle or the possible result
					if iter, ok := result.(Iterator); ok {
						for {
							obj, err := iter.Next(cancelCtx)

							// EOF got, no more Next needed, so we send the last slice of data,
							// return and close the iterator
							if err == ErrIteratorEOF {
								if obj != nil {
									outDataCh <- obj
								}
								return
							}

							if err != nil {
								outErrCh <- err
								return
							}

							// Normal round, continue
							outDataCh <- obj
						}

						return
					}

					// Normal single result
					outDataCh <- result
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

	return outDataCh, outErrCh
}
