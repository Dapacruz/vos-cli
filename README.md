# vos-cli

A lightweight command-line utility for executing commands on Cisco Versa Operating System (VOS) devices via SSH.

[![Go Version](https://img.shields.io/badge/Go-1.19%2B-blue)](https://go.dev/dl/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Version](https://img.shields.io/badge/version-0.10.0-green)](https://github.com/Dapacruz/vos-cli/releases)

---

## Table of Contents

- [Features](#features)
- [Requirements](#requirements)
- [Installation](#installation)
- [First Run & Configuration](#first-run--configuration)
- [Commands](#commands)
  - [device run commands](#device-run-commands)
  - [config](#config)
- [Usage Examples](#usage-examples)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [License](#license)

---

## Features

- Execute one or more CLI commands on Cisco VOS devices over SSH
- Run commands concurrently across multiple devices
- Support for both password-based and SSH key-based authentication
- Pipe a list of hostnames from stdin for scripted workflows
- Colorized, clearly formatted output per device and command
- Optional host key verification bypass for lab/non-production environments
- Persistent configuration file for storing default credentials

---

## Requirements

- Go 1.19 or later
- Linux or macOS (Windows is not supported)

---

## Installation

Install the latest version using `go install`:

```sh
go install github.com/Dapacruz/vos-cli@latest
```

Ensure your Go binary path is in your `$PATH`:

```sh
export PATH="$PATH:$(go env GOPATH)/bin"
```

---

## First Run & Configuration

On the first run, if no configuration file is found, `vos-cli` will prompt you to create one:

```
Initializing configuration file...

Default VOS User: admin
Default Password (admin): ********

Initialization complete.
Configuration file saved to /home/user/.vos-cli.yml.
```

The configuration file is stored at `~/.vos-cli.yml` with the following structure:

```yaml
user: "admin"
password: "yourpassword"
```

> **Security note:** The configuration file is automatically set to `600` permissions (readable only by the file owner).

### Config Commands

You can manage the configuration file using the built-in `config` subcommand:

| Command | Description |
|---|---|
| `vos-cli config list` | Print the path to the active config file |
| `vos-cli config show` | Print the contents of the config file |
| `vos-cli config edit` | Open the config file in your default editor |

To bypass the configuration file entirely, use the global `--no-config` flag.

---

## Commands

### `device run commands`

Executes one or more CLI commands on one or more Cisco VOS devices via SSH. Commands are run concurrently across all specified hosts.

```
vos-cli device run commands [flags] <device> [device]...
```

#### Flags

| Flag | Short | Default | Description |
|---|---|---|---|
| `--command` | `-c` | | Comma-separated list of commands to execute (required) |
| `--user` | | | VOS admin user (overrides config) |
| `--password` | | | Password for VOS user (overrides config) |
| `--password-stdin` | | `false` | Read password from stdin |
| `--key-based-auth` | `-k` | `false` | Use SSH key-based authentication (`~/.ssh/id_rsa`) |
| `--port` | `-p` | `22` | SSH port to connect to |
| `--expect-timeout` | `-e` | `30` | Timeout in seconds waiting for each command response |
| `--ssh-timeout` | `-S` | `30` | Timeout in seconds for the SSH connection |
| `--insecure` | `-K` | `false` | Skip host key verification (not recommended for production) |
| `--sort-output` | `-s` | `false` | Sort command output alphabetically |

#### Global Flags

| Flag | Description |
|---|---|
| `--no-config` | Bypass the configuration file |
| `--version` | Print the version and exit |

### `config`

Manage the `vos-cli` configuration file.

```
vos-cli config <subcommand>
```

| Subcommand | Description |
|---|---|
| `list` | Print the path to the active config file |
| `show` | Print the contents of the config file |
| `edit` | Open the config file in your default editor |

---

## Usage Examples

**Run a single command on one device:**

```sh
vos-cli device run commands --command "show version" router01.example.com
```

**Run multiple commands on one device:**

```sh
vos-cli device run commands --command "show version","show ip interface brief" router01.example.com
```

**Run a command on multiple devices concurrently:**

```sh
vos-cli device run commands --command "show ip route" router01.example.com router02.example.com
```

**Use SSH key-based authentication:**

```sh
vos-cli device run commands --command "show version" --key-based-auth router01.example.com
```

**Skip host key verification (lab environments):**

```sh
vos-cli device run commands --command "show version" --insecure router01.example.com
```

**Pipe a list of hosts from a file:**

```sh
cat hosts.txt | vos-cli device run commands --command "show version"
```

**Override credentials inline:**

```sh
vos-cli device run commands --user admin --password secret --command "show ip bgp summary" router01.example.com
```

**Read password from stdin (useful for scripting):**

```sh
echo "secret" | vos-cli device run commands --password-stdin --command "show version" router01.example.com
```

**Connect on a non-standard SSH port:**

```sh
vos-cli device run commands --port 2222 --command "show version" router01.example.com
```

---

## Troubleshooting

**`unable to load ssh known_hosts`**

The device's host key is not in your `~/.ssh/known_hosts` file. Either connect to the device manually with `ssh` first to accept the key, or use `--insecure` to skip host key checking (only for non-production use).

**`ssh.Dial failed: connection timed out`**

The device is unreachable or the SSH port is blocked. Verify network connectivity and that SSH is enabled on the device. Use `--ssh-timeout` to increase the connection timeout if needed.

**`no commands specified`**

You must provide at least one command using the `--command` / `-c` flag.

**`key based auth, saved password or password flag is required when reading hosts from stdin`**

When piping hosts from stdin, you cannot also be prompted interactively for a password. Use `--key-based-auth`, store a password in the config file, or pass `--password` directly.

**Configuration file not being picked up**

Run `vos-cli config list` to confirm the path to the active config file, and `vos-cli config show` to verify its contents.

---

## Contributing

Contributions are welcome. Please open an issue to discuss your proposed change before submitting a pull request.

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Commit your changes (`git commit -m 'Add my feature'`)
4. Push to the branch (`git push origin feature/my-feature`)
5. Open a pull request

---

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
