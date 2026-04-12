package response

import (
	"TCP_HTTP/internal/headers"
	"fmt"
	"io"
)


type StatusCode int

const (
	OK StatusCode = 200
	BadRequest StatusCode = 400
	ServerError StatusCode = 500
)


type writerState int
const (
	writeStatusLine writerState = iota
	writeHeaders
	writeBody
	writeTrailers
)

type Writer struct{
	Writer io.Writer
	State writerState
}

func NewWriter(w io.Writer) *Writer{
	return &Writer{
		Writer: w,
		State: writeStatusLine,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error{
	if w.State != writeStatusLine{
		return fmt.Errorf("Invalid state: expected writeStatusLine, got %v", w.State)
	}
	var statusText string
	switch statusCode{
	case OK:
		statusText = "HTTP/1.1 200 OK"
	case BadRequest:
		statusText = "HTTP/1.1 400 Bad Request"
	case ServerError:
		statusText = "HTTP/1.1 500 Internal Server Error"
	default:
		statusText = ""
	}
	_, err := w.Writer.Write([]byte(statusText + "\r\n"))
	if err != nil {
		return err
	}
	w.State = writeHeaders
	return nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error{
	
	if w.State != writeHeaders{
		return fmt.Errorf("Invalid state: expected writeHeaders, got %v", w.State)
	}

	for key, value := range headers{
		headerLine := fmt.Sprintf("%s: %s", key, value)
		_, err := w.Writer.Write([]byte(headerLine + "\r\n"))
		if err != nil {
			return err
		}
	}
	_, err := w.Writer.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	w.State = writeBody
	return nil
}

func (w *Writer) WriteBody(body []byte) (int, error){
	if w.State != writeBody{
		return 0, fmt.Errorf("Invalid state: expected writeBody, got %v", w.State)
	}
	n, err := w.Writer.Write(body)
	if err != nil {
		return n, err
	}
	w.State = writeTrailers
	return n, nil
}

// func WriteStatusLine(w io.Writer, statusCode StatusCode) error{
// 	var statusText string
// 	switch statusCode{
// 	case OK:
// 		statusText = "HTTP/1.1 200 OK"
// 	case BadRequest:
// 		statusText = "HTTP/1.1 400 Bad Request"
// 	case ServerError:
// 		statusText = "HTTP/1.1 500 Internal Server Error"
// 	default:
// 		statusText = ""
// 	}
// 	_, err := w.Write([]byte(statusText + "\r\n"))
// 	return err
// }

func GetDefaultHeaders(contentLength int) headers.Headers{
	var headers = headers.NewHeaders()
	headers.Add("Content-Type", "text/html")
	headers.Add("Connection", "close")
	headers.Add("Content-Length", fmt.Sprintf("%d", contentLength))
	return headers
	// headers["Content-Type"] = "text/html"
	// headers["Connection"] = "close"
	// headers["Content-Length"] = fmt.Sprintf("%d", contentLength)
	// return headers

}

// func WriteHeaders(w io.Writer, headers headers.Headers) error{
// 	for key, value := range headers{
// 		headerLine := fmt.Sprintf("%s: %s", key, value)
// 		_, err := w.Write([]byte(headerLine + "\r\n"))
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	w.Write([]byte("\r\n"))
// 	return nil
// }

func (w *Writer) WriteChunkedBody(body []byte) (int, error){
	if w.State != writeBody{
		return 0, fmt.Errorf("Invalid state: expected writeBody, got %v", w.State)
	}
	var hexSize int = len(body)
	var nTotal int = 0

	n, err := fmt.Fprintf(w.Writer, "%x\r\n", hexSize)
	if err != nil {
		return nTotal, err
	}
	nTotal += n
	n, err = w.Writer.Write(body)
	if err != nil {
		return nTotal, err
	}
	nTotal += n
	_, err = w.Writer.Write([]byte("\r\n"))
	if err != nil {
		return nTotal, err
	}
	nTotal += 2
	return nTotal, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	var doneChunk = []byte("0\r\n")
	n, err := w.Writer.Write(doneChunk)
	if err != nil {
		return n, err
	}
	w.State = writeTrailers
	return n, nil
}


func (w *Writer) WriteTrailers(trailers headers.Headers) error{
	if w.State != writeTrailers{
		return fmt.Errorf("Invalid state: expected writeTrailers, got %v", w.State)
	}
	for key, value := range trailers{
		_, err := fmt.Fprintf(w.Writer,"%s: %s\r\n", key, value)
		if err != nil {
			return err
		}
	}
	_, err := w.Writer.Write([]byte("\r\n"))
	return err
}
