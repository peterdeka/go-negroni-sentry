package negronisentry

import (
	"errors"
	"fmt"
	"github.com/getsentry/raven-go"
	"log"
	"net/http"
	"os"
	"runtime"
)

// Recovery is a Negroni middleware that recovers from any panics and writes a 500 if there was one.
type Recovery struct {
	Logger     *log.Logger
	PrintStack bool
	StackAll   bool
	StackSize  int
	SentryCli  *raven.Client
}

// NewRecovery returns a new instance of Recovery
func NewRecovery(dsn string) *Recovery {
	logger := log.New(os.Stdout, "[negroni] ", 0)
	//init sentry client
	client, err := raven.NewClient(dsn, nil)
	if err != nil {
		logger.Fatal("FATAL: ", err)
	}
	return &Recovery{
		Logger:     logger,
		PrintStack: true,
		StackAll:   false,
		StackSize:  1024 * 8,
		SentryCli:  client,
	}
}

func (rec *Recovery) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	defer func() {
		if err := recover(); err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			stack := make([]byte, rec.StackSize)
			stack = stack[:runtime.Stack(stack, rec.StackAll)]

			f := "PANIC: %s\n%s"
			rec.Logger.Printf(f, err, stack)
			rec.SentryCli.CaptureError(errors.New(fmt.Sprintf("%v", err)), nil)
			if rec.PrintStack {
				fmt.Fprintf(rw, f, err, stack)
			}
		}
	}()

	next(rw, r)
}
