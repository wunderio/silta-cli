# Silta CLI

CI/CD deployment command abstraction, utilities and tools for Silta.
See https://github.com/wunderio/silta for more information about Silta.

## Installing

Note: There are multiple binaries available for several systems, examples below install linux/amd64 version. Check release assets for more versions.
```
# Latest tagged release
latest_release_url=$(curl -sL https://api.github.com/repos/wunderio/silta-cli/releases/latest | grep -o -E "https://(.*)silta-(.*)-linux-amd64.tar.gz" | head -1)
curl -sL $latest_release_url | tar xz -C ~/.local/bin

# Selected release (i.e. 0.1.0)
curl -sL https://github.com/wunderio/silta-cli/releases/download/0.1.0/silta-0.1.0-linux-amd64.tar.gz | tar xz -C ~/.local/bin

# Latest build from master branch
curl -sL https://github.com/wunderio/silta-cli/releases/download/master/silta-master-linux-amd64.tar.gz | tar xz -C ~/.local/bin
```

### macOS
Installing should work using brew
```
brew install wunderio/tap/silta-cli
```

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
