package json_codec

import "encoding/json"

type clientRequest struct {
	ServiceMethod string      `json:"ServiceMethod"`
	Sequence      uint64      `json:"Sequence"`
	TraceID       string      `json:"TraceID"`
	ClientName    string      `json:"ClientName"`
	Verbose       int         `json:"Verbose,omitempty"`
	Body          interface{} `json:"Body,omitempty"`
}

type clientResponse struct {
	ServiceMethod string          `json:"ServiceMethod"`
	Sequence      uint64          `json:"Sequence"`
	Code          int             `json:"Code"`
	Desc          string          `json:"Desc,omitempty"`
	Cause         string          `json:"Cause,omitempty"`
	Module        string          `json:"Module,omitempty"`
	Body          json.RawMessage `json:"Body,omitempty"`
}

type serverRequest struct {
	ServiceMethod string          `json:"ServiceMethod"`
	Sequence      uint64          `json:"Sequence"`
	TraceID       string          `json:"TraceID"`
	ClientName    string          `json:"ClientName"`
	Verbose       int             `json:"Verbose,omitempty"`
	Body          json.RawMessage `json:"Body,omitempty"`
}

type serverResponse struct {
	ServiceMethod string      `json:"ServiceMethod"`
	Sequence      uint64      `json:"Sequence"`
	Code          int         `json:"Code"`
	Desc          string      `json:"Desc,omitempty"`
	Cause         string      `json:"Cause,omitempty"`
	Module        string      `json:"Module,omitempty"`
	Body          interface{} `json:"Body,omitempty"`
}
