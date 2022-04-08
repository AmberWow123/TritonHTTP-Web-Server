package tritonhttp

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
)

type Request struct {
	Method string // e.g. "GET"
	URL    string // e.g. "/path/to/a/file"
	Proto  string // e.g. "HTTP/1.1"

	// Header stores misc headers excluding "Host" and "Connection",
	// which are stored in special fields below.
	// Header keys are case-incensitive, and should be stored
	// in the canonical format in this map.
	Header map[string]string

	Host  string // determine from the "Host" header
	Close bool   // determine from the "Connection" header
}

// ReadRequest tries to read the next valid request from br.
//
// If it succeeds, it returns the valid request read. In this case,
// bytesReceived should be true, and err should be nil.
//
// If an error occurs during the reading, it returns the error,
// and a nil request. In this case, bytesReceived indicates whether or not
// some bytes are received before the error occurs. This is useful to determine
// the timeout with partial request received condition.
func ReadRequest(br *bufio.Reader) (req *Request, bytesReceived bool, err error) {
	// panic("todo")

	fmt.Println("starting readrequest...")
	// Read start line
	req = &Request{}
	line, err := ReadLine(br)
	if err != nil {
		// fmt.Printf("1")
		return nil, false, err
	}

	// Read headers
	req.Method, req.URL, req.Proto, err = parseRequestLine(line)
	// fmt.Println("[ReadRequest]:", req.Method, req.URL, req.Proto, err)
	if err != nil {
		// fmt.Println("[ReadRequest] - There aren't 3 components in the first line.")
		return nil, true, err
	}

	// Check required headers
	if !validMethod(req.Method) {
		// fmt.Println(req.Method)
		// fmt.Println("3")
		return nil, true, fmt.Errorf("invalid method found %v", req.Method)
	}

	// Handle special headers
	for {
		line, err := ReadLine(br)
		// fmt.Println("readline in the for loop")
		if err != nil {
			fmt.Println("4")
			fmt.Println("[ReadRequest]:", err)
			return nil, true, err
		}

		// done reading from request
		if line == "" {
			// fmt.Println("end of reading")
			break
		}

		// if header not in a proper form (ex. missing a colon),
		// return nil, err -> signal a 400 error
		if !strings.Contains(line, ":") {
			fmt.Println("5")
			return nil, true, fmt.Errorf("miss colon")
		}

		// fmt.Println("Read line from request", line)
		fmt.Println("Read line from request", line)

		// parse each header
		// everything before the first colon is key
		fields := strings.SplitN(line, ":", 2)
		if len(fields) != 2 {
			// fmt.Println("6")
			return nil, true, fmt.Errorf("Could not parse the request header, for fields %v", fields)
		}

		// convert the key into canonical format
		new_k := CanonicalHeaderKey(fields[0])
		// fmt.Println("new_k:", new_k)

		// if value has spaces at the end
		// val := fields[1]
		// if new_k == "Connection" && val[(len(val))-1:] == " " {
		// 	return nil, true, fmt.Errorf("Connection's value")
		// }

		// get rid of leading empty spaces
		val := strings.TrimSpace(fields[1])

		// // check if value not containing CRLF
		// if strings.Contains(val, "\r\n") {
		// 	fmt.Println("Value not containing CRLF")
		// 	return nil, true, fmt.Errorf("Value not containing CRLF")
		// }

		// check if key is empty (key cannot be empty)
		if new_k == "" {
			// fmt.Println("Key cannot be empty.")
			return nil, true, fmt.Errorf("Key cannot be empty, for fields %v", fields)
		}

		// check if key contains a space
		if strings.Contains(new_k, " ") {
			// fmt.Println("Key contains a space")
			return nil, true, fmt.Errorf("Key could not have a space, for fields %v", fields)
		}

		// check if key is one or more alphanumeric or "-"
		re, _ := regexp.Compile("[0-9a-zA-Z-]+")
		match_key := re.FindString(new_k)
		if match_key != new_k {
			// fmt.Println("Key contains other than alphanumeric or '-'")
			return nil, true, fmt.Errorf("Key contains other than alphanumeric or '-', for fields %v", fields)
		}

		if req.Header == nil {
			req.Header = make(map[string]string, 0)
		}
		// store the key-val pair into map Header
		// excluding "Host" and "Connection"
		if new_k == "Host" {
			req.Host = val
		} else if new_k == "Connection" {
			// value is case-sensitive
			if val == "close" {
				// if fields[1] == "close" {
				// req.Close = true

				// can close the connection only when " close" after ':'
				if (fields[1]) == " close" {
					req.Close = true
				} else {
					req.Close = false
				}

			} else {
				// fmt.Println("Connection is [" + val + "]")
				req.Close = false
			}
		} else {
			req.Header[new_k] = val
			// fmt.Println("--> stroing key=[" + new_k + "], val=[" + val + "]")
		}
	}

	// check if key 'Host' is in request Header
	if req.Host == "" {
		// fmt.Println("Key 'Host' is not in Request Header")
		return nil, true, fmt.Errorf("Key 'Host' is not in Request Header")
	}

	return req, true, nil
}

func validMethod(method string) bool {
	return method == "GET"
}

func parseRequestLine(line string) (string, string, string, error) {
	fields := strings.SplitN(line, " ", -1)
	// check there are 3 components in the first line
	if len(fields) != 3 {
		return "", "", "", fmt.Errorf("Could not parse the request line, for fields %v", fields)
	}

	if (fields[1])[0] != '/' {
		return "", "", "", fmt.Errorf("does not start with /")
	}

	return fields[0], fields[1], fields[2], nil
}
