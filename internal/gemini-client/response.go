package gemini_client

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const (
	URLMaxLength  = 1024
	MetaMaxLength = 1024
)

const (
	StatusInput          = 10
	StatusSensitiveInput = 11

	StatusSuccess = 20

	StatusRedirect          = 30
	StatusRedirectTemporary = 30
	StatusRedirectPermanent = 31

	StatusTemporaryFailure = 40
	StatusUnavailable      = 41
	StatusCGIError         = 42
	StatusProxyError       = 43
	StatusSlowDown         = 44

	StatusPermanentFailure    = 50
	StatusNotFound            = 51
	StatusGone                = 52
	StatusProxyRequestRefused = 53
	StatusBadRequest          = 59

	StatusClientCertRequired = 60
	StatusCertNotAuthorized  = 61
	StatusCertNotValid       = 62
)

var statusText = map[int]string{
	StatusInput:          "Input",
	StatusSensitiveInput: "Sensitive Input",

	StatusSuccess: "Success",

	StatusRedirectTemporary: "Redirect - Temporary",
	StatusRedirectPermanent: "Redirect - Permanent",

	StatusTemporaryFailure: "Temporary Failure",
	StatusUnavailable:      "Server Unavailable",
	StatusCGIError:         "CGI Error",
	StatusProxyError:       "Proxy Error",
	StatusSlowDown:         "Slow Down",

	StatusPermanentFailure:    "Permanent Failure",
	StatusNotFound:            "Not Found",
	StatusGone:                "Gone",
	StatusProxyRequestRefused: "Proxy Request Refused",
	StatusBadRequest:          "Bad Request",

	StatusClientCertRequired: "Client Certificate Required",
	StatusCertNotAuthorized:  "Certificate Not Authorized",
	StatusCertNotValid:       "Certificate Not Valid",
}

func StatusText(code int) string {
	return statusText[code]
}

type GeminiResponse struct {
	Status int
	Meta   string
	Body   string
}

type header struct {
	status int
	meta   string
}

func getResponse(res *GeminiResponse, reader *bufio.Reader) error {
	header, err := getHeader(reader)
	if err != nil {
		return err
	}

	var body string
	if header.status == StatusSuccess {
		bodyBytes, err := io.ReadAll(reader)
		if err != nil {
			return fmt.Errorf("Failed to read body: %w\n", err)
		}
		body = string(bodyBytes)
	}

	res.Body = body
	res.Status = header.status
	res.Meta = header.meta

	return nil
}

func getHeader(reader *bufio.Reader) (header, error) {
	headerLine, err := reader.ReadString('\n')
	if err != nil {
		return header{}, fmt.Errorf("Failed to read header: %w\n", err)
	}

	parts := strings.SplitN(headerLine, " ", 2)
	if len(parts) < 2 {
		return header{}, fmt.Errorf("Invalid header line: %s\n", headerLine)
	}

	statusCode, err := strconv.Atoi(parts[0])
	if err != nil {
		return header{}, fmt.Errorf("Invalid status code: %s\n", parts[0])
	}

	return header{
		status: statusCode,
		meta:   parts[1],
	}, nil
}
