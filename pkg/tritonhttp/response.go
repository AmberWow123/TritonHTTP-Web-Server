package tritonhttp

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
)

type Response struct {
	StatusCode int    // e.g. 200
	Proto      string // e.g. "HTTP/1.1"

	// Header stores all headers to write to the response.
	// Header keys are case-incensitive, and should be stored
	// in the canonical format in this map.
	Header map[string]string

	// Request is the valid request that leads to this response.
	// It could be nil for responses not resulting from a valid request.
	Request *Request

	// FilePath is the local path to the file to serve.
	// It could be "", which means there is no file to serve.
	FilePath string
}

var statusText = map[int]string{
	200: "OK",
	400: "Bad Request",
	404: "Not Found",
}

// Write writes the res to the w.
func (res *Response) Write(w io.Writer) error {
	if err := res.WriteStatusLine(w); err != nil {
		return err
	}
	if err := res.WriteSortedHeaders(w); err != nil {
		return err
	}
	if err := res.WriteBody(w); err != nil {
		return err
	}
	return nil
}

// WriteStatusLine writes the status line of res to w, including the ending "\r\n".
// For example, it could write "HTTP/1.1 200 OK\r\n".
func (res *Response) WriteStatusLine(w io.Writer) error {
	// panic("todo")
	bw := bufio.NewWriter(w)
	statusLine := fmt.Sprintf("%v %v %v\r\n", res.Proto, res.StatusCode, statusText[res.StatusCode])

	if _, err := bw.WriteString(statusLine); err != nil {
		return err
	}

	if err := bw.Flush(); err != nil {
		return err
	}

	return nil
}

// WriteSortedHeaders writes the headers of res to w, including the ending "\r\n".
// For example, it could write "Connection: close\r\nDate: foobar\r\n\r\n".
// For HTTP, there is no need to write headers in any particular order.
// TritonHTTP requires to write in sorted order for the ease of testing.
func (res *Response) WriteSortedHeaders(w io.Writer) error {
	// panic("todo")
	keys := make([]string, 0, len(res.Header))
	for k := range res.Header {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	bw := bufio.NewWriter(w)
	for _, k := range keys {
		headerLine := fmt.Sprintf("%v: %v\r\n", k, res.Header[k])
		if _, err := bw.WriteString(headerLine); err != nil {
			return err
		}
	}

	if _, err := bw.WriteString("\r\n"); err != nil {
		return err
	}

	if err := bw.Flush(); err != nil {
		return err
	}

	return nil

}

// WriteBody writes res' file content as the response body to w.
// It doesn't write anything if there is no file to serve.
func (res *Response) WriteBody(w io.Writer) error {
	// panic("todo")

	// nothing to do when there is no file to serve
	if res.FilePath != "" {
		// read file
		file, err := os.Open(res.FilePath)
		if err != nil {
			fmt.Println("Open file error:", err)
			return err
		}
		defer file.Close()

		fileinfo, err := file.Stat()
		if err != nil {
			fmt.Println(err)
			return err
		}

		filesize := fileinfo.Size()
		buffer := make([]byte, filesize)

		bytesread, err := file.Read(buffer)
		if err != nil {
			fmt.Println("Read file error:", err)
			return err
		}
		fmt.Println("bytes read: ", bytesread)
		// fmt.Println("Read file content from request:", string(buffer))

		// write the whole content to response
		bw := bufio.NewWriter(w)
		// fileContent := fmt.Sprintf("%v\r\n", string(buffer))
		fileContent := fmt.Sprintf("%v", string(buffer))

		if _, err := bw.WriteString(fileContent); err != nil {
			return err
		}

		if err := bw.Flush(); err != nil {
			return err
		}

		return nil
	}

	return nil
}
