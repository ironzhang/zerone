package rpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

type trace interface {
	PrintRequest(body interface{}) error
	PrintResponse(err error, body interface{}) error
}

const timeLayout = "2006-01-02 15:04:05.999999 -0700 MST"

func printRequest(w io.Writer, now time.Time, prefix, traceID, clientName, serviceMethod string, data []byte) {
	fmt.Fprintf(w, "%s %s.Request[%s][%s][%s]: %s\n", now.Format(timeLayout), prefix, traceID, clientName, serviceMethod, data)
}

func printError(w io.Writer, now time.Time, elapse time.Duration, prefix, traceID, clientName, serviceMethod string, err error) {
	fmt.Fprintf(w, "%s %s.Error[%s][%s][%s][%s]: %s\n", now.Format(timeLayout), prefix, elapse, traceID, clientName, serviceMethod, err)
}

func printResult(w io.Writer, now time.Time, elapse time.Duration, prefix, traceID, clientName, serviceMethod string, data []byte) {
	fmt.Fprintf(w, "%s %s.Result[%s][%s][%s][%s]: %s\n", now.Format(timeLayout), prefix, elapse, traceID, clientName, serviceMethod, data)
}

type nopTrace struct {
}

func (t nopTrace) PrintRequest(body interface{}) error {
	return nil
}

func (t nopTrace) PrintResponse(err error, body interface{}) error {
	return nil
}

type errorTrace struct {
	out           io.Writer
	prefix        string
	traceID       string
	clientName    string
	serviceMethod string
	start         time.Time
	buffer        bytes.Buffer
}

func (t *errorTrace) PrintRequest(body interface{}) error {
	t.start = time.Now()
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	printRequest(&t.buffer, t.start, t.prefix, t.traceID, t.clientName, t.serviceMethod, data)
	return nil
}

func (t *errorTrace) PrintResponse(err error, body interface{}) error {
	if err != nil {
		end := time.Now()
		t.out.Write(t.buffer.Bytes())
		printError(t.out, end, end.Sub(t.start), t.prefix, t.traceID, t.clientName, t.serviceMethod, err)
	}
	return nil
}

type verboseTrace struct {
	out           io.Writer
	prefix        string
	traceID       string
	clientName    string
	serviceMethod string
	start         time.Time
}

func (t *verboseTrace) PrintRequest(body interface{}) error {
	t.start = time.Now()
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	printRequest(t.out, t.start, t.prefix, t.traceID, t.clientName, t.serviceMethod, data)
	return nil
}

func (t *verboseTrace) PrintResponse(err error, body interface{}) error {
	end := time.Now()
	if err != nil {
		printError(t.out, end, end.Sub(t.start), t.prefix, t.traceID, t.clientName, t.serviceMethod, err)
		return nil
	}
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	printResult(t.out, end, end.Sub(t.start), t.prefix, t.traceID, t.clientName, t.serviceMethod, data)
	return nil
}

type traceLogger struct {
	out     io.Writer
	verbose int
}

func (p *traceLogger) SetOutput(out io.Writer) {
	if out == nil {
		p.out = os.Stdout
	} else {
		p.out = out
	}
}

func (p *traceLogger) SetVerbose(verbose int) {
	p.verbose = verbose
}

func (p *traceLogger) NewTrace(prefix string, verbose int, traceID, clientName, serviceMethod string) trace {
	v := max(p.verbose, verbose)
	if v < 0 {
		return nopTrace{}
	} else if v == 0 {
		return &errorTrace{
			out:           p.out,
			prefix:        prefix,
			traceID:       traceID,
			clientName:    clientName,
			serviceMethod: serviceMethod,
		}
	} else {
		return &verboseTrace{
			out:           p.out,
			prefix:        prefix,
			traceID:       traceID,
			clientName:    clientName,
			serviceMethod: serviceMethod,
		}
	}
}

func max(x, y int) int {
	if x > y {
		return x
	} else {
		return y
	}
}
