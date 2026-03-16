---
name: HotPlex Diagnostics
description: Use this skill when the user asks to "diagnose hotplex", "check health", "view logs", "debug session", "check status", "get stats", "container logs", "check error". Provides monitoring and diagnostic capabilities for hotplex services.
version: 0.2.0
---

# HotPlex Diagnostics

Monitor and diagnose hotplex service health, logs, and session statistics.

## Critical: Working Directory

**All docker compose commands MUST be executed from the compose directory.**

```bash
COMPOSE_DIR="~/hotplex/docker/matrix"
```

**Pattern**: Always prefix docker compose commands with `cd $COMPOSE_DIR &&`:
```bash
cd ~/hotplex/docker/matrix && docker compose ps
```

## Container Reference

| Container | Port | Bot ID | Role |
|:----------|:-----|:-------|:-----|
| hotplex-01 | 18080 | U0AHRCL1KCM | Primary |
| hotplex-02 | 18081 | U0AJVRH4YF6 | Secondary |
| hotplex-03 | 18082 | U0AL7H8UU75 | Secondary |

## Quick Diagnostics

### Full System Health Check

```bash
cd ~/hotplex/docker/matrix && \
for bot in hotplex-01 hotplex-02 hotplex-03; do
  echo "=== $bot ==="
  docker compose ps $bot 2>/dev/null | tail -1
done
```

### View All Errors (All Containers)

```bash
cd ~/hotplex/docker/matrix && \
docker compose logs --tail=100 2>&1 | grep -i "error\|fatal\|panic"
```

### Check Socket Mode Status

```bash
cd ~/hotplex/docker/matrix && \
docker compose logs --tail=50 2>&1 | grep -E "Socket Mode|invalid_auth|Connected"
```

## Health Checks

### HTTP Health Endpoint

```bash
curl -s http://localhost:18080/health
curl -s http://localhost:18081/health
curl -s http://localhost:18082/health
```

### Container Health Status

```bash
docker inspect hotplex-01 --format='{{.State.Health.Status}}'
docker inspect hotplex-02 --format='{{.State.Health.Status}}'
docker inspect hotplex-03 --format='{{.State.Health.Status}}'
```

## Log Analysis

### View Recent Logs

```bash
# Single container
cd ~/hotplex/docker/matrix && \
docker compose logs --tail=200 hotplex-01

# All containers
cd ~/hotplex/docker/matrix && \
docker compose logs --tail=100
```

### Search Logs for Errors

```bash
# Single container errors
docker logs hotplex-01 --since "1h" 2>&1 | grep -i error

# All containers errors
cd ~/hotplex/docker/matrix && \
for bot in hotplex-01 hotplex-02 hotplex-03; do
  echo "=== $bot ==="
  docker logs $bot --since "1h" 2>&1 | grep -i error | head -10
done
```

### Session Tracing

```bash
# Replace SESSION_ID with actual session ID
docker logs hotplex-01 2>&1 | grep "SESSION_ID"
```

### Follow Logs in Real-time

```bash
cd ~/hotplex/docker/matrix && \
docker compose logs -f hotplex-01
```

## Performance Monitoring

### Container Resources

```bash
docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}" \
  hotplex-01 hotplex-02 hotplex-03
```

### Disk Usage

```bash
docker exec hotplex-01 du -sh /home/hotplex/.hotplex
docker exec hotplex-01 du -sh /home/hotplex/.claude
```

## Debugging Sessions

### Enter Container Shell

```bash
docker exec -it hotplex-01 /bin/sh
```

### Check Running Processes

```bash
docker exec hotplex-01 ps aux | grep -E "(claude|opencode)"
```

### Network Diagnostics

```bash
docker exec hotplex-01 wget -qO- http://localhost:8080/health
```

## Container Discovery

If user doesn't specify which bot:

```bash
cd ~/hotplex/docker/matrix && docker compose ps
```

Then use the container name (hotplex-01, hotplex-02, or hotplex-03) in subsequent commands.

## Additional Resources

### Reference Files

- `internal/server/hotplex_ws.go` - WebSocket API implementation

### Related Skills

- `docker-container-ops` - For container lifecycle management
- `hotplex-data-mgmt` - For data and session management
