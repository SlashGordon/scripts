# NAS Manager

[![CI](https://github.com/SlashGordon/scripts/actions/workflows/ci.yml/badge.svg)](https://github.com/SlashGordon/scripts/actions/workflows/ci.yml)
[![Release](https://github.com/SlashGordon/scripts/actions/workflows/release.yml/badge.svg)](https://github.com/SlashGordon/scripts/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/SlashGordon/scripts)](https://goreportcard.com/report/github.com/SlashGordon/scripts)
[![Latest Release](https://img.shields.io/github/v/release/SlashGordon/scripts)](https://github.com/SlashGordon/scripts/releases/latest)

A CLI tool for managing scripts and tasks on your NAS system.

## Features

- **Script Management**: Run and list shell/Python scripts
- **Task Management**: Monitor system status and processes  
- **ACME Certificates**: Issue/renew Let's Encrypt certificates via Cloudflare DNS

## Installation

Download the appropriate binary for your system from the releases page.

## Configuration

Configuration is loaded in this priority order:
1. `NAS_CONFIG` environment variable (custom path)
2. `.nasrc` in working directory
3. `.nasrc` in home directory
4. Environment variables

Copy `.nasrc.example` to `.nasrc` and set your credentials:

```bash
cp .nasrc.example .nasrc
# Edit .nasrc with your actual values
```

Or use a custom config path:
```bash
NAS_CONFIG=/path/to/config nas-manager ddns update
```

Required environment variables:
- `CF_API_TOKEN` - Cloudflare API token (used for both DDNS and ACME)
- `CF_ZONE_ID` - Cloudflare zone ID (for DDNS)

## Usage

```bash
# Show help
./nas-manager --help

# Script commands
./nas-manager script run /path/to/script.sh
./nas-manager script list /scripts/directory

# Task commands  
./nas-manager task status
./nas-manager task ps nginx

# ACME certificate management
./nas-manager acme issue
```

## Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Clean build artifacts
make clean
```

## Release

Binaries are automatically built for:
- Linux (amd64, arm64)
- macOS (amd64, arm64)