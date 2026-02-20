# HotPlex Guidelines for Claude

## Build and Run
- Build the daemon: `make build`
- Run the daemon locally: `make run`
- Clean build artifacts: `make clean`

## Testing and Quality
- Run all tests: `make test`
- Run linter: `make lint`
- Install Git hooks: `make install-hooks`
- Tidy Go modules: `make tidy`

## Coding Standards
- **Style**: Standard Go formatting (run `make fmt` or `go fmt ./...`).
- **Safety**: Ensure all OS processes use PGID isolation (see `pkg/hotplex/sys_unix.go`).
- **Security**: All prompts must pass through the `Detector` in `pkg/hotplex/danger.go`.
- **Commits**: Follow [Conventional Commits](https://www.conventionalcommits.org/).
