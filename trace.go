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

	go func() {
		for {
			name := <-startCh
			if "" == name {
				name = "ChanControl"
			}
			f := newTraceFile(name)

			err := trace.Start(f)
			if nil != err {
				delTraceFile(f)
				return
			}
			<-stopCh
			trace.Stop()

			runTraceUI(f)
		}
	}()
}

//StartTrace start trace by function call, must with a StopTrace call later
func StartTrace(name string) {
	startCh <- name
}

//StopTrace stop trace by function call,  along with a StartTrace call
func StopTrace() {
	var empty struct{}
	stopCh <- empty
}

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

//WithCancel attach the trace with a cancel context
func WithCancel(name string) (tracectx context.Context, tracecancel context.CancelFunc) {
	if "" == name {
		name = "WithCancel"
	}

	tracectx, tracecancel = context.WithCancel(context.Background())
	f := newTraceFile(name)
	go func() {
		err := trace.Start(f)
		if nil != err {
			delTraceFile(f)
			return
		}
		<-tracectx.Done()
		trace.Stop()

		runTraceUI(f)
	}()
	return tracectx, tracecancel

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
