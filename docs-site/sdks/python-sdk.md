# Python SDK

The Python SDK provides WebSocket client integration with HotPlex server.

## Installation

```bash
pip install hotplex
```

## Usage

```python
from hotplex import HotPlexClient, HotPlexConfig

# Configure client
config = HotPlexConfig(
    base_url="ws://localhost:8080/ws/v1/agent",
    timeout=300.0
)

# Create client
client = HotPlexClient(config)

# Execute with callback
def on_event(event_type, data):
    if event_type == "answer":
        print(f"🤖 {data}")
    elif event_type == "tool_use":
        print(f"🔧 Tool: {data}")

# Execute prompt
result = client.execute(
    session_id="my-session",
    prompt="List the files in the current directory",
    work_dir="/tmp/ai-sandbox",
    callback=on_event
)
```

## Async Support

```python
import asyncio
from hotplex import AsyncHotPlexClient

async def main():
    client = AsyncHotPlexClient(config)
    
    async for event in client.stream(session_id, prompt, work_dir):
        print(f"{event.type}: {event.data}")

asyncio.run(main())
```

## Error Handling

```python
from hotplex import (
    HotPlexError,
    ConnectionError,
    TimeoutError,
    DangerBlockedError
)

try:
    client.execute(session_id, prompt, work_dir, callback)
except DangerBlockedError:
    print("Dangerous operation blocked")
except TimeoutError:
    print("Operation timed out")
except ConnectionError:
    print("Connection failed")
```

## Source

The Python SDK source is available at [`sdks/python/`](https://github.com/hrygo/hotplex/tree/main/sdks/python).
