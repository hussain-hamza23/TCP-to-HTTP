package main

import (
	"TCP_HTTP/internal/headers"
	"TCP_HTTP/internal/request"
	"TCP_HTTP/internal/response"
	"TCP_HTTP/internal/server"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const port = 42069

func main() {
	server, err := server.Serve(port, videoHandler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func videoHandler(w *response.Writer, req *request.Request){
	if req.RequestLine.RequestTarget != "/video"{
		w.WriteStatusLine(response.BadRequest)
		var html string = `<html>
			<head>
				<title>400 Bad Request</title>
			</head>
			<body>
				<h1>Bad Request</h1>
				<p>Only /video is supported.</p>
			</body>
		</html>`
		var headers = response.GetDefaultHeaders(len([]byte(html)))
		w.WriteHeaders(headers)
		w.WriteBody([]byte(html))
		return
	}

	const videoPath string = "assets/vim.mp4"
	videoFile, err := os.ReadFile(videoPath)
	if err != nil {
		w.WriteStatusLine(response.ServerError)
		var html string = `<html>
			<head>
				<title>500 Internal Server Error</title>
			</head>
			<body>
				<h1>Internal Server Error</h1>
				<p>Failed to read video file.</p>
			</body>
		</html>`
		var headers = response.GetDefaultHeaders(len([]byte(html)))
		w.WriteHeaders(headers)
		w.WriteBody([]byte(html))
		return
	}

	w.WriteStatusLine(response.OK)
	var headers = response.GetDefaultHeaders(len(videoFile))
	headers.Override(strings.ToLower("Content-Type"), "video/mp4")
	fmt.Printf("Headers: %v\n", headers)
	w.WriteHeaders(headers)
	w.WriteBody(videoFile)
}

func httpHandler(w *response.Writer, req *request.Request) {
	var title string
	var h1 string
	var body string
	var responseCode response.StatusCode

	if 	strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/html") {
		proxyHandler(w, req)
		return
	}

	switch req.RequestLine.RequestTarget{
	case "/yourproblem":
		responseCode = response.BadRequest
		title = "400 Bad Request"
		h1 = "Bad Request"
		body = "Your request honestly kinda sucked."
	case "/myproblem":
		responseCode = response.ServerError
		title = "500 Internal Server Error"
		h1 = "Internal Server Error"
		body = "Okay, you know what? This one is on me."
	default:
		responseCode = response.OK
		title = "200 OK"
		h1 = "Success!"
		body = "Your request was an absolute banger."
	}

	var html string = fmt.Sprintf(
	`<html>
		<head>
			<title>%s</title>
		</head>
		<body>
			<h1>%s</h1>
			<p>%s</p>
		</body>
	</html>`, title, h1, body)

	w.WriteStatusLine(responseCode)
	var headers = response.GetDefaultHeaders(len([]byte(html)))
	w.WriteHeaders(headers)
	w.WriteBody([]byte(html))
		
}

func proxyHandler(w *response.Writer, req *request.Request){
	const domain string = "https://httpbin.org/"
	const bufferSize int = 1024
	var targetURL string = domain + strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
	resp, err := http.Get(targetURL)
	if err != nil {
		w.WriteStatusLine(response.ServerError)
		var html string = `<html>
			<head>
				<title>500 Internal Server Error</title>
			</head>
			<body>
				<h1>Internal Server Error</h1>
				<p>Failed to fetch data from httpbin.org.</p>
			</body>
		</html>`
		var headers = response.GetDefaultHeaders(len([]byte(html)))
		w.WriteHeaders(headers)
		w.WriteBody([]byte(html))
		return
	}
	defer resp.Body.Close()

	var h = response.GetDefaultHeaders(0)
	h.Remove("Content-Length")
	h.Add("Transfer-Encoding", "chunked")
	h.Add("Trailer", "X-Content-SHA256", "X-Content-Length")
	w.WriteStatusLine(response.OK)
	w.WriteHeaders(h)

	var buffer = make([]byte, bufferSize)
	var bodyBytes []byte
	for {
		bytes, err := resp.Body.Read(buffer)

		if bytes > 0 {
			bodyBytes = append(bodyBytes, buffer[:bytes]...)
			_, ok := w.WriteChunkedBody(buffer[:bytes])
			if ok != nil {
				log.Printf("Error writing chunked body")
				break
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF){
				_, err := w.WriteChunkedBodyDone()
				if err != nil {
					log.Printf("Error writing chunked body done: %v", err)
				}
			}
			break
		}
	}
	var hashedBody [32]byte = sha256.Sum256(bodyBytes)
	var length int = len(bodyBytes)
	trailers := headers.NewHeaders()
	trailers.Add("X-Content-SHA256", fmt.Sprintf("%x", hashedBody))
	trailers.Add("X-Content-Length", fmt.Sprintf("%d", length))
	w.WriteTrailers(trailers)
}
