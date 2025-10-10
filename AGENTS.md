# Repository Guidelines

## Project Structure & Module Organization
- `cmd/omnidrop-server` holds the entrypoint wired for dependency injection.
- `internal/` contains domain packages (`app`, `handlers`, `services`, `middleware`, `observability`) designed for Go's internal visibility; add new logic here.
- `docs/` captures architecture notes, while `deployments/` and `init/` hold LaunchAgent templates and service assets.
- `scripts/` and `build/` provide helper automation; integration fixtures and mocks live in `test/`.

## Build, Test, and Development Commands
- `make build` compiles the server into `build/bin/omnidrop-server`.
- `make run` launches the locally built binary; use `TOKEN=dev-token make dev` for a live `go run`.
- `make test` runs `go test ./...` including integration suites; add `-run <pattern>` when debugging.
- `make install` provisions the LaunchAgent, AppleScript, and workspace for macOS deployments.
- `make clean` and `make deps` clear artifacts and sync module dependencies.

## Coding Style & Naming Conventions
- Format Go code with `go fmt ./...` (CI enforces standard gofmt layout and tabs for indentation).
- Favor package-level constructors (`NewService`, etc.) and descriptive error wrappers using `fmt.Errorf("context: %w", err)`.
- File names should be lowercase with underscores only for tests (e.g., `server_test.go`); keep exported identifiers PascalCase and unexported camelCase.

## Testing Guidelines
- Unit tests colocate with code in `internal/...`; maintain the `*_test.go` suffix and table-driven style.
- Integration tests reside in `test/integration`; run them via `go test ./test/integration -v` once the OmniFocus mock is configured.
- Maintain meaningful assertions and log silencing in tests to keep CI output readable; target full package coverage with `go test -cover ./...` before submitting PRs.

## Commit & Pull Request Guidelines
- Follow the existing title pattern: `<emoji> <scope>: <imperative message>` (e.g., `🧪 test: update server tests for OAuth integration`).
- Keep commits focused; include refactorings or generated assets separately with clear scopes (`build`, `docs`, `ci`, etc.).
- Pull requests should explain the change, list testing performed (`make test`, manual flows), link related issues, and attach screenshots or logs for UI/observability updates.

## Security & Configuration Tips
- Never commit `.env`; duplicate `.env.example` when adding required variables such as `TOKEN`, `OMNIDROP_ENV`, or `PORT`.
- Restrict sensitive paths by honoring `OMNIDROP_FILES_DIR`; validate new file operations against traversal protections already present in `internal/services`.
- When touching authentication, update docs in `docs/` and regenerate fixtures in `test/mocks` to keep contract expectations aligned.
