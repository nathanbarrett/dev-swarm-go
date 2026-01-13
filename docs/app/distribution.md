# Distribution

## Overview

dev-swarm-go is distributed as a Go binary via npm. This approach:
- Leverages npm for easy installation (`npm install -g dev-swarm-go`)
- Uses GoReleaser for cross-platform binary builds
- Downloads platform-specific binary at install time

## Installation Flow

```
npm install -g dev-swarm-go
        │
        ▼
  Run postinstall script
        │
        ▼
  Detect platform (darwin/linux) and arch (amd64/arm64)
        │
        ▼
  Download binary from GitHub Releases
        │
        ▼
  Extract and make executable
        │
        ▼
  Shell wrapper invokes binary
```

## Package Structure

```
npm/
├── package.json              # npm package definition
├── install.js                # Post-install binary downloader
└── bin/
    └── dev-swarm-go          # Shell wrapper script
```

### package.json

Key fields:
- `name`: "dev-swarm-go"
- `bin`: Points to shell wrapper
- `scripts.postinstall`: Runs install.js
- `os`: darwin, linux
- `cpu`: x64, arm64

### install.js

Post-install script that:
1. Detects platform and architecture
2. Constructs download URL from GitHub Releases
3. Downloads the appropriate tarball
4. Extracts binary to `bin/dev-swarm-go-binary`
5. Sets executable permissions

### Shell Wrapper

The `bin/dev-swarm-go` script:
1. Locates the downloaded binary
2. Passes all arguments through
3. Provides helpful error if binary missing

## Build Process

### Local Development

Build for current platform:
1. Run build script with version
2. Binary output to `bin/dev-swarm-go`
3. Version info injected via ldflags

### Release Build

GoReleaser handles cross-platform builds:

**Platforms:**
- darwin/amd64 (Intel Mac)
- darwin/arm64 (Apple Silicon)
- linux/amd64 (x86_64)
- linux/arm64 (ARM64)

**Build settings:**
- CGO disabled for portability
- Static linking
- Version/commit/date injected

## Release Process

1. **Tag the release**
   - Create annotated git tag: `v{version}`
   - Push tag to GitHub

2. **GoReleaser runs** (via GitHub Actions)
   - Builds all platform binaries
   - Creates tarballs
   - Uploads to GitHub Releases
   - Generates checksums

3. **Publish npm package**
   - Update version in package.json
   - Run `npm publish`
   - Package references same version for downloads

## Version Management

Version info comes from:
- Git tag (primary)
- Commit hash (for development builds)
- Build timestamp

Injected via Go ldflags to `pkg/version` package.

## Platform Support

| Platform | Architecture | Status |
|----------|-------------|--------|
| macOS | Intel (amd64) | Supported |
| macOS | Apple Silicon (arm64) | Supported |
| Linux | x86_64 (amd64) | Supported |
| Linux | ARM64 | Supported |
| Windows | Any | Not supported |

## Dependencies

End users need:

| Dependency | Purpose | Installation |
|------------|---------|--------------|
| Node.js | npm package installation | nodejs.org |
| gh | GitHub CLI for all GitHub operations | cli.github.com |
| claude | Claude Code CLI for AI sessions | anthropic.com |
| git | Repository operations | git-scm.com |

## Troubleshooting

### Binary not found
- Re-run `npm install -g dev-swarm-go`
- Check for download errors in npm output
- Verify platform is supported

### Download fails
- Check network connectivity
- Verify GitHub Releases are accessible
- Check for proxy/firewall issues

### Wrong architecture
- Check `uname -m` output
- Verify npm sees correct platform
- May need to reinstall after architecture change (e.g., Rosetta)
