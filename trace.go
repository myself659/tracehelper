package tracehelper

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof" //init http server for trace

	"sync"

	"os"
	"os/exec"
	"os/signal"
	"runtime/trace"
	"strconv"
	"syscall"
	"time"
)

var seqCh = make(chan uint64)

func init() {
	go func() {
		var seq uint64
		for {
			seqCh <- seq
			seq++
		}
	}()
}

// WithSwitch help to start or stop trace by calling SwitchFunc
// name: trace name
// start: when start is true, start stace at once, when start is false, must call startswitch to start trace
// startswitch: function for startting trace
// stopswitch:  function for stopping trace
func WithSwitch(name string, start bool) (startswitch SwitchFunc, stopswitch SwitchFunc) {

	switchCh := make(chan struct{})
	if "" == name {
		name = "WithSwitch"
	}
	var f *os.File
	if true == start {
		f = newTraceFile(name)
		err := trace.Start(f)
		if nil != err {
			log.Printf("Cannot trace for %s: %v", f.Name(), err)
			delTraceFile(f)
			emptySwitch := func() {}
			return emptySwitch, emptySwitch
		}
	}

	go func() {
		if false == start {
			f = newTraceFile(name)
			_, ok := <-switchCh
			if false == ok {
				delTraceFile(f)
				return
			}
			err := trace.Start(f)
			if nil != err {
				log.Printf("Cannot trace for %s: %v", f.Name(), err)
				delTraceFile(f)
				return
			}
		}

		<-switchCh
		trace.Stop()
		runTraceUI(f)

	}()

	stopswitch = func() {
		close(switchCh)
	}

	if start == false {
		var doOnce sync.Once
		startswitch = func() {
			doOnce.Do(func() {
				var empty struct{}
				switchCh <- empty
			})
		}
	} else {
		startswitch = func() {}
	}

	return startswitch, stopswitch
}

//SwitchFunc start or stop trace function type
type SwitchFunc func()

//FilterFunc filter for trace
type FilterFunc func() bool

// WithFilter help trace with a  filter function
// name: trace name
// filter: filter function
// stoptrace: function for stopping trace
func WithFilter(name string, filter FilterFunc) (stoptrace SwitchFunc) {

	if true == filter() {
		switchCh := make(chan struct{})
		if "" == name {
			name = "WithFilter"
		}
		f := newTraceFile(name)
		err := trace.Start(f)
		if nil != err {
			log.Printf("Cannot trace for %s: %v", f.Name(), err)
			delTraceFile(f)
			stoptrace = func() {}
			return stoptrace
		}
		go func() {
			<-switchCh
			trace.Stop()
			runTraceUI(f)
		}()
		stoptrace = func() { close(switchCh) }
		return stoptrace
	}

	return func() {}
}

//WithHTTP trigger trace by http request
func WithHTTP(port string) {
	go func() {
		// For security, the ports of traceserver just open for the localhost users
		http.ListenAndServe("localhost"+port, http.DefaultServeMux)
	}()
}

//WithSignal trigger trace by signal
func WithSignal(duration time.Duration) {
	go func() {
		name := "WithSignal"
		sigch := make(chan os.Signal, 1)
		signal.Notify(sigch, syscall.SIGQUIT) // todo: support windows
		fmt.Println("Send SIGQUIT (CTRL+\\) to the process to capture...")
		for {
			<-sigch
			f := newTraceFile(name)
			err := trace.Start(f)
			if nil != err {
				log.Printf("Cannot trace for %s: %v", f.Name(), err)
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

//WithContext help trace with scope limit by context
func WithContext(ctx context.Context, name string) {
	if "" == name {
		name = "WithContext"
	}

	f := newTraceFile(name)
	go func() {
		err := trace.Start(f)
		if nil != err {
			log.Printf("Cannot trace for %s: %v", f.Name(), err)
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
