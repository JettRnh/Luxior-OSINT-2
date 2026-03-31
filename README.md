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
| **C++** | Network Probe | Raw socket scanning, SYN probes, banner grabbing |
| **Go** | Web Crawler | Concurrent crawling, link extraction, email harvesting |
| **Rust** | Data Parser | Memory-safe pattern extraction, parallel processing |
| **Python** | API + Worker | FastAPI backend, Redis queue, PostgreSQL storage |
| **Node.js** | Darkweb Module | Tor integration, onion crawling, leak detection |

---

## Why I built this

I was tired of OSINT tools that are:
- Slow (single-threaded Python scraping)
- Limited (only one data source)
- Unreliable (no proper error handling)
- Not scalable

So I built something that actually works for real investigations.

---

## Features

### Network Intelligence
- SYN scan with banner grabbing
- DNS enumeration (A, AAAA, MX, NS, TXT, SOA)
- Service detection
- SSL certificate analysis

### Web Intelligence
- Concurrent crawling (20 workers default)
- Email and phone extraction
- Link mapping
- JavaScript rendering support

### Data Extraction (Rust)
- Email addresses
- Phone numbers
- IP addresses
- Crypto wallets (BTC, ETH)
- Social handles
- File hashes

### Darkweb Intelligence
- Tor integration (socks5)
- Onion crawling
- Credential leak detection
- Darkweb queries

### Production Features
- REST API with Bearer token
- Redis queue
- PostgreSQL (JSONB)
- Prometheus metrics
- Docker & Kubernetes
- Horizontal scaling

---

## Quick Start

### Prerequisites

Ubuntu/Debian:

    sudo apt update
    sudo apt install -y build-essential golang rustc cargo nodejs npm python3 python3-pip redis-server postgresql tor

macOS:

    brew install go rust node python redis postgresql tor

---

### Installation

    git clone https://github.com/JettRnh/Luxior-OSINT-2.git
    cd Luxior-OSINT-2
    chmod +x scripts/setup.sh
    ./scripts/setup.sh
    make build
    make setup-db
    make run

---

### Docker (Production)

    make docker-build
    make docker-up

---

## API Usage

### Authentication

All endpoints require Bearer token.  
Default: luxior-secret-token (change in .env)

---

### Start a Scan

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

### Check Status

    curl http://localhost:8080/api/v1/scan/{scan_id} \
      -H "Authorization: Bearer luxior-secret-token"

---

### Health Check

    curl http://localhost:8080/api/v1/health

---

## Command Line Tools

### C++ Network Probe

    ./cmd/probe/lux_probe example.com 1 65535

Output:

    DNS_RESULTS
    93.184.216.34
    SCAN_START 93.184.216.34 1 65535
    PORT_OPEN 80|HTTP/1.0 200 OK
    PORT_OPEN 443|
    SCAN_COMPLETE 2

---

### Go Web Crawler

    ./cmd/crawler/lux_crawler -target https://example.com -depth 3 -max 1000

---

### Rust Data Parser

    ./cmd/parser/lux_parser ./data/

---

## Performance

On a 4-core, 8GB VPS:

- C++ probe: 1000 ports < 10 seconds  
- Go crawler: 500 pages/min  
- Rust parser: 10,000 files < 2 seconds  
- System: 10+ scans simultaneously  

---

## Repository Structure

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

## Real Use Cases

1. Investigate domain

    curl -X POST http://localhost:8080/api/v1/scan \
      -d '{"target": "suspicious.com"}'

---

2. Monitor many domains

    submit to Redis queue  
    workers process parallel  
    results stored in PostgreSQL  

---

3. Darkweb intelligence

    Tor crawler scheduled  
    keyword monitoring  
    leak alerts  

---

## Credits

Built by Jet  

GitHub: https://github.com/JettRnh  
TikTok: @jettinibos_  

---

## License

MIT — Use it, break it, fix it. Just don't blame me.
