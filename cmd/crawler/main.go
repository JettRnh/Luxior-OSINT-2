package main

import (
    "context"
    "crypto/tls"
    "database/sql"
    "encoding/json"
    "flag"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "os"
    "os/signal"
    "regexp"
    "strings"
    "sync"
    "sync/atomic"
    "syscall"
    "time"
    
    _ "github.com/mattn/go-sqlite3"
    "github.com/redis/go-redis/v9"
    "golang.org/x/net/html"
)

type CrawlResult struct {
    URL         string    `json:"url"`
    Title       string    `json:"title"`
    StatusCode  int       `json:"status_code"`
    ContentType string    `json:"content_type"`
    Links       []string  `json:"links"`
    Emails      []string  `json:"emails"`
    Phones      []string  `json:"phones"`
    IPs         []string  `json:"ips"`
    CrawledAt   time.Time `json:"crawled_at"`
}

type Crawler struct {
    client       *http.Client
    rdb          *redis.Client
    db           *sql.DB
    targetDomain string
    maxDepth     int
    maxURLs      int
    visited      sync.Map
    urlQueue     chan string
    results      chan CrawlResult
    wg           sync.WaitGroup
    ctx          context.Context
    cancel       context.CancelFunc
    activeCount  int32
    stopped      atomic.Bool
    emailRegex   *regexp.Regexp
    phoneRegex   *regexp.Regexp
    ipRegex      *regexp.Regexp
}

func NewCrawler(domain string, depth int, max int, redisURL string, insecureTLS bool) (*Crawler, error) {
    transport := &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureTLS},
        MaxIdleConns:    100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout: 90 * time.Second,
    }
    client := &http.Client{Transport: transport, Timeout: 30 * time.Second}
    opt, _ := redis.ParseURL(redisURL)
    rdb := redis.NewClient(opt)
    db, _ := sql.Open("sqlite3", "lux_crawl.db")
    db.Exec(`CREATE TABLE IF NOT EXISTS crawled_pages (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        url TEXT UNIQUE, title TEXT, status_code INTEGER, content_type TEXT,
        emails TEXT, phones TEXT, ips TEXT, crawled_at DATETIME)`)
    ctx, cancel := context.WithCancel(context.Background())
    return &Crawler{
        client: client, rdb: rdb, db: db, targetDomain: domain,
        maxDepth: depth, maxURLs: max, urlQueue: make(chan string, 10000),
        results: make(chan CrawlResult, 1000), ctx: ctx, cancel: cancel,
        emailRegex: regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
        phoneRegex: regexp.MustCompile(`(\+?[0-9]{1,3}[-.\s]?)?\(?[0-9]{3}\)?[-.\s]?[0-9]{3}[-.\s]?[0-9]{4}`),
        ipRegex:    regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`),
    }, nil
}

func (c *Crawler) extractData(body io.Reader, baseURL string) (title string, links, emails, phones, ips []string) {
    doc, _ := html.Parse(body)
    var f func(*html.Node)
    f = func(n *html.Node) {
        if n.Type == html.ElementNode {
            if n.Data == "title" && n.FirstChild != nil { title = n.FirstChild.Data }
            if n.Data == "a" {
                for _, attr := range n.Attr {
                    if attr.Key == "href" {
                        link := c.resolveURL(attr.Val, baseURL)
                        if strings.Contains(link, c.targetDomain) { links = append(links, link) }
                    }
                }
            }
        }
        if n.Type == html.TextNode {
            emails = append(emails, c.emailRegex.FindAllString(n.Data, -1)...)
            phones = append(phones, c.phoneRegex.FindAllString(n.Data, -1)...)
            ips = append(ips, c.ipRegex.FindAllString(n.Data, -1)...)
        }
        for child := n.FirstChild; child != nil; child = child.NextSibling { f(child) }
    }
    f(doc)
    emails = uniqueStrings(emails)
    phones = uniqueStrings(phones)
    ips = uniqueStrings(ips)
    links = uniqueStrings(links)
    return
}

func (c *Crawler) resolveURL(href, base string) string {
    parsedBase, _ := url.Parse(base)
    parsedHref, _ := url.Parse(href)
    resolved := parsedBase.ResolveReference(parsedHref)
    return resolved.String()
}

func (c *Crawler) fetchAndParse(rawURL string) CrawlResult {
    result := CrawlResult{URL: rawURL, CrawledAt: time.Now()}
    req, _ := http.NewRequestWithContext(c.ctx, "GET", rawURL, nil)
    req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Luxior OSINT Agent/2.0")
    resp, err := c.client.Do(req)
    if err != nil { return result }
    defer resp.Body.Close()
    result.StatusCode = resp.StatusCode
    result.ContentType = resp.Header.Get("Content-Type")
    if resp.StatusCode != 200 { return result }
    if strings.Contains(result.ContentType, "text/html") {
        title, links, emails, phones, ips := c.extractData(resp.Body, rawURL)
        result.Title = title
        result.Links = links
        result.Emails = emails
        result.Phones = phones
        result.IPs = ips
    }
    return result
}

func uniqueStrings(slice []string) []string {
    keys := make(map[string]bool)
    var list []string
    for _, entry := range slice {
        if entry != "" && !keys[entry] {
            keys[entry] = true
            list = append(list, entry)
        }
    }
    return list
}

func (c *Crawler) worker() {
    defer c.wg.Done()
    for {
        select {
        case <-c.ctx.Done(): return
        case url, ok := <-c.urlQueue:
            if !ok { return }
            if _, loaded := c.visited.LoadOrStore(url, true); loaded { continue }
            atomic.AddInt32(&c.activeCount, 1)
            result := c.fetchAndParse(url)
            c.results <- result
            jsonData, _ := json.Marshal(result)
            c.rdb.LPush(c.ctx, "lux:crawl:queue", jsonData)
            c.saveToDB(result)
            for _, link := range result.Links {
                select { case c.urlQueue <- link: default: }
            }
            atomic.AddInt32(&c.activeCount, -1)
        }
    }
}

func (c *Crawler) saveToDB(result CrawlResult) {
    stmt, _ := c.db.Prepare(`INSERT OR REPLACE INTO crawled_pages 
        (url, title, status_code, content_type, emails, phones, ips, crawled_at) 
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
    defer stmt.Close()
    stmt.Exec(result.URL, result.Title, result.StatusCode, result.ContentType,
        strings.Join(result.Emails, ","), strings.Join(result.Phones, ","),
        strings.Join(result.IPs, ","), result.CrawledAt)
}

func (c *Crawler) Start(startURL string) {
    c.urlQueue <- startURL
    for i := 0; i < 20; i++ { c.wg.Add(1); go c.worker() }
    go func() { c.wg.Wait(); close(c.results) }()
    for result := range c.results {
        fmt.Printf("CRAWLED %s|%d|%s\n", result.URL, result.StatusCode, result.Title)
    }
}

func (c *Crawler) Stop() { c.cancel(); close(c.urlQueue); c.client.CloseIdleConnections() }

func main() {
    var target, redisURL string
    var depth, maxURLs int
    var insecureTLS bool
    flag.StringVar(&target, "target", "", "Target URL")
    flag.IntVar(&depth, "depth", 3, "Crawl depth")
    flag.IntVar(&maxURLs, "max", 1000, "Max URLs")
    flag.StringVar(&redisURL, "redis", "redis://localhost:6379", "Redis URL")
    flag.BoolVar(&insecureTLS, "insecure", false, "Skip TLS verification")
    flag.Parse()
    if target == "" { fmt.Fprintf(os.Stderr, "Usage: lux_crawler -target <url>\n"); os.Exit(1) }
    crawler, _ := NewCrawler(target, depth, maxURLs, redisURL, insecureTLS)
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    go func() { <-sigChan; crawler.Stop() }()
    fmt.Printf("Starting crawl on %s (depth: %d, max: %d)\n", target, depth, maxURLs)
    crawler.Start(target)
    fmt.Println("Crawl completed")
}
