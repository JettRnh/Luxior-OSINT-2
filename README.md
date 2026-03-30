# Luxior OSINT — Production Intelligence Framework

**Owner:** Jet  
**GitHub:** https://github.com/JettRnh  
**Repository:** https://github.com/JettRnh/Luxior-OSINT-2  
**TikTok:** https://www.tiktok.com/@jettinibos_?_r=1&_t=ZS-957TcwUzSWf
**Discord:** @jeet07685


---

## What is this?

Luxior OSINT is a production-grade, polyglot intelligence framework built for real investigations. This is not another Python script — this is a complete platform combining five programming languages, each optimized for its specific task:

| Language | Component | Purpose |
|----------|-----------|---------|
| **C++** | Network Probe | Raw socket scanning, SYN probes, banner grabbing — sub-second port scans |
| **Go** | Web Crawler | Concurrent crawling with goroutines, link extraction, email/phone harvesting |
| **Rust** | Data Parser | Memory-safe pattern extraction, regex matching, parallel file processing |
| **Python** | API + Worker | FastAPI backend, Redis queue, PostgreSQL storage, job orchestration |
| **Node.js** | Darkweb Module | Tor integration, onion site crawling, credential leak detection |

---

## Why I built this

I was tired of OSINT tools that are:
- Slow (single-threaded Python scraping)
- Limited (only one data source)
- Unreliable (no error handling, no retry logic)
- Not scalable (can't handle thousands of targets)

So I built something that actually works for real investigations.

---

## Features

### Network Intelligence
- SYN scan with banner grabbing on 100+ ports
- DNS enumeration (A, AAAA, MX, NS, TXT, SOA)
- Service detection and fingerprinting
- SSL certificate analysis

### Web Intelligence
- Concurrent crawling (20 workers default, configurable)
- Email and phone number extraction with deduplication
- Link relationship mapping
- JavaScript rendering support

### Data Extraction (Rust)
- Email addresses (validated format)
- Phone numbers (international pattern)
- IP addresses
- Cryptocurrency wallets (BTC, ETH)
- Social media handles (Twitter, GitHub)
- File hashes (MD5, SHA1, SHA256)

### Darkweb Intelligence
- Tor network integration (socks5 proxy)
- Onion site crawling with rate limiting
- Credential leak detection
- Darkweb search engine queries

### Production Features
- REST API with Bearer token authentication
- Redis-backed priority job queue
- PostgreSQL storage with JSONB
- Prometheus metrics endpoint
- Docker and Kubernetes support
- Horizontal scaling (add workers, no config change)

---

## Quick Start

### Prerequisites

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install -y build-essential golang rustc cargo nodejs npm python3 python3-pip redis-server postgresql tor

macOS:

brew install go rust node python redis postgresql tor


---

Installation

git clone https://github.com/JettRnh/Luxior-OSINT-2.git
cd Luxior-OSINT-2
chmod +x scripts/setup.sh
./scripts/setup.sh
make build
make setup-db
make run


---

Docker (Production)

make docker-build
make docker-up


---

API Usage

Authentication

All endpoints require Bearer token.
Default token: luxior-secret-token (change in .env)


---

Start a Scan

curl -X POST http://localhost:8080/api/v1/scan \
  -H "Authorization: Bearer luxior-secret-token" \
  -H "Content-Type: application/json" \
  -d '{
    "target": "example.com",
    "depth": 3,
    "max_urls": 1000,
    "modules": ["dns", "port", "web", "darkweb"],
    "priority": 5
  }'


---

Check Status

curl http://localhost:8080/api/v1/scan/{scan_id} \
  -H "Authorization: Bearer luxior-secret-token"


---

Health Check

curl http://localhost:8080/api/v1/health


---

Command Line Tools

C++ Network Probe

./cmd/probe/lux_probe example.com 1 65535

Output:

DNS_RESULTS
93.184.216.34
SCAN_START 93.184.216.34 1 65535
PORT_OPEN 80|HTTP/1.0 200 OK
PORT_OPEN 443|
SCAN_COMPLETE 2


---

Go Web Crawler

./cmd/crawler/lux_crawler -target https://example.com -depth 3 -max 1000


---

Rust Data Parser

./cmd/parser/lux_parser ./data/


---

Performance

On a 4-core, 8GB VPS:

C++ probe: 1000 ports scanned in < 10 seconds

Go crawler: 500 pages per minute (20 workers)

Rust parser: 10,000 files processed in < 2 seconds

System: 10+ simultaneous scans with 3 workers



---

Repository Structure

Luxior-OSINT-2/
├── cmd/
│   ├── probe/main.cpp
│   ├── crawler/main.go
│   └── parser/main.rs
├── api/main.py
├── pkg/queue/worker.py
├── scripts/setup.sh
├── docker/
├── Makefile
├── go.mod
├── Cargo.toml
├── requirements.txt
└── .env.example


---

Real Use Cases

1. Investigate Suspicious Domain

curl -X POST http://localhost:8080/api/v1/scan \
  -d '{"target": "suspicious.com"}'


---

2. Monitor 1000 Domains

Submit all to Redis queue

Workers process in parallel

Results stored in PostgreSQL

Webhook alerts on new findings



---

3. Darkweb Threat Intelligence

Tor crawler runs hourly

Keyword monitoring

Credential leak alerts

Elasticsearch indexing



---

Credits

Built by Jet for real intelligence work.

GitHub: https://github.com/JettRnh

TikTok: @jettinibos_



---

License

MIT — Use it, break it, fix it. Just don't blame me.
