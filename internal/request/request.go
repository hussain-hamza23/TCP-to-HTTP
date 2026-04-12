package request

import (
	"TCP_HTTP/internal/headers"
	"bytes"
	"errors"
	"io"
	"strconv"
	"strings"
	"unicode"
)

type State int

const bufferSize int = 1
const CRLF string = "\r\n"
const (
	initialized State = iota
	parsingHeaders
	parsingBody
	done
)

type Request struct{
	RequestLine RequestLine
	Headers headers.Headers
	Body []byte
	State State
}

type RequestLine struct{
	HTTPVersion string
	RequestTarget string
	Method string
}



func RequestFromReader(reader io.Reader) (*Request, error){
	var buffer = make([]byte, bufferSize)
	var req = &Request{
		Headers: headers.NewHeaders(),
		State: initialized,
		Body: make([]byte, 0),
	}
	var idxRead int = 0
	for req.State != done{
		if idxRead >= len(buffer){
			// If the buffer is full and we haven't found the end of the request line, we need to increase the buffer size
			newBuffer := make([]byte, len(buffer)*2)
			copy(newBuffer, buffer)
			buffer = newBuffer
		}

		
		//parse the request data in the buffer
		bytesParsed, err := req.parseRequestData(buffer[:idxRead])
		if err != nil{
			return nil, err
		}

		if bytesParsed == 0 {
			if req.State == done{
				break
			}
			// If no bytes were parsed, we need to read more data
			//read into buffer starting from the last read position
			bytesRead, err := reader.Read(buffer[idxRead:])
			if err != nil{
				if errors.Is(err, io.EOF){
					if req.State != done{
						return nil, errors.New("Unexpected end of request while parsing headers")
					}
					break
				}
				return nil, err
			}
			idxRead += bytesRead
		}
		copy(buffer, buffer[bytesParsed:])
		idxRead -= bytesParsed
	}
	return req, nil
}

func (req *Request)parseRequestData(data []byte) (int, error){

	//Check for CRLF in the request
	var idx int
	if req.State != parsingBody{
		idx = bytes.Index(data, []byte(CRLF))
		if idx == -1{
			return 0, nil
		}
	}
	switch req.State{

	case initialized:
		//Split only the request line
		
		var requestHeader []string = strings.Split(string(data[:idx]), CRLF)
		requestLine, err := parseRequestLine(requestHeader[0])
		if err != nil{
			return 0, err
		}
		req.RequestLine = *requestLine
		req.State = parsingHeaders
		return idx + 2, nil
	case parsingHeaders:
		idx, parsed, err := req.Headers.Parse(data)
		if err != nil{
			return 0, err
		}
		if parsed{
			req.State = parsingBody
		}
		return idx, nil
	case parsingBody:
		body, err := req.parseBody(data)
		if err != nil{
			return 0, err
		}
		return body, nil

	default:
		return 0, errors.New("Invalid request state")
	}
}

func parseRequestLine(line string) (*RequestLine, error) {
	var parts []string = strings.Split(line, " ")
	if len(parts) != 3{
		return nil, errors.New("Invalid request line: " + line)
	}

	for _, part := range parts{
		if part == ""{
			return nil, errors.New("Request line parts cannot be empty: " + line)
		}
	}

	var method string = parts[0]
	var requestTarget string = parts[1]
	var httpVersion string = strings.TrimPrefix(parts[2], "HTTP/")



	for _, r := range method{
		if !unicode.IsLetter(r) || !unicode.IsUpper(r){
			return nil, errors.New("Invalid method format: " + method)
		}
	}

	if httpVersion != "1.1" {
		return nil, errors.New("Unsupported HTTP version: " + httpVersion)
	}

	return &RequestLine{
		HTTPVersion: httpVersion,
		RequestTarget: requestTarget,
		Method: method,
	}, nil
}

func (req *Request) parseBody(data []byte) (int, error){
	lengthStr, exists := req.Headers.Get("Content-Length") 
	if !exists{
		req.State = done
		return len(data), nil
	}
	length, err := strconv.Atoi(lengthStr)
	if err != nil{
		return 0, errors.New("Invalid Content-Length header")
	}

	req.Body = append(req.Body, data...)

	if len(req.Body) > length{
		return 0, errors.New("Body length exceeds Content-Length")
	}

	if len(req.Body) == length{
		req.State = done
	}

	return len(data), nil
}
