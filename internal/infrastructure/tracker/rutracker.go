package tracker

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"net/http"
	"net/http/cookiejar"
	"net/url"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"
	"golang.org/x/net/publicsuffix"
	// ...
)

type RutrackerClient struct {
	httpClient *http.Client
	loggedIn   bool
}

func NewRutrackerClient() (*RutrackerClient, error) {
	jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})

	// Настраиваем транспорт для большей стабильности
	transport := &http.Transport{
		DisableKeepAlives: true, // Исключает проблему с полузакрытыми соединениями
	}

	client := &http.Client{
		Jar:       jar,
		Timeout:   30 * time.Second,
		Transport: transport,
	}
	return &RutrackerClient{httpClient: client}, nil
}

func (c *RutrackerClient) getDocument(ctx context.Context, url string) (*goquery.Document, error) {
	// 1. Делаем запрос: http.NewRequestWithContext(ctx, "GET", url, nil)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}

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

func (c *RutrackerClient) Login(ctx context.Context, username, password string) error {
	// 1. GET для получения свежих кук и формы
	resp, err := c.httpClient.Get("https://rutracker.org/forum/login.php")
	if err != nil {
		return err
	}
	// Читаем тело, чтобы CookieJar подхватил куки
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	// 2. Формируем POST
	data := url.Values{}
	data.Set("login_username", username)
	data.Set("login_password", password)
	data.Set("login", "вход")
	data.Set("ssl", "1")

	req, err := http.NewRequestWithContext(ctx, "POST", "https://rutracker.org/forum/login.php", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	// Самые важные заголовки
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://rutracker.org/forum/login.php")
	req.Header.Set("Origin", "https://rutracker.org")

	resp, err = c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 3. Проверка
	u, _ := url.Parse("https://rutracker.org")
	cookies := c.httpClient.Jar.Cookies(u)

	for _, cookie := range cookies {
		if cookie.Name == "bb_session" {
			fmt.Println("Ура! Кука bb_session получена.")
			return nil
		}
	}

	return fmt.Errorf("логин не удался: кука bb_session не обнаружена")
}

func (c *RutrackerClient) FetchInfo(ctx context.Context, url string) (string, string, error) {
	// Загружаем тело ответа в goquery.NewDocumentFromReader(resp.Body)
	doc, err := c.getDocument(ctx, url)
	if err != nil {
		return "", "", fmt.Errorf("parse document: %w", err)
	}
	// Ищем заголовок (h1.maintitle) и info_hash (обычно он в ссылке на скачивание или в тексте)
	title := strings.TrimSpace(doc.Find("h1.maintitle").Text())

	// Ищем ссылку на скачивание (dl.php)
	// Rutracker обычно использует такие ссылки: <a href="dl.php?t=6880649" class="med bold">Скачать</a>
	downloadURL, exists := doc.Find("a[href^='dl.php?t=']").Attr("href")
	if !exists {
		html, _ := doc.Html()
		os.WriteFile("debug.html", []byte(html), 0644)
		fmt.Println("Debug HTML saved to debug.html")
		return "", "", fmt.Errorf("download link not found")
	}

	// Формируем полный URL для скачивания
	// Rutracker требует относительный путь от корня форума
	fullDownloadURL := "https://rutracker.org/forum/" + downloadURL
	log.Printf("download url: %v", fullDownloadURL)

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

func (c *RutrackerClient) DownloadTorrent(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/x-bittorrent") {
		return nil, fmt.Errorf("unexpected content type: %s", contentType)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
