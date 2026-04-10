package extractor

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Extracted struct {
	Title           string
	MetaDescription string
	Text            string
}

func FetchAndExtract(rawURL string, maxBytes int64, timeoutSec int) (*Extracted, error) {
	if _, err := ValidateURL(rawURL); err != nil {
		return nil, fmt.Errorf("URL validation failed: %w", err)
	}

	client := &http.Client{
		Timeout: time.Duration(timeoutSec) * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			if _, err := ValidateURL(req.URL.String()); err != nil {
				return fmt.Errorf("redirect to disallowed URL: %w", err)
			}
			return nil
		},
	}

	resp, err := client.Get(rawURL)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "text/html") && !strings.Contains(ct, "application/xhtml") {
		return nil, fmt.Errorf("unexpected Content-Type: %s", ct)
	}

	limited := io.LimitReader(resp.Body, maxBytes)
	doc, err := goquery.NewDocumentFromReader(limited)
	if err != nil {
		return nil, fmt.Errorf("HTML parse failed: %w", err)
	}

	return extractFromDoc(doc), nil
}

func extractFromDoc(doc *goquery.Document) *Extracted {
	e := &Extracted{}

	e.Title = strings.TrimSpace(doc.Find("title").First().Text())
	doc.Find(`meta[name="description"]`).Each(func(_ int, s *goquery.Selection) {
		if v, ok := s.Attr("content"); ok {
			e.MetaDescription = strings.TrimSpace(v)
		}
	})

	doc.Find("script, style, noscript, nav, footer, header, iframe").Remove()

	var parts []string

	doc.Find("h1, h2, h3").Each(func(_ int, s *goquery.Selection) {
		t := strings.TrimSpace(s.Text())
		if t != "" {
			parts = append(parts, t)
		}
	})

	candidates := []string{"main", "article", "[role=main]"}
	var body string
	for _, sel := range candidates {
		node := doc.Find(sel).First()
		if node.Length() > 0 {
			body = strings.TrimSpace(node.Text())
			break
		}
	}
	if body == "" {
		body = strings.TrimSpace(doc.Find("body").Text())
	}
	if body != "" {
		parts = append(parts, body)
	}

	raw := strings.Join(parts, "\n\n")
	e.Text = cleanText(raw)
	return e
}

var multiSpace = regexp.MustCompile(`[ \t]+`)
var multiNewline = regexp.MustCompile(`\n{3,}`)

func cleanText(s string) string {
	lines := strings.Split(s, "\n")
	var out []string
	for _, l := range lines {
		l = multiSpace.ReplaceAllString(l, " ")
		l = strings.TrimSpace(l)
		out = append(out, l)
	}
	result := strings.Join(out, "\n")
	result = multiNewline.ReplaceAllString(result, "\n\n")
	return strings.TrimSpace(result)
}

func TruncateText(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "\n...(truncated)"
}
