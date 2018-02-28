package trace

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

type Trace interface {
	PrintRequest(body interface{}) error
	PrintResponse(err error, body interface{}) error
}

const timeLayout = "2006-01-02 15:04:05.999999 -0700 MST"

func printRequest(w io.Writer, now time.Time, traceID, clientName, serviceMethod string, data []byte) {
	fmt.Fprintf(w, "%s Request[%s][%s][%s]: %s\n", now.Format(timeLayout), traceID, clientName, serviceMethod, data)
}

func printError(w io.Writer, now time.Time, elapse time.Duration, traceID, clientName, serviceMethod string, err error) {
	fmt.Fprintf(w, "%s Error[%s][%s][%s][%s]: %s\n", now.Format(timeLayout), elapse, traceID, clientName, serviceMethod, err)
}

func printResult(w io.Writer, now time.Time, elapse time.Duration, traceID, clientName, serviceMethod string, data []byte) {
	fmt.Fprintf(w, "%s Result[%s][%s][%s][%s]: %s\n", now.Format(timeLayout), elapse, traceID, clientName, serviceMethod, data)
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
	printRequest(&t.buffer, t.start, t.traceID, t.clientName, t.serviceMethod, data)
	return nil
}

func (t *errorTrace) PrintResponse(err error, body interface{}) error {
	if err != nil {
		end := time.Now()
		t.out.Write(t.buffer.Bytes())
		printError(t.out, end, end.Sub(t.start), t.traceID, t.clientName, t.serviceMethod, err)
	}
	return nil
}

type verboseTrace struct {
	out           io.Writer
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
	printRequest(t.out, t.start, t.traceID, t.clientName, t.serviceMethod, data)
	return nil
}

func (t *verboseTrace) PrintResponse(err error, body interface{}) error {
	end := time.Now()
	if err != nil {
		printError(t.out, end, end.Sub(t.start), t.traceID, t.clientName, t.serviceMethod, err)
		return nil
	}
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	printResult(t.out, end, end.Sub(t.start), t.traceID, t.clientName, t.serviceMethod, data)
	return nil
}
