)

Luxior OSINT 2 — Production Intelligence Framework

Owner: Jet
GitHub: JettRnh
TikTok: @jettinibos_


---

What is this?

Luxior OSINT is a production-grade, polyglot intelligence framework built for real investigations.

This is not another Python script — it’s a full system where each language is used for what it does best:

Language	Component	Purpose

C++	Network Probe	Raw socket scanning, SYN probes, banner grabbing
Go	Web Crawler	High-concurrency crawling, link extraction, data harvesting
Rust	Data Parser	Memory-safe pattern extraction, parallel processing
Python	API + Workers	FastAPI backend, Redis queue, orchestration
Node.js	Darkweb Module	Tor integration, onion crawling, automation



---

Why I built this

Most OSINT tools are:

Slow (single-threaded scraping)

Limited (one data source)

Unreliable (no retry / weak error handling)

Not scalable


So I built something that actually works at scale.


---

Features

Network Intelligence

SYN scanning with banner grabbing

DNS enumeration (A, AAAA, MX, NS, TXT, SOA)

Service fingerprinting

SSL/TLS inspection


Web Intelligence

Concurrent crawling (20+ workers)

Email & phone extraction (deduplicated)

Link graph mapping

JS-aware crawling (extensible)


Data Extraction (Rust)

Emails, phones, IPs

Crypto wallets (BTC, ETH)

Social handles (GitHub, Twitter)

File hashes (MD5, SHA1, SHA256)


Darkweb Intelligence

Tor (SOCKS5) integration

Onion crawling with rate limiting

Credential leak detection

Darkweb search queries


Production Features

REST API (FastAPI)

Bearer authentication

Redis priority queue

PostgreSQL (JSONB storage)

Prometheus metrics

Docker + Kubernetes ready

Horizontal scaling



---

Architecture

┌───────────────────────────────┐ │        FastAPI API            │ │ Auth • Rate Limit • WebSocket │ └───────────────┬───────────────┘ │ ┌──────────────┼──────────────┐ ▼              ▼              ▼ PostgreSQL     Redis       Prometheus (Storage)     (Queue)      (Metrics) │ ┌──────────────┼──────────────┐ ▼              ▼              ▼ C++ Worker    Go Worker     Rust Worker (Network)     (Crawler)     (Parser) │ Node Worker (Darkweb)


---

Quick Start

Requirements

Linux (Ubuntu/Debian):

sudo apt update  
sudo apt install -y build-essential golang rustc cargo nodejs npm python3 python3-pip redis-server postgresql tor  
  
macOS:  
  
brew install go rust node python redis postgresql tor  
  
  
---  
  
Installation  
  
git clone https://github.com/JettRnh/luxior-osint.git  
cd luxior-osint  
  
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
  
All endpoints require a Bearer token.  
  
Default:  
  
luxior-secret-token  
  
  
---  
  
Start Scan  
  
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
  
CLI Tools  
  
C++ Network Probe  
  
./cmd/probe/lux_probe example.com 1 65535  
  
Example output:  
  
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
  
Rust Parser  
  
./cmd/parser/lux_parser ./data/  
  
  
---  
  
Performance  
  
Tested on 4-core / 8GB VPS:  
  
C++ probe → ~1000 ports in <10s  
  
Go crawler → ~500 pages/min  
  
Rust parser → ~10k files in ~2s  
  
System → 10+ concurrent scans  
  
  
  
---  
  
Repository Structure  
  
luxior-osint/  
├── cmd/  
├── api/  
├── pkg/  
├── internal/  
├── web/  
├── docker/  
├── scripts/  
├── configs/  
├── Makefile  
  
  
---  
  
Use Cases  
  
Domain investigation  
  
Large-scale monitoring  
  
Threat intelligence  
  
Darkweb tracking  
  
Data correlation pipelines  
  
  
  
---  
  
Credits  
  
Built by Jet.  
  
GitHub: https://github.com/JettRnh  
  
TikTok: @jettinibos_  
  
  
  
---  
  
License  
  
MIT License  
  
Use it, modify it, break it — responsibly.
