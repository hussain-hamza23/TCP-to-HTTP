package server

import (
	"TCP_HTTP/internal/request"
	"TCP_HTTP/internal/response"
	"io"
)



type HandlerError struct{
	StatusCode response.StatusCode
	Message string
}


type Handler func(w *response.Writer, req *request.Request)

func (h *HandlerError) WriteError(w io.Writer){
	var wr = response.Writer{Writer: w}
	wr.WriteStatusLine(h.StatusCode)
	var messageBytes = []byte(h.Message)
	var headers = response.GetDefaultHeaders(len(messageBytes))
	wr.WriteHeaders(headers)
	wr.WriteBody(messageBytes)
}