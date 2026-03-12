package cmd

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

type PingCommand struct {
	URL string `long:"url" description:"URL to ping" required:"true"`
}

func printResolvConf() {
	fmt.Println("[diag] /etc/resolv.conf")

	data, err := os.ReadFile("/etc/resolv.conf")
	if err != nil {
		fmt.Printf("[diag] read error: %v\n\n", err)
		return
	}

	fmt.Println(string(data))
}

func printInterfaces() {
	fmt.Println("[diag] network interfaces")

	ifaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("[diag] interface error:", err)
		return
	}

	for _, iface := range ifaces {
		addrs, _ := iface.Addrs()
		fmt.Printf("[diag] %s %v\n", iface.Name, addrs)
	}

	fmt.Println()
}

func printRoute() {
	fmt.Println("[diag] route table")

	data, err := os.ReadFile("/proc/net/route")
	if err != nil {
		fmt.Println("[diag] route read error:", err)
		return
	}

	fmt.Println(string(data))
}

func checkDNS(host string) {
	fmt.Printf("[diag] DNS lookup %s\n", host)

	start := time.Now()
	ips, err := net.LookupIP(host)
	elapsed := time.Since(start)

	if err != nil {
		fmt.Printf("[diag] DNS failed: %v (elapsed %v)\n\n", err, elapsed)
		return
	}

	fmt.Printf("[diag] DNS result: %v (elapsed %v)\n\n", ips, elapsed)
}

func checkTCP(host, port string) {
	fmt.Printf("[diag] TCP connect %s:%s\n", host, port)

	start := time.Now()
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), 5*time.Second)
	elapsed := time.Since(start)

	if err != nil {
		fmt.Printf("[diag] TCP failed: %v (elapsed %v)\n\n", err, elapsed)
		return
	}

	conn.Close()
	fmt.Printf("[diag] TCP OK (elapsed %v)\n\n", elapsed)
}

func (c *PingCommand) Execute(args []string) error {
	url := c.URL
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	host := req.URL.Hostname()
	port := req.URL.Port()

	if port == "" {
		if req.URL.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}

	fmt.Println("========== Network Diagnostics ==========")

	printResolvConf()
	printInterfaces()
	printRoute()
	checkDNS(host)
	checkTCP(host, port)

	fmt.Println("=========================================")

	fmt.Printf("=> Connecting %s\n\n", url)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req.Header.Set("User-Agent", "curl")

	fmt.Printf("> GET %s HTTP/1.1\n", url)
	fmt.Printf("> Host: %s\n", req.Host)
	for k, v := range req.Header {
		fmt.Printf("> %s: %s\n", k, v[0])
	}
	fmt.Println(">")

	start := time.Now()
	resp, err := client.Do(req)
	elapsed := time.Since(start)

	if err != nil {
		fmt.Printf("\nConnection failed: %v (elapsed: %v)\n", err, elapsed)
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("< HTTP/1.1 %d %s\n", resp.StatusCode, http.StatusText(resp.StatusCode))
	for k, v := range resp.Header {
		fmt.Printf("< %s: %s\n", k, v[0])
	}
	fmt.Println("<")

	body, _ := io.ReadAll(resp.Body)
	if len(body) > 0 {
		fmt.Printf("< Response (%d bytes):\n%s\n", len(body), string(body))
	}

	fmt.Printf("=> Success (elapsed: %v)\n", elapsed)
	return nil
}
