# Contributing to pass-along

Contributions are welcome. Here are the guidelines and requirements for proposing changes.

## Development Setup

The project is written in Go and requires Go 1.24 or newer.

Clone the repository and run:

```bash
go mod download
```

To build the executable:

```bash
go build -o pass-along .
```

To run the application locally:

```bash
go run .
```

## Running Tests

All automated tests should pass locally before opening a pull request.

Run standard tests:

```bash
go test ./...
```

Run tests with race detection:

```bash
go test -race ./...
```

Run fuzz tests:

```bash
go test ./... -run=^$ -fuzz=Fuzz -fuzztime=10s
```

## Testing Policy

1. Every pull request introducing new features, datastore modifications, or bug fixes must include corresponding automated unit tests.
2. Changes affecting parsing, serialization, or cryptographic workflows must update or add fuzz tests.
3. Tests run automatically on pull requests via GitHub Actions.

## Coding Standards

1. Code must be formatted with standard Go tooling (`go fmt ./...`).
2. Code must pass `golangci-lint` without errors or unaddressed warnings.
3. Keep external dependencies minimal and vetted.

## Proposing Changes

1. Fork the repository and create a descriptive feature branch.
2. Commit changes with clear, descriptive commit messages.
3. Open a pull request against the `master` branch.
4. Ensure all CI checks (tests, fuzz tests, linting, CodeQL) pass.
