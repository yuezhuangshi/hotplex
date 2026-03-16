# HotPlex Base Configurations (SSOT)

This directory contains the **Single Source of Truth (SSOT)** base configuration templates for HotPlex deployments.

## Overview

These templates are designed to be copied and customized for your specific deployment environment. They use environment variable substitution (`${VAR}`) for sensitive values and deployment-specific settings.

## Files

| File | Description |
|------|-------------|
| `server.yaml` | Core server configuration (engine, security, logging) |
| `slack.yaml` | Slack adapter configuration (Socket Mode, permissions, AI identity) |
| `slack_capabilities.yaml` | App Home Capability Center definitions for Slack |
| `feishu.yaml` | Feishu/Lark adapter configuration (HTTP Webhook mode) |

## Usage

### 1. Copy Templates to Your Config Directory

```bash
# For Docker deployment
mkdir -p configs/chatapps
cp configs/base/slack.yaml configs/chatapps/slack.yaml
cp configs/base/feishu.yaml configs/chatapps/feishu.yaml

# For server config
cp configs/base/server.yaml configs/server.yaml
```

### 2. Set Required Environment Variables

Create a `.env` file or export environment variables:

```bash
# Core settings
export HOTPLEX_API_KEY="your-api-key"
export HOTPLEX_PROJECTS_DIR="/path/to/projects"

# Slack credentials
export HOTPLEX_SLACK_BOT_USER_ID="UXXXXXXXXXX"
export HOTPLEX_SLACK_BOT_TOKEN="xoxb-..."
export HOTPLEX_SLACK_APP_TOKEN="xapp-..."
export HOTPLEX_SLACK_PRIMARY_OWNER="UXXXXXXXXXX"

# Feishu credentials
export HOTPLEX_FEISHU_APP_ID="cli_..."
export HOTPLEX_FEISHU_APP_SECRET="..."
export HOTPLEX_FEISHU_VERIFICATION_TOKEN="..."
export HOTPLEX_FEISHU_ENCRYPT_KEY="..."

# Message storage
export HOTPLEX_MESSAGE_STORE_SQLITE_PATH="/data/messages.db"
```

### 3. Customize as Needed

Edit the copied YAML files to adjust:
- `system_prompt`: Customize AI identity and behavior
- `task_instructions`: Add task-specific guidance
- Security policies (`dm_policy`, `group_policy`)
- Feature toggles (chunking, threading, markdown)
- Capabilities for App Home

## Environment Variable Syntax

**IMPORTANT**: Go's `os.ExpandEnv` only supports basic variable substitution:

```yaml
# Supported
bot_user_id: ${HOTPLEX_SLACK_BOT_USER_ID}

# NOT supported (shell-style defaults)
bot_user_id: ${HOTPLEX_SLACK_BOT_USER_ID:-}  # Will NOT work!
```

If an environment variable is not set, it will be replaced with an empty string. Ensure all required variables are defined.

## Configuration Priority

Configuration values are loaded in this priority order (highest wins):

1. Environment variables
2. Custom config files (e.g., `configs/chatapps/slack.yaml`)
3. Base templates (`configs/base/`)

## Multi-Bot Deployment

For running multiple bots, each bot must have a **unique `bot_user_id`** to prevent session ID collisions:

```yaml
# Bot 1
security:
  permission:
    bot_user_id: ${HOTPLEX_SLACK_BOT_USER_ID}  # Bot A

# Bot 2
security:
  permission:
    bot_user_id: ${HOTPLEX_SLACK_BOT2_USER_ID}  # Bot B
```

## Further Documentation

- Slack Manual: `docs/chatapps/chatapps-slack-manual.md`
- Feishu Manual: `docs/chatapps/chatapps-feishu-manual.md`
- Main Documentation: `CLAUDE.md`
