Readme

observability-demo
A minimal observability demo stack for a Go microservice. This repository boots a small mock Go service
that emits structured logs, exposes Prometheus metrics, and publishes messages to RabbitMQ. The stack
includes Prometheus, Loki + Promtail, Grafana, and RabbitMQ, orchestrated with Docker Compose.

This README assumes you’ve cloned the repo and are running Docker Desktop on Windows
(or Docker CE on Linux). All commands use the modern docker compose  (space) CLI.

## Overview

The  mock-service :  
- serves  HTTP endpoints  ( / ,  /health ,  /metrics )  
- logs  structured JSON to stdout  
- exposes  Prometheus  metrics  (request  count  +  latency  histogram)  
- publishes  messages  to RabbitMQ for each request  

The stack includes:  
- Prometheus → collects metrics  
- Loki + Promtail → collects logs  
- Grafana → dashboards  
- RabbitMQ → message broker  
- mock-service → demo Go service  

## Prerequisites

- Docker Desktop (Windows) or Docker CE (Linux)  
- Docker Compose v2 (bundled with Docker Desktop)  
- Git installed  

Verify Docker and Compose work:

```bash
docker --version
docker compose version
```

## How to run

Clone the repo:

```bash
git clone https://github.com/you/observability-demo.git
cd observability-demo
```

Start the stack:

```bash
docker compose up -d --build
```

Check running containers:

```bash
docker compose ps
```

View logs of a container:

```bash
docker compose logs -f mock-service
```

Stop everything:

```bash
docker compose down
```

Stop and remove volumes (clean state):

```bash
docker compose down -v
```

## Quick tests

Test service response:

```bash
curl http://localhost:8080/
```

Health endpoint:

```bash
curl http://localhost:8080/health
```

Metrics endpoint:

```bash
curl http://localhost:8080/metrics
```

Generate load (PowerShell example):

```powershell
for ($i=0; $i -lt 20; $i++) { curl http://localhost:8080/ > $null; Start-Sleep -Milliseconds 200 }
```

## Web UIs

- Grafana → [http://localhost:3000](http://localhost:3000) (user: `admin`, pass: `admin`)  
- Prometheus → [http://localhost:9090](http://localhost:9090)  
- RabbitMQ management → [http://localhost:15672](http://localhost:15672) (guest/guest)  
- Loki API → [http://localhost:3100](http://localhost:3100)  

## Validation checklist

Follow these steps to confirm everything works end-to-end:

1. **Check service is reachable**  
   ```bash
   curl http://localhost:8080/
   ```  
   You should see output like:  
   ```
   hello from mock-service at 2025-08-31T10:15:30Z
   ```

2. **Check health endpoint**  
   ```bash
   curl http://localhost:8080/health
   ```  
   Should return `ok`.

3. **Check Prometheus metrics**  
   ```bash
   curl http://localhost:8080/metrics | findstr http_requests_total
   ```  
   After some requests, you should see metrics like:  
   ```
   http_requests_total{method="GET",path="/",status="200"} 5
   ```

4. **Confirm Prometheus is scraping**  
   Visit [http://localhost:9090/targets](http://localhost:9090/targets).  
   The `mock-service` target should show **UP**.

5. **Check logs in Grafana (via Loki)**  
   - Open Grafana at [http://localhost:3000](http://localhost:3000)  
   - Login with `admin/admin`  
   - Add a Loki datasource with URL `http://loki:3100`  
   - Go to *Explore* → run query:  
     ```
     {service="mock-service"}
     ```  
   You should see structured logs like:  
   ```
   level=info msg="handled request" path=/ method=GET status=200
   ```

6. **Confirm RabbitMQ messages**  
   - Go to [http://localhost:15672](http://localhost:15672) (guest/guest)  
   - Check the `tasks` exchange → you should see messages published whenever `/` is requested.

## Next improvements

- Add Grafana dashboards for request rates & latency  
- Add Prometheus alerting rules  
- Integrate distributed tracing  

---
This README is meant as a quick start and validation guide. For deeper setup details, check the configuration files in the repo.
