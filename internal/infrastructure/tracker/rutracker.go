package tracker

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"net/http"
	"net/http/cookiejar"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"
	"golang.org/x/net/publicsuffix"
)

type RutrackerClient struct {
	httpClient *http.Client
}

func NewRutrackerClient() (*RutrackerClient, error) {
	jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})

	transport := &http.Transport{
		DisableKeepAlives: true,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			DualStack: false, // Отключаем попытки использовать IPv6
		}).DialContext,
	}

	client := &http.Client{
		Jar:       jar,
		Timeout:   30 * time.Second,
		Transport: transport,
	}
	return &RutrackerClient{httpClient: client}, nil
}

func (c *RutrackerClient) getDocument(ctx context.Context, url string) (*goquery.Document, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request error: %w", err)
	}
	defer resp.Body.Close()

	utf8Reader, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		return nil, fmt.Errorf("charset parse error: %w", err)
	}

	return goquery.NewDocumentFromReader(utf8Reader)
}

func (c *RutrackerClient) FetchInfo(ctx context.Context, url string) (string, string, error) {
	doc, err := c.getDocument(ctx, url)
	if err != nil {
		return "", "", fmt.Errorf("fail to parse document: %w", err)
	}

	title := strings.TrimSpace(doc.Find("h1.maintitle").Text())
	magnetLink, _ := doc.Find("a[href^='magnet:']").Attr("href")
	// Вырезаем кусок после "urn:btih:"
	// magnet:?xt=urn:btih:50AD87D0A55D32440B91FA66EE24B71AD8A3E190&tr=...
	parts := strings.Split(magnetLink, "urn:btih:")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("hash prefix not found in %v (%v)", magnetLink, parts)
	}
	parts = strings.Split(parts[1], "&")
	if len(parts) < 1 {
		return "", "", fmt.Errorf("hash postfix not found in %v", parts[1])
	}
	// 5. Возвращаем title и hash
	return title, parts[0], nil
}
