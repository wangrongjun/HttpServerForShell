package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

var version = "<none>"
var versionFlag bool
var help bool
var bashPath string
var bashShell string
var port int

func init() {
	flag.BoolVar(&versionFlag, "v", false, "show version")
	flag.BoolVar(&help, "h", false, "show help doc")
	flag.StringVar(&bashPath, "b", "", "bash program path, such as: /bin/bash")
	flag.StringVar(&bashShell, "s", "./http-server.sh", "bash shell file to handle http request")
	flag.IntVar(&port, "p", 8080, "http server port")

	if bashPath == "" {
		osName := runtime.GOOS
		switch osName {
		case "windows":
			bashPath = "C:/Program Files/Git/bin/bash.exe"
		case "darwin":
			bashPath = "/bin/bash"
		case "linux":
			bashPath = "/bin/bash"
		default:
			log.Fatalln("unsupported os type: " + osName)
		}
	}
}

func main() {
	flag.Parse()
	if versionFlag {
		fmt.Println("Welcome to use Http Server For Shell, version: " + version)
		os.Exit(0)
	}
	if help {
		fmt.Println("Welcome to use Http Server For Shell, version: " + version)
		flag.PrintDefaults()
		os.Exit(0)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("[Start]", r.Proto, r.RemoteAddr, r.Method, r.URL.Path)
		statusCode, responseData := handleRequest(bashPath, bashShell, r)
		w.WriteHeader(statusCode)
		_, _ = w.Write(responseData)
		log.Println("[End]", statusCode)
	})

	portStr := strconv.Itoa(port)
	log.Println("Starting http server on localhost:" + portStr + " ...")
	err := http.ListenAndServe(":"+portStr, nil)
	if err != nil {
		panic(err)
	}
}

func handleRequest(bashPath string, bashShell string, r *http.Request) (statusCode int, responseData []byte) {
	var stdin bytes.Buffer
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// get request headers
	requestHeaders := ""
	var requestHeaderKeys []string
	for key := range r.Header {
		requestHeaderKeys = append(requestHeaderKeys, key)
	}
	sort.Strings(requestHeaderKeys)
	for _, key := range requestHeaderKeys {
		values := r.Header.Values(key)
		requestHeaders += key + "=" + strings.Join(values, ",") + "\n"
	}
	requestHeaders = strings.TrimSpace(requestHeaders)

	// get request params
	requestParams := ""
	var requestParamKeys []string
	for key := range r.URL.Query() {
		requestParamKeys = append(requestParamKeys, key)
	}
	sort.Strings(requestParamKeys)
	for _, key := range requestParamKeys {
		value := r.URL.Query().Get(key)
		requestParams += key + "=" + value + "\n"
	}
	requestParams = strings.TrimSpace(requestParams)

	// get request body
	var requestBody []byte
	if r.Method == http.MethodPost || r.Method == http.MethodPut {
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(r.Body)
		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			return http.StatusInternalServerError, []byte("read request body error: " + stdout.String() + "\n" + stderr.String())
		}
		requestBody = buf.Bytes()
	}

	// timeout config
	timeout := 30
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	// execute shell script to handle request
	cmd := exec.CommandContext(ctx, bashPath, bashShell)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Stdin = &stdin
	stdin.Write(requestBody)
	cmd.Env = []string{
		"REMOTE_ADDR=" + r.RemoteAddr,
		"REQUEST_METHOD=" + r.Method,
		"REQUEST_URI=" + r.RequestURI, // request path(contains params), such as: /users?id=1&name=wrj
		"REQUEST_HOST=" + r.Host,
		"REQUEST_PROTO=" + r.Proto,
		"REQUEST_PATH=" + r.URL.Path,        // request path(without params), such as: /users
		"REQUEST_HEADERS=" + requestHeaders, // request headers, such as: User-Agent=curl/7.70.0\nAccept=*/*
		"REQUEST_PARAMS=" + requestParams,   // request params, such as: id=1\nname=wrj
		"CONTENT_LENGTH=" + strconv.Itoa(int(r.ContentLength)),
	}
	_ = cmd.Run()
	if ctx.Err() == context.DeadlineExceeded { // if timeout
		return http.StatusInternalServerError, []byte("execute timeout, error msg: " + stdout.String() + "\n" + stderr.String())
	} else {
		exitCode := cmd.ProcessState.ExitCode()
		var statusCode int
		if exitCode == 0 {
			statusCode = http.StatusOK
		} else if exitCode == 400 {
			statusCode = http.StatusBadRequest
		} else if exitCode == 401 {
			statusCode = http.StatusUnauthorized
		} else if exitCode == 403 {
			statusCode = http.StatusForbidden
		} else if exitCode == 404 {
			statusCode = http.StatusNotFound
		} else if exitCode == 502 {
			statusCode = http.StatusBadGateway
		} else {
			statusCode = http.StatusInternalServerError
		}
		if exitCode == 0 {
			return statusCode, stdout.Bytes()
		} else {
			return statusCode, stderr.Bytes()
		}
	}

}
