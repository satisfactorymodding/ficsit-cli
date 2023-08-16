# ficsit-cli [![push](https://github.com/Vilsol/ficsit-cli/actions/workflows/push.yaml/badge.svg)](https://github.com/Vilsol/ficsit-cli/actions/workflows/push.yaml) ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/vilsol/ficsit-cli) ![GitHub tag (latest by date)](https://img.shields.io/github/v/tag/vilsol/ficsit-cli) [![GitHub license](https://img.shields.io/github/license/Vilsol/ficsit-cli)](https://github.com/Vilsol/ficsit-cli/blob/master/LICENSE) ![GitHub all releases](https://img.shields.io/github/downloads/vilsol/ficsit-cli/total)

A CLI tool for managing mods for the game Satisfactory

## Installation

### Windows

Download the appropriate `.exe` for your CPU architecture.

* [AMD64 (64-bit)](https://github.com/Vilsol/ficsit-cli/releases/latest/download/ficsit_windows_amd64.exe)
* [386 (32-bit)](https://github.com/Vilsol/ficsit-cli/releases/latest/download/ficsit_windows_386.exe)
* [ARM64 (64-bit ARM)](https://github.com/Vilsol/ficsit-cli/releases/latest/download/ficsit_windows_arm64.exe)
* [ARMv7 (32-bit ARM)](https://github.com/Vilsol/ficsit-cli/releases/latest/download/ficsit_windows_armv7.exe)

### Linux

#### Arch

A package is published to AUR under the name [`ficsit-cli-bin`](https://aur.archlinux.org/packages/ficsit-cli-bin)

```shell
yay -S ficsit-cli-bin
```

#### Debian (inc. Ubuntu, Mint, PopOS!, etc)

Download the appropriate `.deb` for your CPU architecture.

* [AMD64 (64-bit)](https://github.com/Vilsol/ficsit-cli/releases/latest/download/ficsit_linux_amd64.deb)
* [386 (32-bit)](https://github.com/Vilsol/ficsit-cli/releases/latest/download/ficsit_linux_386.deb)
* [ARM64 (64-bit ARM)](https://github.com/Vilsol/ficsit-cli/releases/latest/download/ficsit_linux_arm64.deb)
* [ARMv7 (32-bit ARM)](https://github.com/Vilsol/ficsit-cli/releases/latest/download/ficsit_linux_armv7.deb)
* [PowerPC64](https://github.com/Vilsol/ficsit-cli/releases/latest/download/ficsit_linux_ppc64le.deb)

#### Fedora

Download the appropriate `.rpm` for your CPU architecture.

* [AMD64 (64-bit)](https://github.com/Vilsol/ficsit-cli/releases/latest/download/ficsit_linux_amd64.rpm)
* [386 (32-bit)](https://github.com/Vilsol/ficsit-cli/releases/latest/download/ficsit_linux_386.rpm)
* [ARM64 (64-bit ARM)](https://github.com/Vilsol/ficsit-cli/releases/latest/download/ficsit_linux_arm64.rpm)
* [ARMv7 (32-bit ARM)](https://github.com/Vilsol/ficsit-cli/releases/latest/download/ficsit_linux_armv7.rpm)
* [PowerPC64](https://github.com/Vilsol/ficsit-cli/releases/latest/download/ficsit_linux_ppc64le.rpm)

#### Alpine

Download the appropriate `.apk` for your CPU architecture.

* [AMD64 (64-bit)](https://github.com/Vilsol/ficsit-cli/releases/latest/download/ficsit_linux_amd64.apk)
* [386 (32-bit)](https://github.com/Vilsol/ficsit-cli/releases/latest/download/ficsit_linux_386.apk)
* [ARM64 (64-bit ARM)](https://github.com/Vilsol/ficsit-cli/releases/latest/download/ficsit_linux_arm64.apk)
* [ARMv7 (32-bit ARM)](https://github.com/Vilsol/ficsit-cli/releases/latest/download/ficsit_linux_armv7.apk)
* [PowerPC64](https://github.com/Vilsol/ficsit-cli/releases/latest/download/ficsit_linux_ppc64le.apk)

### macOS

Download the "all" build [here](https://github.com/Vilsol/ficsit-cli/releases/latest/download/ficsit_darwin_all).

## Usage

### Interactive CLI

To launch the interactive CLI, run the executable without any arguments.

### Command Line

Run `ficsit help` to see a list of available commands.

## Troubleshooting

* Profile and installation records are located in `%APPDATA%\ficsit\`
* Downloads are cached in `%LOCALAPPDATA%\ficsit\downloadCache\`

## Development

### Dependencies

* [Go 1.18](https://go.dev/doc/install)
* IDE of Choice. Goland or VSCode suggested.

### Code Generation

If you update any of the GraphQL queries, run this to update generated code:

```bash
(echo "y") | npx graphqurl https://api.ficsit.app/v2/query --introspect -H 'content-type: application/json' > schema.graphql
go generate -tags tools -x ./...
```

## Building

```bash
go build
```

Will produce `ficsit-cli.exe` in the repo root directory.

### Linting

Install `golangci-lint` via the directions [here](https://golangci-lint.run/usage/install/#local-installation),
but make sure to install the version specified in `.github/workflows/push.yaml` instead of whatever it suggests.

Then, to run it, use:

```bash
golangci-lint run --fix
```

### Updating generated docs

The files within `./docs` are generated using cobra, use the following to update
them.

```
go run tools.go
```