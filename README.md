# ficsit-cli [![push](https://github.com/Vilsol/ficsit-cli/actions/workflows/push.yaml/badge.svg)](https://github.com/Vilsol/ficsit-cli/actions/workflows/push.yaml) ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/vilsol/ficsit-cli) ![GitHub tag (latest by date)](https://img.shields.io/github/v/tag/vilsol/ficsit-cli) [![GitHub license](https://img.shields.io/github/license/Vilsol/ficsit-cli)](https://github.com/Vilsol/ficsit-cli/blob/master/LICENSE) ![GitHub all releases](https://img.shields.io/github/downloads/vilsol/ficsit-cli/total)

A CLI tool for managing mods for the game Satisfactory

## Installation

<table>
  <tr>
    <th></th>
    <th>amd64</th>
    <th>386</th>
    <th>arm64</th>
    <th>armv7</th>
    <th>ppc64le</th>
  </tr>
  <tr>
    <th>Windows</th>
    <td><a href="https://github.com/Vilsol/ficsit-cli/releases/latest/download/ficsit_windows_amd64.exe">amd64</a></td>
    <td><a href="https://github.com/Vilsol/ficsit-cli/releases/latest/download/ficsit_windows_386.exe">386</a></td>
    <td><a href="https://github.com/Vilsol/ficsit-cli/releases/latest/download/ficsit_windows_arm64.exe">arm64</a></td>
    <td><a href="https://github.com/Vilsol/ficsit-cli/releases/latest/download/ficsit_windows_armv7.exe">armv7</a></td>
    <td>N/A</td>
  </tr>
  <tr>
    <th>Arch</th>
    <td colspan="5" style="text-align: center"><a href="https://aur.archlinux.org/packages/ficsit-cli-bin"><code>yay -S ficsit-cli-bin</code></a></td>
  </tr>
  <tr>
    <th>Debian</th>
    <td><a href="https://github.com/Vilsol/ficsit-cli/releases/latest/download/ficsit_linux_amd64.deb">amd64</a></td>
    <td><a href="https://github.com/Vilsol/ficsit-cli/releases/latest/download/ficsit_linux_386.deb">386</a></td>
    <td><a href="https://github.com/Vilsol/ficsit-cli/releases/latest/download/ficsit_linux_arm64.deb">arm64</a></td>
    <td><a href="https://github.com/Vilsol/ficsit-cli/releases/latest/download/ficsit_linux_armv7.deb">armv7</a></td>
    <td><a href="https://github.com/Vilsol/ficsit-cli/releases/latest/download/ficsit_linux_ppc64le.deb">ppc64le</a></td>
  </tr>
  <tr>
    <th>Fedora</th>
    <td><a href="https://github.com/Vilsol/ficsit-cli/releases/latest/download/ficsit_linux_amd64.rpm">amd64</a></td>
    <td><a href="https://github.com/Vilsol/ficsit-cli/releases/latest/download/ficsit_linux_386.rpm">386</a></td>
    <td><a href="https://github.com/Vilsol/ficsit-cli/releases/latest/download/ficsit_linux_arm64.rpm">arm64</a></td>
    <td><a href="https://github.com/Vilsol/ficsit-cli/releases/latest/download/ficsit_linux_armv7.rpm">armv7</a></td>
    <td><a href="https://github.com/Vilsol/ficsit-cli/releases/latest/download/ficsit_linux_ppc64le.rpm">ppc64le</a></td>
  </tr>
  <tr>
    <th>Alpine</th>
    <td><a href="https://github.com/Vilsol/ficsit-cli/releases/latest/download/ficsit_linux_amd64.apk">amd64</a></td>
    <td><a href="https://github.com/Vilsol/ficsit-cli/releases/latest/download/ficsit_linux_386.apk">386</a></td>
    <td><a href="https://github.com/Vilsol/ficsit-cli/releases/latest/download/ficsit_linux_arm64.apk">arm64</a></td>
    <td><a href="https://github.com/Vilsol/ficsit-cli/releases/latest/download/ficsit_linux_armv7.apk">armv7</a></td>
    <td><a href="https://github.com/Vilsol/ficsit-cli/releases/latest/download/ficsit_linux_ppc64le.apk">ppc64le</a></td>
  </tr>
  <tr>
    <th>macOS</th>
    <td colspan="4" style="text-align: center"><a href="https://github.com/Vilsol/ficsit-cli/releases/latest/download/ficsit_darwin_all">darwin_all</a></td>
    <td>N/A</td>
  </tr>
</table>


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

* [Go 1.21](https://go.dev/doc/install)
* IDE of Choice. Goland or VSCode suggested.

## Building

```bash
go build
```

Will produce `ficsit-cli.exe` in the repo root directory.
