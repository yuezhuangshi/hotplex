# HotPlex Configuration Directory

This directory contains all configuration files for HotPlex deployments.

## Directory Structure

```
configs/
├── base/              # SSOT base configuration templates
│   ├── server.yaml    # Core server config
│   ├── slack.yaml     # Slack adapter config
│   ├── feishu.yaml    # Feishu adapter config
│   └── README.md      # Base config documentation
└── admin/             # Admin service configurations
```

## Quick Start

1. Copy base templates to your deployment directory:
   ```bash
   cp configs/base/slack.yaml configs/chatapps/slack.yaml
   cp configs/base/server.yaml configs/server.yaml
   ```

2. Set required environment variables (see `configs/base/README.md`)

3. Customize as needed

---

## Migrating from `configs/chatapps/`

> **Note**: The `configs/chatapps/` directory has been deprecated. All base configurations have moved to `configs/base/`.

### What Changed

| Old Path | New Path |
|----------|----------|
| `configs/chatapps/slack.yaml` | `configs/base/slack.yaml` |
| `configs/chatapps/feishu.yaml` | `configs/base/feishu.yaml` |
| `configs/chatapps/feishu.env.example` | Use `.env` file instead |
| `configs/server.yaml` | `configs/base/server.yaml` |

### Migration Steps

1. **Backup Old Config** (if still exists)
   ```bash
   cp -r configs/chatapps/ configs/chatapps.backup/
   ```

2. **Use New Base Templates**
   ```bash
   # Create your instance directory
   mkdir -p configs/instances/my-bot

   # Copy base configs
   cp configs/base/slack.yaml configs/instances/my-bot/
   cp configs/base/server.yaml configs/instances/my-bot/
   ```

3. **Set Environment Variables**

   Create a `.env` file with your credentials:
   ```bash
   # Slack credentials
   HOTPLEX_SLACK_BOT_USER_ID=UXXXXXXXXXX
   HOTPLEX_SLACK_BOT_TOKEN=xoxb-...
   HOTPLEX_SLACK_APP_TOKEN=xapp-...

   # Feishu credentials (if using)
   HOTPLEX_FEISHU_APP_ID=cli_...
   HOTPLEX_FEISHU_APP_SECRET=...
   ```

4. **Delete Old Config Directory**
   ```bash
   rm -rf configs/chatapps/
   ```

### Configuration Inheritance (Optional)

You can use the `inherits` field to extend base configs:

```yaml
# configs/instances/my-bot/slack.yaml
inherits: ../../base/slack.yaml

# Only override what you need
ai:
  system_prompt: |
    Your custom system prompt here...
```

### Key Differences

1. **No more `.env.example` in config dirs**: Use root `.env` file for all secrets
2. **SSOT principle**: Base templates in `configs/base/` are the source of truth
3. **Instance isolation**: Each bot instance should have its own config directory

## Further Documentation

- Base Config Details: `configs/base/README.md`
- Slack Manual: `docs/chatapps/chatapps-slack-manual.md`
- Feishu Manual: `docs/chatapps/chatapps-feishu-manual.md`
