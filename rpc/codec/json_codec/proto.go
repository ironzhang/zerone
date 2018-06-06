package json_codec

import "encoding/json"

type clientRequest struct {
	ClassMethod string      `json:"ClassMethod"`
	Sequence    uint64      `json:"Sequence"`
	TraceID     string      `json:"TraceID"`
	ClientName  string      `json:"ClientName"`
	Verbose     int         `json:"Verbose,omitempty"`
	Body        interface{} `json:"Body,omitempty"`
}

type clientResponse struct {
	ClassMethod string          `json:"ClassMethod"`
	Sequence    uint64          `json:"Sequence"`
	Code        int             `json:"Code"`
	Desc        string          `json:"Desc,omitempty"`
	Cause       string          `json:"Cause,omitempty"`
	ServerName  string          `json:"ServerName,omitempty"`
	Body        json.RawMessage `json:"Body,omitempty"`
}

type serverRequest struct {
	ClassMethod string          `json:"ClassMethod"`
	Sequence    uint64          `json:"Sequence"`
	TraceID     string          `json:"TraceID"`
	ClientName  string          `json:"ClientName"`
	Verbose     int             `json:"Verbose,omitempty"`
	Body        json.RawMessage `json:"Body,omitempty"`
}

type serverResponse struct {
	ClassMethod string      `json:"ClassMethod"`
	Sequence    uint64      `json:"Sequence"`
	Code        int         `json:"Code"`
	Desc        string      `json:"Desc,omitempty"`
	Cause       string      `json:"Cause,omitempty"`
	ServerName  string      `json:"ServerName,omitempty"`
	Body        interface{} `json:"Body,omitempty"`
}
