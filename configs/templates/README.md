# Role Templates

This directory contains system prompt templates for different bot roles.

## Available Roles

| Role | Description |
|------|-------------|
| `go` | Go Backend Engineer |
| `frontend` | React/Next.js Frontend Engineer |
| `devops` | Docker/K8s DevOps Engineer |
| `custom` | User-defined (edit manually) |

## Usage

These templates are used by `docker/matrix/add-bot.sh` when creating new bot instances.

### Customization

For custom roles:
1. Copy `custom.yaml` to your own template
2. Edit the `system_prompt` content
3. Reference it in add-bot.sh or use directly in bot config
