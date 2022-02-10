# Silta CLI

CI/CD deployment command abstraction, utilities and tools for Silta.
See https://github.com/wunderio/silta for more information about Silta.

## Usage

Run cli commands

```
silta version
```

List available commands
```
silta --help
```

Full usage documentation: [docs/silta.md](docs/silta.md)

## Building

If You need to build binary from sources, You'll need to install `go` first. 
Follow go [Download and install](https://go.dev/doc/install) document for 
installation steps. 

Building application:
```
make build
```

## Testing

```
go test ./tests
```