package main

import (
	"bufio"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"log"
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

	reverseProxy := httputil.NewSingleHostReverseProxy(originServerURL)

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