# CyberArk Workload Password Management Summon Provider <!-- omit in toc -->

[![Test and Build](https://github.com/infamousjoeg/summon-wpm/actions/workflows/test-and-build.yml/badge.svg)](https://github.com/infamousjoeg/summon-wpm/actions/workflows/test-and-build.yml)
[![Release](https://github.com/infamousjoeg/summon-wpm/actions/workflows/release.yml/badge.svg)](https://github.com/infamousjoeg/summon-wpm/actions/workflows/release.yml)

A [Summon](https://github.com/cyberark/summon) provider for retrieving credentials from CyberArk Workload Password Management.

This provider supports both interactive and non-interactive authentication methods:
- Interactive: Normal users can authenticate with MFA challenges
- Non-interactive: Service users can authenticate with OIDC client credentials

## Table of Contents <!-- omit in toc -->

- [Installation](#installation)
  - [From Source](#from-source)
  - [From GitHub Releases](#from-github-releases)
    - [Linux](#linux)
    - [macOS](#macos)
    - [Windows](#windows)
- [Usage](#usage)
  - [Setup and Configuration](#setup-and-configuration)
  - [Interactive Authentication](#interactive-authentication)
  - [Using with Summon](#using-with-summon)
  - [Non-Interactive Usage](#non-interactive-usage)
- [Command Line Options](#command-line-options)
- [Environment Variables](#environment-variables)
- [Configuration File Location](#configuration-file-location)
- [Development](#development)
  - [Running Tests](#running-tests)
  - [Building Locally](#building-locally)
- [Security Considerations](#security-considerations)
- [License](#license)


## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/infamousjoeg/summon-wpm.git
cd summon-wpm

# Build the binary
go build -o summon-wpm

# Move to a location in your PATH
sudo mv summon-wpm /usr/local/bin/
```

### From GitHub Releases

#### Linux

```bash
# Download the latest release
VERSION=$(curl -s https://api.github.com/repos/infamousjoeg/summon-wpm/releases/latest | grep -Po '"tag_name": "\K[^"]*')
curl -sL "https://github.com/infamousjoeg/summon-wpm/releases/download/${VERSION}/summon-wpm-${VERSION#v}-linux-amd64.tar.gz" | tar xz

# Install into Summon providers directory
sudo mkdir -p /usr/local/lib/summon
sudo mv summon-wpm-linux-amd64 /usr/local/lib/summon/summon-wpm
sudo chmod +x /usr/local/lib/summon/summon-wpm
```

#### macOS

```bash
# Download the latest release (universal binary for both Intel and Apple Silicon)
VERSION=$(curl -s https://api.github.com/repos/infamousjoeg/summon-wpm/releases/latest | grep -Po '"tag_name": "\K[^"]*')
curl -sL "https://github.com/infamousjoeg/summon-wpm/releases/download/${VERSION}/summon-wpm-${VERSION#v}-darwin-universal.tar.gz" | tar xz

# Install into Summon providers directory
sudo mkdir -p /usr/local/lib/summon
sudo mv summon-wpm-darwin-universal /usr/local/lib/summon/summon-wpm
sudo chmod +x /usr/local/lib/summon/summon-wpm
```

#### Windows

```powershell
# Download latest release version
$VERSION = (Invoke-RestMethod -Uri "https://api.github.com/repos/infamousjoeg/summon-wpm/releases/latest").tag_name

# Download and extract the ZIP file
$URL = "https://github.com/infamousjoeg/summon-wpm/releases/download/$VERSION/summon-wpm-$($VERSION.Substring(1))-windows-amd64.zip"
Invoke-WebRequest -Uri $URL -OutFile "summon-wpm.zip"
Expand-Archive -Path "summon-wpm.zip" -DestinationPath "."
Remove-Item "summon-wpm.zip"

# Create Summon providers directory and install
$SummonDir = "$env:ProgramFiles\summon\lib"
New-Item -ItemType Directory -Path $SummonDir -Force
Copy-Item "summon-wpm-windows-amd64.exe" -Destination "$SummonDir\summon-wpm.exe"

# Add to PATH if not already there
if ($env:Path -notlike "*$SummonDir*") {
    [Environment]::SetEnvironmentVariable("Path", $env:Path + ";$SummonDir", [EnvironmentVariableTarget]::Machine)
    $env:Path += ";$SummonDir"
    Write-Host "Added Summon directory to PATH"
}
```

## Usage

### Setup and Configuration

Before using the provider, you need to configure it with your CyberArk Identity tenant information:

```bash
summon-wpm --config
```

This will prompt you for:
- Tenant URL
- Username
- (Optional) Client ID and Secret for service account

### Interactive Authentication

```bash
summon-wpm --login
```

This will initiate an interactive authentication flow, presenting available authentication mechanisms and prompting for responses.

### Using with Summon

Once configured, you can use this provider with Summon:

```bash
summon -p summon-wpm \
  --yaml 'DB_PASSWORD: !var "my-app-credentials"' \
  your-command
```

This will:
1. Authenticate to CyberArk Identity (if not already authenticated)
2. Retrieve the password for "my-app-credentials"
3. Make it available as DB_PASSWORD environment variable to your-command

### Non-Interactive Usage

For non-interactive environments (like CI/CD pipelines), configure the provider with a service account:

```bash
summon-wpm --config
# Enter tenant URL, username, and client credentials
```

Then use it as normal:

```bash
summon -p summon-wpm \
  --yaml 'API_KEY: !var "api-credentials"' \
  your-command
```

## Command Line Options

- `--help` or `-h`: Show help information
- `--version` or `-v`: Show version information
- `--config`: Run the configuration wizard
- `--login`: Authenticate to CyberArk Identity
- `--verbose`: Enable verbose output

## Environment Variables

- `SUMMON_WPM_CONFIG_DIR`: Override the default config directory location

## Configuration File Location

The configuration is stored in:
- **Linux/macOS**: `$XDG_CONFIG_HOME/summon-wpm/cyberark-wpm.json` or `$HOME/.config/summon-wpm/cyberark-wpm.json`
- **Windows**: `%APPDATA%\summon-wpm\cyberark-wpm.json`

## Development

### Running Tests

```bash
go test -v ./...
```

### Building Locally

```bash
make build
```

## Security Considerations

- The configuration file contains sensitive information and is stored with permissions restricted to the current user
- Authentication tokens are cached to minimize authentication requests
- For production environments, consider using a dedicated service account

## License

[MIT License](LICENSE)