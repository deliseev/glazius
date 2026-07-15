package tracker

import (
	"context"
	"fmt"
	"strings"

	"net/http"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"
	// ...
)

type RutrackerClient struct {
	httpClient *http.Client
}

func NewRutrackerClient(client *http.Client) *RutrackerClient {
	return &RutrackerClient{httpClient: client}
}

func (c *RutrackerClient) getDocument(ctx context.Context, url string) (*goquery.Document, error) {
	// 1. Делаем запрос: http.NewRequestWithContext(ctx, "GET", url, nil)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	// defer req.Body.Close()

	// 2. Делаем c.httpClient.Do(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request error: %w", err)
	}
	defer resp.Body.Close()

	// 1. Используем charset.NewReader, чтобы конвертировать поток в UTF-8
	// Она сама прочитает заголовок или meta-тег и сделает нужную конвертацию
	utf8Reader, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		return nil, fmt.Errorf("charset parse error: %w", err)
	}

	// 2. Теперь отдаем "чистый" UTF-8 в goquery
	return goquery.NewDocumentFromReader(utf8Reader)
}

func (c *RutrackerClient) FetchInfo(ctx context.Context, url string) (string, string, error) {
	// 3. Загружаем тело ответа в goquery.NewDocumentFromReader(resp.Body)
	doc, err := c.getDocument(ctx, url)
	if err != nil {
		return "", "", fmt.Errorf("parse document: %w", err)
	}
	// 4. Ищем заголовок (h1.maintitle) и info_hash (обычно он в ссылке на скачивание или в тексте)
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
