package main

import (
	"TCP_HTTP/internal/request"
	"fmt"
	"net"
)

const port = ":42069"

func main(){
	listner, err := net.Listen("tcp", port)
	if err != nil{
		fmt.Printf("Error starting server: %v\n", err)
		return
	}
	defer listner.Close()
	fmt.Println("Server is listening on port 42069...")

	for {
		conn, err := listner.Accept()
		if err != nil{
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}
		fmt.Printf("Accepted connection from %s\n", conn.RemoteAddr())
		reader, err := request.RequestFromReader(conn)
		if err != nil{
			fmt.Printf("Error reading request: %v\n", err)
			continue
		}
		fmt.Printf("Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n", reader.RequestLine.Method, reader.RequestLine.RequestTarget, reader.RequestLine.HTTPVersion)
		fmt.Printf("Headers:\n")
		for key, value := range reader.Headers {
			fmt.Printf("- %s: %s\n", key, value)
		}
		fmt.Printf("Body:\n%s\n", string(reader.Body))
		conn.Close()
		fmt.Printf("Connection from %s closed\n", conn.RemoteAddr())
		
	}
}


// func getLinesChannel(file io.ReadCloser) <-chan string {
// 	var lines = make(chan string)
	
// 	go func(){

// 		defer file.Close()	
// 		defer close(lines)
// 		var buffer = make([]byte, 8)
// 		var currentLine string = ""

// 		for {
// 			n, err := file.Read(buffer)
// 			if err != nil{
// 				if currentLine != ""{
// 					lines <- currentLine
// 					currentLine = ""
// 				}
// 				if errors.Is(err, io.EOF){
// 					break
// 				}
// 				fmt.Printf("Error reading file: %v\n", err)
// 				break
// 			}
// 			var chunk []string = strings.Split(string(buffer[:n]), "\n")
// 			for _, b := range chunk[:len(chunk) - 1]{
// 				lines <- currentLine + b
// 				currentLine = ""
// 			}
// 			currentLine += chunk[len(chunk) - 1]
// 		}
// 	}()
// 	return lines
// }