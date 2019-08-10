package tracehelper

import (
	"context"
	"log"
	"net/http"
	_ "net/http/pprof" //init http server for trace

	"os"
	"os/exec"
	"os/signal"
	"runtime/trace"
	"strconv"
	"syscall"
	"time"
)

var seqCh = make(chan uint64)
var startCh = make(chan string)
var stopCh = make(chan struct{})

func init() {
	go func() {
		var seq uint64
		for {
			seqCh <- seq
			seq++
		}
	}()
}

//WithSwitch start trace by function call, must with a StopTrace call later
func WithSwitch(name string, start bool) (SwitchFunc, SwitchFunc) {

	switchCh := make(chan struct{})
	go func() {

		if "" == name {
			name = "WithSwitch"
		}
		f := newTraceFile(name)
		if false == start {
			<-switchCh
		}
		err := trace.Start(f)
		if nil != err {
			delTraceFile(f)
			return
		}
		<-switchCh
		trace.Stop()

		runTraceUI(f)

	}()
	stopswitch := func() {
		close(switchCh)
	}

	startswitch := func() {
		if start == false {
			var empty struct{}
			switchCh <- empty
			start = true
		}
	}

	return startswitch, stopswitch

}

//SwitchFunc stop trace by function call,  along with a StartTrace call
type SwitchFunc func()

//WithHTTP trigger trace by http request
func WithHTTP(port string) {
	go func() {
		// For security, the ports of traceserver just open for the localhost users
		http.ListenAndServe("localhost"+port, http.DefaultServeMux)
	}()
}

//WithSignal trigger trace by send signal
func WithSignal(duration time.Duration) {
	go func() {

		for {
			sigch := make(chan os.Signal, 1)
			signal.Notify(sigch, syscall.SIGQUIT)
			<-sigch
			name := "WithSignal"

			f := newTraceFile(name)

			err := trace.Start(f)
			if nil != err {
				delTraceFile(f)
				return
			}
			if 0 == duration {
				duration = 10 * time.Second
			}
			<-time.After(duration)
			trace.Stop()

			runTraceUI(f)
		}
	}()
}

//WithContext trace with scope limit by context
func WithContext(ctx context.Context, name string) {
	if "" == name {
		name = "WithContext"
	}

	f := newTraceFile(name)
	go func() {
		err := trace.Start(f)
		if nil != err {
			delTraceFile(f)
			return
		}
		<-ctx.Done()
		trace.Stop()

		runTraceUI(f)
	}()
}

func delTraceFile(f *os.File) {
	fname := f.Name()
	f.Close()
	os.Remove(fname)
}

func runTraceUI(f *os.File) {
	// close tracefile
	if err := f.Close(); err != nil {
		log.Fatalf("Cannot close new temp profile file: %v", err)
	}
	// Open tracefile with trace
	log.Printf("Starting go tool trace %v", f.Name())
	cmd := exec.Command("go", "tool", "trace", f.Name())
	if err := cmd.Run(); err != nil {
		log.Printf("Cannot start trace UI: %v", err)
	}
}

func newTraceFile(name string) (f *os.File) {
	seq := <-seqCh
	fname := "trace-" + strconv.Itoa(os.Getpid()) + "-" + name + "-" + strconv.FormatUint(seq, 10)
	f, err := os.Create(fname)
	if err != nil {
		log.Fatalf("Cannot create new temp profile file: %v", err)
	}
	return f
}
