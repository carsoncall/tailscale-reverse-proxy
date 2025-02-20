package main

import (
	"bufio"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/joho/godotenv"
	"tailscale.com/tsnet"
)

var (
	configPath = flag.String("config", "proxy.conf", "path to the config file")
)

func parseProxies(configPath string) map[string]string {
	file, err := os.Open(configPath)
	if err != nil {
		log.Fatalf("failed to open config file: %v", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	proxies := make(map[string]string)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			log.Fatalf("invalid line: lines must be <desired tailscale name>=<tailnet url>: %s", line)
		}
		proxies[parts[0]] = parts[1]
	}

	return proxies
}

func udsReverseProxy(url *url.URL) (udsProxy *httputil.ReverseProxy) {
	uds := url.Path
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			if (req.URL.Scheme == "") {
				req.URL.Scheme = "http" // Adjust if needed
			}
			req.URL.Host = "unix"    // Placeholder, not used for Unix sockets
			//req.URL.Path = "" // Path to your Unix socket
			req.Proto = "HTTP/1.1"
			req.ProtoMajor = 1
			req.ProtoMinor = 1
			req.Header.Set("X-Real-IP", req.RemoteAddr)
			req.Header.Set("X-Original-URI", strings.Split(req.RequestURI, ":")[0])
			req.Header.Set("X-Forwarded-Port", "80")
		},
		Transport: &http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				return net.Dial("unix", uds)
			},
		},
	}
	return proxy
}

func createProxy(hostname, origin string) error {
	defaultDirectory, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("can't find default user config directory: %v", err)
	}

	stateDir := filepath.Join(defaultDirectory, "ts-reverse-proxy", hostname)
	err = os.MkdirAll(stateDir, 0700)
		
	if err != nil {
		log.Fatalf("can't make proxy state directory: %v", err)
	}
		
	server := &tsnet.Server{
		Hostname: hostname,
		Dir: stateDir,
	}

	defer server.Close()

	listener, err := server.Listen("tcp", ":80")
	if err != nil {
		log.Fatal(err)
	}

	defer listener.Close()

	originServerURL, err := url.Parse(origin)
	if err != nil {
		log.Fatal("invalid origin server URL")
	}
	var reverseProxy *httputil.ReverseProxy
	if (originServerURL.Scheme == "unix") {
		reverseProxy = udsReverseProxy(originServerURL)
	} else {
		reverseProxy = httputil.NewSingleHostReverseProxy(originServerURL)
	}

	err = http.Serve(listener, reverseProxy)
	if err != nil {
		log.Printf("Error serving proxy for %s: %v", hostname, err)
	}
	return err
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	flag.Parse()

	proxies := parseProxies(*configPath)

	var wg sync.WaitGroup

	for hostname, origin := range proxies {
		wg.Add(1)
		go func(hostname, origin string) {
			defer wg.Done()
			createProxy(hostname, origin)
		}(hostname, origin)
	}

	wg.Wait()
}
