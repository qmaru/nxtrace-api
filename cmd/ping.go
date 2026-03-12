package cmd

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type PingCommand struct {
	URL string `long:"url" description:"URL to ping" required:"true"`
}

func (c *PingCommand) Execute(args []string) error {
	url := c.URL
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}

	fmt.Printf("=> Connecting %s\n\n", url)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, _ := http.NewRequest("GET", url, nil)
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

	fmt.Printf("\nSuccess (elapsed: %v)\n", elapsed)
	return nil
}
