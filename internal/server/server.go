package server

import (
	"TCP_HTTP/internal/request"
	"TCP_HTTP/internal/response"
	"fmt"
	"net"
	"sync/atomic"
)



type Server struct{
	Listener net.Listener
	Closed atomic.Bool
	handler Handler
}

func Serve(port int, handler Handler) (*Server, error){
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	var server = &Server{
		Listener: listener,
		handler:  handler,
	}

	go server.listen()
	return server, nil
}

func (s *Server) Close() error {
	s.Closed.Store(true)
	if s.Listener != nil {
		return s.Listener.Close()
	}
	return nil
}

func (s *Server) listen() {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			if s.Closed.Load(){
				return
			}
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}

		fmt.Printf("Accepted connection from %s\n", conn.RemoteAddr())
		go s.handle(conn)
	}

}

func (s *Server) handle(conn net.Conn){
	defer conn.Close()
	var responseWriter = response.NewWriter(conn)
	parsedReq, err := request.RequestFromReader(conn)
	if err != nil{
		responseWriter.WriteStatusLine(response.BadRequest)
		var messageBytes = []byte(err.Error())
		var headers = response.GetDefaultHeaders(len(messageBytes))
		responseWriter.WriteHeaders(headers)
		responseWriter.WriteBody(messageBytes)
		return
	}

	s.handler(responseWriter, parsedReq)

	// var statusErr error = wr.WriteStatusLine(response.OK)
	// if statusErr != nil{
	// 	hErr := &HandlerError{
	// 		StatusCode: response.ServerError,
	// 		Message: statusErr.Error(),
	// 	}
	// 	hErr.WriteError(conn)
	// 	return
	// }
	// var headers = response.GetDefaultHeaders(len(buffer.Bytes()))
	// var headersErr error = wr.WriteHeaders(headers)
	// if headersErr != nil{
	// 	hErr := &HandlerError{
	// 		StatusCode: response.ServerError,
	// 		Message: headersErr.Error(),
	// 	}
	// 	hErr.WriteError(conn)
	// 	return
	// }
	// _, err = wr.WriteBody(buffer.Bytes())
	// if err != nil{
	// 	hErr := &HandlerError{
	// 		StatusCode: response.ServerError,
	// 		Message: err.Error(),
	// 	}
	// 	hErr.WriteError(conn)
	// 	return
	// }
}