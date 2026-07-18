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
	"github.com/deliseev/glazius/internal/domain/entity"
	"golang.org/x/net/html/charset"
	"golang.org/x/net/publicsuffix"
)

type RutrackerClient struct {
	httpClient *http.Client
	loggedIn   bool
	config     *entity.Config
}

func NewRutrackerClient(config *entity.Config) (*RutrackerClient, error) {
	jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})

	transport := &http.Transport{
		DisableKeepAlives: true,
	}

	client := &http.Client{
		Jar:       jar,
		Timeout:   30 * time.Second,
		Transport: transport,
	}
	return &RutrackerClient{httpClient: client, config: config}, nil
}

func (c *RutrackerClient) getDocument(ctx context.Context, url string) (*goquery.Document, error) {
	if !c.loggedIn {
		err := c.Login(ctx, c.config.Username, c.config.Password)
		if err != nil {
			return nil, err
		}
	}
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

func (c *RutrackerClient) Login(ctx context.Context, username, password string) error {
	// Формируем POST. rutracker использует кодировку Windows-1251,
	// поэтому кириллическое значение кнопки "вход" задаём сразу байтами
	// Windows-1251 (\xe2\xf5\xee\xe4) — url.Values.Encode() процентит их
	// как %E2%F5%EE%E4. Остальные поля ASCII, для них UTF-8 == Windows-1251.
	data := url.Values{}
	data.Set("login_username", username)
	data.Set("login_password", password)
	data.Set("login", "\xe2\xf5\xee\xe4")
	data.Set("ssl", "1")

	req, err := http.NewRequestWithContext(ctx, "POST", "https://rutracker.org/forum/login.php", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://rutracker.org/forum/login.php")
	req.Header.Set("Origin", "https://rutracker.org")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("POST login: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login failed: status %d, body: %s", resp.StatusCode, respBody)
	}
	io.Copy(io.Discard, resp.Body)

	// Проверка. bb_session ставится с path=/forum/ (см. BB.cookie_defaults
	// на странице rutracker), поэтому ищем куку по URL с путём /forum/,
	// иначе cookiejar её не вернёт для https://rutracker.org.
	u, _ := url.Parse("https://rutracker.org/forum/")
	for _, cookie := range c.httpClient.Jar.Cookies(u) {
		if cookie.Name == "bb_session" {
			c.loggedIn = true
			fmt.Printf("Кука bb_session получена: %v\n", cookie.Value)
			return nil
		}
	}
	return fmt.Errorf("логин не удался: кука bb_session не обнаружена")
}

func (c *RutrackerClient) FetchInfo(ctx context.Context, url string) (string, string, string, error) {
	doc, err := c.getDocument(ctx, url)
	if err != nil {
		return "", "", "", fmt.Errorf("fail to parse document: %w", err)
	}

	// Проверяем, не выкинули ли нас на страницу логина
	if doc.Find("form#login-form-quick").Length() > 0 {
		err = c.Login(ctx, c.config.Username, c.config.Password)
		if err != nil {
			return "", "", "", fmt.Errorf("session expired and relogin failed: %w", err)
		}
		doc, err = c.getDocument(ctx, url)
		if err != nil {
			return "", "", "", err
		}
	}

	title := strings.TrimSpace(doc.Find("h1.maintitle").Text())
	downloadURL, exists := doc.Find("a[href^='dl.php?t=']").Attr("href")
	if !exists {
		html, _ := doc.Html()
		os.WriteFile("debug.html", []byte(html), 0644)
		fmt.Println("Debug HTML saved to debug.html")
		return "", "", "", fmt.Errorf("download link not found")
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
		return "", "", "", fmt.Errorf("hash prefix not found in %v (%v)", magnetLink, parts)
	}
	parts = strings.Split(parts[1], "&")
	if len(parts) < 1 {
		return "", "", "", fmt.Errorf("hash postfix not found in %v", parts[1])
	}
	// 5. Возвращаем title и hash
	return title, parts[0], fullDownloadURL, nil
}

func (c *RutrackerClient) DownloadTorrent(ctx context.Context, url string) ([]byte, error) {
	log.Printf("GET torrent file: %v", url)
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
