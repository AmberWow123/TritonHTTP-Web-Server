package tritonhttp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	// Addr specifies the TCP address for the server to listen on,
	// in the form "host:port". It shall be passed to net.Listen()
	// during ListenAndServe().
	Addr string // e.g. ":0"

	// DocRoot specifies the path to the directory to serve static files from.
	DocRoot string
}

func (s *Server) ValidateServerSetup() error {
	fi, err := os.Stat(s.DocRoot)

	if os.IsNotExist(err) {
		return err
	}

	if !fi.IsDir() {
		return fmt.Errorf("doc root %q is not a directory", s.DocRoot)
	}

	return nil
}

// ListenAndServe listens on the TCP network address s.Addr and then
// handles requests on incoming connections.
func (s *Server) ListenAndServe() error {
	// panic("todo")

	// Validate the Server Configuration
	err := s.ValidateServerSetup()
	if err != nil {
		return err
	}

	// Listen on a port
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}

	fmt.Println("Listening on ... ", ln.Addr())

	// Accept connection and server them
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error in accepting connection", err)
			continue
		}
		fmt.Println("Accepting connection from", conn.RemoteAddr())
		go s.HandleConnection(conn)
	}

	// Hint: call HandleConnection
}

// HandleConnection reads requests from the accepted conn and handles them.
func (s *Server) HandleConnection(conn net.Conn) {
	// panic("todo")

	// fmt.Println("starting handleconnection...")
	br := bufio.NewReader(conn)
	// Hint: use the other methods below

	for {
		// Set timeout
		if err := conn.SetReadDeadline(time.Now().Add(time.Second * 5)); err != nil {
			fmt.Printf("Failed to set timeout for the conn %v", conn.RemoteAddr())
			_ = conn.Close()
			return
		}

		// Try to read next request
		req, bytesReceived, err := ReadRequest(br)

		fmt.Println("[server]-err:", err)
		// fmt.Println("[server]-bytesreceived:", bytesReceived)

		// fmt.Print("[server]: req.URL: ")
		// fmt.Println(req.URL)
		// // Handle invalid URL (starts with a /)
		// // if err == fmt.Errorf("URL doesn't start with a / character") {
		// if req.URL[0] != '/' {
		// 	// fmt.Println("[server]:", err)
		// 	fmt.Println("before")
		// 	res := &Response{}
		// 	fmt.Println("after")
		// 	res.HandleBadRequest()

		// 	_ = res.Write(conn)
		// 	_ = conn.Close()
		// 	return
		// }

		// Handle EOF
		if errors.Is(err, io.EOF) {
			fmt.Printf("Connection closed by the client %v", conn.RemoteAddr())
			_ = conn.Close()
			return
		}

		// Handle timeout
		// if err, ok := err.(net.Error); ok && err.Timeout() {
		if err, ok := err.(net.Error); ok && err != nil && err.Timeout() {
			fmt.Printf("Connection to %v timed out.\n", conn.RemoteAddr())
			res := &Response{}
			// signal a 400 error if a partial request is received
			if bytesReceived {

				// fmt.Println("Calling HandleBadRequest to signal a 400 error....")
				res.HandleBadRequest()
				_ = res.Write(conn)
				_ = conn.Close()
				return
			} else {
				// close the connection if no partial request is received
				// fmt.Println("if no partial req and timeout")
				// should I write back to response????????????
				// _ = res.Write(conn)

				_ = conn.Close()
			}
			return
		}

		// Handle not found

		// // check if file path exists
		// // if req == nil, then it is a bad request
		// if req != nil {

		// 	// check if starts with /, if so, 400 error
		// 	if (req.URL)[0] != '/' {
		// 		fmt.Println("URL doesn't start with a / character. -> HandleBadRequest")
		// 		res := &Response{}
		// 		res.HandleBadRequest()
		// 		_ = res.Write(conn)
		// 		_ = conn.Close()
		// 		return
		// 	}

		// 	// check if ends with /, if so, append "index.html"
		// 	if (req.URL)[len(req.URL)-1:] == "/" {
		// 		fmt.Println("URL ends with a /")
		// 		req.URL = req.URL + "index.html"
		// 		fmt.Println("URL now is -> " + req.URL)
		// 	}

		// 	// do filepath Clean
		// 	fmt.Println("Before Clean ->", s.DocRoot+req.URL)
		// 	file_path := filepath.Clean(s.DocRoot + req.URL)
		// 	fmt.Println("filepath.Clean ->", file_path)

		// 	// check if go beyond the docRoot, if so, 404 error

		// 	// check if file path exists, if so, 404 error
		// 	if _, err := os.Stat(file_path); errors.Is(err, os.ErrNotExist) {
		// 		fmt.Println("file path does not exist")
		// 		log.Printf("Handling Not Found for error: %v", err)
		// 		res := &Response{}
		// 		res.HandleNotFound(req)
		// 		_ = res.Write(conn)
		// 		_ = conn.Close()
		// 		return
		// 	}
		// }

		// if errors.Is(err, os.ErrNotExist) {
		// 	log.Printf("Handling Not Found for error: %v", err)
		// 	res := &Response{}
		// 	res.HandleNotFound(req)
		// 	_ = res.Write(conn)
		// 	_ = conn.Close()
		// 	return
		// }

		// Handle bad request
		if err != nil {
			fmt.Printf("Handling Bad Request for error : %v", err)
			res := &Response{}
			res.HandleBadRequest()
			_ = res.Write(conn)
			_ = conn.Close()
			return
		}

		// == URL ==
		// Handle bad request (check if URL starts with a /)
		// if not, 400 (bad request)
		if (req.URL)[0] != '/' {
			fmt.Println("URL doesn't start with a / character. -> HandleBadRequest")
			res := &Response{}
			res.HandleBadRequest()
			_ = res.Write(conn)
			_ = conn.Close()
			return
		}

		// fmt.Println("req.URL:", req.URL)
		// check if URL ends with a /; if so, + "index.html"
		if (req.URL)[len(req.URL)-1:] == "/" {
			req.URL = req.URL + "index.html"
		}
		// fmt.Println("now req.URL:", req.URL)

		// Handle good request
		fmt.Printf("Handle good request : %v", req)
		res := s.HandleGoodRequest(req)
		if err := res.Write(conn); err != nil {
			fmt.Println(err)
		}

		// Close conn if requested
		// if the connection is set to "close" from the request, close it here
		if req.Close {
			fmt.Println("let's close connection here....")
			conn.Close()
			return
		}

	}
}

// HandleGoodRequest handles the valid req and generates the corresponding res.
func (s *Server) HandleGoodRequest(req *Request) (res *Response) {
	// panic("todo")
	res = &Response{}
	// Hint: use the other methods below

	// ==check URL==
	// fmt.Println("req.URL:", req.URL)
	// check if URL ends with a /; if so, + "index.html"
	if (req.URL)[len(req.URL)-1:] == "/" {
		req.URL = req.URL + "index.html"
	}
	// fmt.Println("now req.URL:", req.URL)

	// do filepath.Clean()
	// fmt.Println("Before Clean():", s.DocRoot+req.URL)
	file_path := filepath.Clean(s.DocRoot + req.URL)
	// fmt.Println("After Clean():", file_path)

	// check if URL go beyond docRoot; if so, 404 (not found)
	if !strings.Contains(file_path, s.DocRoot) {
		// fmt.Println("Go beyond the docRoot...")
		res.HandleNotFound(req)
		return res
	}

	// fmt.Println("check URL:", file_path)
	// check if URL exists; if not, 404 (not found)
	if _, err := os.Stat(file_path); errors.Is(err, os.ErrNotExist) {
		fmt.Println("file path / dir does not exist")
		log.Printf("Handling Not Found for error: %v", err)
		res.HandleNotFound(req)
		// fmt.Println(res.StatusCode)
		return res
	}

	// check if reading a existing directory; if so, 404
	if stat, err := os.Stat(file_path); err == nil && stat.IsDir() {
		// path is a directory
		fmt.Println("try to read a dir")
		res.HandleNotFound(req)
		// fmt.Println(res.StatusCode)
		return res
	}

	// edit
	// local_f_path := filepath.Clean(req.URL)
	// fmt.Println("after Clean(req.URL)->", local_f_path)
	// res.HandleOK(req, local_f_path)
	// res.FilePath = file_path
	res.HandleOK(req, file_path)

	// res.Proto = "HTTP/1.1"
	// res.StatusCode = 200
	// res.FilePath = filepath.Join(s.DocRoot, req.URL)
	res.Request = req
	if res.Header == nil {
		res.Header = make(map[string]string, 0)
	}
	res.Header["Date"] = FormatTime(time.Now())

	// info, err := os.Stat(res.FilePath)
	info, err := os.Stat(file_path)
	if err != nil {
		fmt.Println("os.Stat error:", err)
		return res
	}
	// these 3 headers are only for statuscode 200
	res.Header["Last-Modified"] = FormatTime(info.ModTime())
	// res.Header["Content-Type"] = MIMETypeByExtension(filepath.Ext(res.FilePath))
	res.Header["Content-Type"] = MIMETypeByExtension(filepath.Ext(file_path))
	res.Header["Content-Length"] = strconv.FormatInt(info.Size(), 10)

	// res.Header["Host"] = req.Host
	if req.Close {
		res.Header["Connection"] = "close"
	}

	// fmt.Println(res.Proto)
	// fmt.Println(res.Header)
	// fmt.Println(res.Request)
	// fmt.Println(res.StatusCode)
	// fmt.Println(res.FilePath)
	return res
}

// HandleOK prepares res to be a 200 OK response
// ready to be written back to client.
func (res *Response) HandleOK(req *Request, path string) {
	// panic("todo")
	res.Proto = "HTTP/1.1"
	res.StatusCode = 200
	res.FilePath = path
	// fmt.Println("[HandleOK]:", res.FilePath)
}

// HandleBadRequest prepares res to be a 400 Bad Request response
// ready to be written back to client.
func (res *Response) HandleBadRequest() {
	// panic("todo")
	res.Proto = "HTTP/1.1"
	res.StatusCode = 400
	res.FilePath = ""
	// nil if not a valid request
	res.Request = nil
	if res.Header == nil {
		res.Header = make(map[string]string, 0)
	}
	res.Header["Date"] = FormatTime(time.Now())
	// not sure about this
	res.Header["Connection"] = "close"
}

// HandleNotFound prepares res to be a 404 Not Found response
// ready to be written back to client.
func (res *Response) HandleNotFound(req *Request) {
	// panic("todo")
	res.Proto = "HTTP/1.1"
	res.StatusCode = 404
	res.FilePath = ""
	// nil if not a valid request
	res.Request = nil
	if res.Header == nil {
		res.Header = make(map[string]string, 0)
	}
	res.Header["Date"] = FormatTime(time.Now())
	// not sure about this
	if req.Close {
		res.Header["Connection"] = "close"
	}

	// fmt.Println("[handleNotFound]-code:", res.StatusCode)

}
