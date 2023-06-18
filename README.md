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

## Troubleshooting

* Config files are located in `%APPDATA%\ficsit\`

## Development

### Dependencies

* [Go 1.18](https://go.dev/doc/install)
* IDE of Choice. Goland or VSCode suggested.

## Building

```bash
go build
```

Will produce `ficsit-cli.exe` in the repo root directory.
