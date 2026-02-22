# TypeScript SDK

The TypeScript SDK provides WebSocket client integration with HotPlex server.

## Installation

```bash
npm install hotplex
# or
yarn add hotplex
# or
pnpm add hotplex
```

## Usage

```typescript
import { HotPlexClient, HotPlexConfig } from 'hotplex';

// Configure client
const config: HotPlexConfig = {
  baseUrl: 'ws://localhost:8080/ws/v1/agent',
  timeout: 300000
};

// Create client
const client = new HotPlexClient(config);

// Execute with callback
client.execute({
  sessionId: 'my-session',
  prompt: 'List the files in the current directory',
  workDir: '/tmp/ai-sandbox',
  callback: (eventType, data) => {
    switch (eventType) {
      case 'answer':
        console.log(`🤖 ${data}`);
        break;
      case 'tool_use':
        console.log(`🔧 Tool: ${data}`);
        break;
    }
  }
});
```

## Async Iterator

```typescript
// Stream events as async iterator
for await (const event of client.stream(sessionId, prompt, workDir)) {
  console.log(`${event.type}: ${event.data}`);
}
```

## Error Handling

```typescript
import { 
  HotPlexError, 
  ConnectionError, 
  TimeoutError,
  DangerBlockedError 
} from 'hotplex';

try {
  await client.execute({ sessionId, prompt, workDir, callback });
} catch (error) {
  if (error instanceof DangerBlockedError) {
    console.error('Dangerous operation blocked');
  } else if (error instanceof TimeoutError) {
    console.error('Operation timed out');
  } else if (error instanceof ConnectionError) {
    console.error('Connection failed');
  }
}
```

## Types

```typescript
interface HotPlexConfig {
  baseUrl: string;
  timeout?: number;
}

interface ExecuteOptions {
  sessionId: string;
  prompt: string;
  workDir: string;
  instructions?: string;
  callback?: (eventType: string, data: any) => void;
}
```

## Source

The TypeScript SDK source is available at [`sdks/typescript/`](https://github.com/hrygo/hotplex/tree/main/sdks/typescript).
