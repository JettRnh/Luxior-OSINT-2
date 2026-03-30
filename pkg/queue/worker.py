#!/usr/bin/env python3
import asyncio, json, os, sys, signal, logging, subprocess
from datetime import datetime
import redis.asyncio as redis
import asyncpg

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class LuxWorker:
    def __init__(self, worker_id: int):
        self.worker_id = worker_id
        self.running = True
        self.redis = None
        self.pg = None
    
    async def initialize(self):
        self.redis = await redis.from_url(os.getenv("REDIS_URL", "redis://localhost:6379"), decode_responses=True)
        self.pg = await asyncpg.create_pool(os.getenv("DATABASE_URL", "postgresql://lux_user:luxpass@localhost/lux_osint"), min_size=1, max_size=5)
        logger.info(f"Worker {self.worker_id} initialized")
    
    async def fetch_job(self):
        for priority in range(1, 11):
            job = await self.redis.rpop(f"jobs:priority:{priority}")
            if job: return json.loads(job)
        return None
    
    async def update_status(self, scan_id, status, result=None):
        async with self.pg.acquire() as conn:
            if result:
                await conn.execute("UPDATE scans SET status=$1, result=$2, updated_at=$3 WHERE id=$4",
                                   status, json.dumps(result), datetime.now(), scan_id)
            else:
                await conn.execute("UPDATE scans SET status=$1, updated_at=$2 WHERE id=$3",
                                   status, datetime.now(), scan_id)
    
    async def run_probe(self, target):
        result = {"dns": [], "ports": {}}
        try:
            proc = await asyncio.create_subprocess_exec("./cmd/probe/lux_probe", target, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
            stdout, _ = await proc.communicate()
            for line in stdout.decode().strip().split('\n'):
                if line.startswith("PORT_OPEN"):
                    try:
                        header, banner = line.split('|', 1)
                        port = int(header.split()[1])
                        result["ports"][port] = banner
                    except: continue
                elif line and not line.startswith(("SCAN", "DNS_RESULTS")):
                    result["dns"].append(line)
        except Exception as e: logger.error(f"Probe failed: {e}")
        return result
    
    async def run_crawler(self, target, depth, max_urls):
        result = {"urls": []}
        try:
            proc = await asyncio.create_subprocess_exec(
                "./cmd/crawler/lux_crawler", "-target", target, "-depth", str(depth), "-max", str(max_urls),
                stdout=subprocess.PIPE, stderr=subprocess.PIPE)
            stdout, _ = await proc.communicate()
            for line in stdout.decode().strip().split('\n'):
                if line.startswith("CRAWLED"):
                    parts = line.split('|')
                    if len(parts) >= 2: result["urls"].append(parts[1])
        except Exception as e: logger.error(f"Crawler failed: {e}")
        return result
    
    async def process_job(self, job):
        scan_id, target, depth, max_urls, modules = job["scan_id"], job["target"], job["depth"], job["max_urls"], job["modules"]
        logger.info(f"Worker {self.worker_id} processing {scan_id} on {target}")
        await self.update_status(scan_id, "running")
        results = {}
        if "dns" in modules or "port" in modules:
            results["network"] = await self.run_probe(target)
        if "web" in modules:
            results["web"] = await self.run_crawler(target, depth, max_urls)
        await self.update_status(scan_id, "completed", results)
        logger.info(f"Worker {self.worker_id} completed {scan_id}")
    
    async def run(self):
        await self.initialize()
        logger.info(f"Worker {self.worker_id} started")
        while self.running:
            job = await self.fetch_job()
            if job: await self.process_job(job)
            else: await asyncio.sleep(1)
        await self.redis.close()
        await self.pg.close()
    
    def stop(self): self.running = False

async def main():
    worker = LuxWorker(int(os.getenv("WORKER_ID", "1")))
    loop = asyncio.get_event_loop()
    for sig in (signal.SIGINT, signal.SIGTERM):
        loop.add_signal_handler(sig, worker.stop)
    await worker.run()

if __name__ == "__main__": asyncio.run(main())
