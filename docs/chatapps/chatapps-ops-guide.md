# Appendix: ChatApp Local Development & Quick Start

This document provides operational instructions for local development and testing of ChatApp integrations.

## 1. Quick Start

### Running Examples
```bash
# Enter HotPlex directory
go run _examples/chatapps_dingtalk/main.go
```

## 2. Local Development (Intranet Penetration)

Since chat platforms require a verifiable public Webhook URL, use the following tools during development:

```bash
# 1. Start local HotPlex
go run main.go

# 2. Start ngrok / cloudflared
ngrok http 8080

# 3. Fill the generated URL back into the platform console (e.g., DingTalk)
```

## 3. Environment Variables Reference

| Variable                          | Description                 | Example         |
| :-------------------------------- | :-------------------------- | :-------------- |
| `HOTPLEX_CHATAPPS_ADDR`           | Adapter listening port      | `:8080`         |
| `HOTPLEX_TELEGRAM_TOKEN`          | Telegram Bot Token          | `12345:ABCDE`   |
| `HOTPLEX_TELEGRAM_SECRET`         | Webhook security token      | `secure_secret` |
| `HOTPLEX_DINGTALK_APP_ID`         | DingTalk App Key            | `dingXXXX`      |
| `HOTPLEX_DINGTALK_APP_SECRET`     | DingTalk App Secret         | `secretXXXX`    |
| `HOTPLEX_DINGTALK_CALLBACK_TOKEN` | Callback verification token | `tokenXXXX`     |
