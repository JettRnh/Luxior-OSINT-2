#!/usr/bin/env python3
import asyncio, json, os, sys, uuid, logging
from datetime import datetime
from fastapi import FastAPI, HTTPException, Depends, status
from fastapi.middleware.cors import CORSMiddleware
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from pydantic import BaseModel, Field
import redis.asyncio as redis
import asyncpg

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class ScanRequest(BaseModel):
    target: str = Field(..., description="Target domain or IP")
    depth: int = Field(3, ge=1, le=5)
    max_urls: int = Field(1000, ge=10, le=10000)
    modules: list = Field(default=["dns", "port", "web"])
    priority: int = Field(5, ge=1, le=10)

class ScanResponse(BaseModel):
    scan_id: str
    target: str
    status: str
    created_at: datetime

class Database:
    def __init__(self): self.pool = None
    async def connect(self):
        self.pool = await asyncpg.create_pool(
            os.getenv("DATABASE_URL", "postgresql://lux_user:luxpass@localhost/lux_osint"),
            min_size=5, max_size=20
        )
        async with self.pool.acquire() as conn:
            await conn.execute('''
                CREATE TABLE IF NOT EXISTS scans (
                    id TEXT PRIMARY KEY, target TEXT NOT NULL, depth INTEGER,
                    max_urls INTEGER, modules JSONB, priority INTEGER,
                    status TEXT, result JSONB, created_at TIMESTAMP, updated_at TIMESTAMP
                )
            ''')
    async def create_scan(self, scan_id, target, depth, max_urls, modules, priority):
        async with self.pool.acquire() as conn:
            await conn.execute('''
                INSERT INTO scans (id, target, depth, max_urls, modules, priority, status, created_at)
                VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
            ''', scan_id, target, depth, max_urls, json.dumps(modules), priority, "pending", datetime.now())
    async def get_scan(self, scan_id):
        async with self.pool.acquire() as conn:
            row = await conn.fetchrow("SELECT * FROM scans WHERE id = $1", scan_id)
            return dict(row) if row else None

class RedisClient:
    def __init__(self): self.client = None
    async def connect(self):
        self.client = await redis.from_url(os.getenv("REDIS_URL", "redis://localhost:6379"), decode_responses=True)
    async def publish_job(self, scan_id, target, depth, max_urls, modules, priority):
        await self.client.lpush(f"jobs:priority:{priority}", json.dumps({
            "scan_id": scan_id, "target": target, "depth": depth,
            "max_urls": max_urls, "modules": modules, "priority": priority,
            "created_at": datetime.now().isoformat()
        }))

db = Database()
redis_client = RedisClient()

app = FastAPI(title="Luxior OSINT API", version="2.0.0")
app.add_middleware(CORSMiddleware, allow_origins=["*"], allow_credentials=True, allow_methods=["*"], allow_headers=["*"])
security = HTTPBearer()

async def verify_token(creds: HTTPAuthorizationCredentials = Depends(security)):
    if creds.credentials != os.getenv("API_TOKEN", "luxior-secret-token"):
        raise HTTPException(status_code=401, detail="Invalid token")
    return creds.credentials

@app.on_event("startup")
async def startup():
    await db.connect()
    await redis_client.connect()

@app.post("/api/v1/scan", response_model=ScanResponse, status_code=202)
async def start_scan(req: ScanRequest, token=Depends(verify_token)):
    scan_id = str(uuid.uuid4())
    await db.create_scan(scan_id, req.target, req.depth, req.max_urls, req.modules, req.priority)
    await redis_client.publish_job(scan_id, req.target, req.depth, req.max_urls, req.modules, req.priority)
    return ScanResponse(scan_id=scan_id, target=req.target, status="pending", created_at=datetime.now())

@app.get("/api/v1/scan/{scan_id}")
async def get_scan(scan_id: str, token=Depends(verify_token)):
    scan = await db.get_scan(scan_id)
    if not scan: raise HTTPException(status_code=404, detail="Not found")
    return scan

@app.get("/api/v1/health")
async def health(): return {"status": "healthy"}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8080)

