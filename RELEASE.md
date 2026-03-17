# ZeaShell Release Process

This document describes the complete release process for ZeaShell, including creating releases, updating the Homebrew tap, and updating documentation.

## Prerequisites

Before starting a release, ensure you have:

- [ ] Write access to the `open-tempest-labs/zeashell` repository
- [ ] Write access to the `open-tempest-labs/homebrew-zeashell` repository
- [ ] GitHub CLI (`gh`) installed and authenticated
- [ ] Go installed (for testing builds)
- [ ] Homebrew installed (for testing the tap)
- [ ] All changes merged to `main` branch
- [ ] Working tree is clean (`git status` shows no uncommitted changes)

## Version Numbering

ZeaShell follows [Semantic Versioning](https://semver.org/):

- **MAJOR.MINOR.PATCH** (e.g., `0.1.0`, `1.2.3`)
- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

The version is defined in `internal/cli/root.go`:

```go
Version: "0.1.0",
```

## Release Checklist

### 1. Pre-Release Preparation

- [ ] Review all commits since last release
- [ ] Update version in `internal/cli/root.go` if needed
- [ ] Update CHANGELOG.md (if exists) with new features, fixes, and breaking changes
- [ ] Test all major features locally
- [ ] Run tests: `go test ./...`
- [ ] Build binary: `go build -o zea ./cmd/zea`
- [ ] Verify binary works: `./zea --version`

### 2. Create Git Tag

Create an annotated tag with release notes:

```bash
git tag -a v0.X.Y -m "$(cat <<'EOF'
Release v0.X.Y - [Release Title]

Major Features:
- Feature 1
- Feature 2
- Feature 3

Bug Fixes:
- Fix 1
- Fix 2

Breaking Changes (if any):
- Change 1

Commands:
- List of available commands

Key Capabilities:
- Capability 1
- Capability 2
EOF
)"
```

Verify the tag:

```bash
git tag -l -n20 v0.X.Y
```

### 3. Push Tag to GitHub

```bash
git push origin v0.X.Y
```

This will make the tag available on GitHub.

### 4. Create GitHub Release

Create a release with comprehensive notes:

```bash
gh release create v0.X.Y \
  --title "ZeaShell v0.X.Y - [Release Title]" \
  --notes "$(cat <<'EOF'
# ZeaShell v0.X.Y - [Release Title]

**DataFrame Shell - CSV to petabytes, one pipe at a time**

Brief description of this release.

## 🎉 Major Features

### Feature Category 1
- Feature description
- Feature details

### Feature Category 2
- Feature description
- Feature details

## 📦 Installation

### Homebrew (macOS/Linux)
\`\`\`bash
brew tap open-tempest-labs/zeashell
brew install zeashell
\`\`\`

### Using Go Install
\`\`\`bash
go install github.com/open-tempest-labs/zeashell/cmd/zea@v0.X.Y
\`\`\`

### From Source
\`\`\`bash
git clone https://github.com/open-tempest-labs/zeashell
cd zeashell
git checkout v0.X.Y
go build -o zea ./cmd/zea
\`\`\`

## 🚀 Quick Start

\`\`\`bash
# Example commands
zea load data.csv | zea filter "amount > 100"
\`\`\`

## 📚 Documentation

See [README.md](https://github.com/open-tempest-labs/zeashell/blob/main/README.md) for complete documentation.

## 🔄 Upgrading

### Homebrew
\`\`\`bash
brew update
brew upgrade zeashell
\`\`\`

### Go Install
\`\`\`bash
go install github.com/open-tempest-labs/zeashell/cmd/zea@v0.X.Y
\`\`\`

---

**Full Changelog**: https://github.com/open-tempest-labs/zeashell/compare/v0.PREV.VERSION...v0.X.Y
EOF
)"
```

### 5. Calculate Source Tarball SHA256

Download the source tarball and calculate its SHA256:

```bash
cd /tmp
curl -L -o zeashell-0.X.Y.tar.gz \
  https://github.com/open-tempest-labs/zeashell/archive/refs/tags/v0.X.Y.tar.gz
sha256sum zeashell-0.X.Y.tar.gz
```

Copy the SHA256 hash for the next step.

### 6. Update Homebrew Formula

Clone or update the Homebrew tap repository:

```bash
cd /path/to/projects
git clone https://github.com/open-tempest-labs/homebrew-zeashell.git
# OR if already cloned:
cd homebrew-zeashell
git pull origin main
```

Update `zeashell.rb` with the new version:

```ruby
class Zeashell < Formula
  desc "DataFrame shell for CSV, JSON, XML, Parquet with Unix pipe semantics"
  homepage "https://github.com/open-tempest-labs/zeashell"
  url "https://github.com/open-tempest-labs/zeashell/archive/refs/tags/v0.X.Y.tar.gz"
  sha256 "PASTE_SHA256_HERE"  # From step 5
  license "Apache-2.0"
  head "https://github.com/open-tempest-labs/zeashell.git", branch: "main"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w"), "./cmd/zea"
  end

  test do
    # Test version command
    assert_match "0.X.Y", shell_output("#{bin}/zea --version")

    # Add more tests as needed
    # ...
  end
end
```

Verify the formula syntax:

```bash
ruby -c zeashell.rb
```

### 7. Commit and Push Formula Update

```bash
git add zeashell.rb
git commit -m "Update ZeaShell to v0.X.Y

Updated formula for ZeaShell v0.X.Y release.

Changes:
- Updated version to 0.X.Y
- Updated SHA256 hash
- [List any other changes]
"
git push origin main
```

### 8. Test the Homebrew Installation

```bash
# Remove old tap if testing locally
brew untap open-tempest-labs/zeashell

# Add the tap
brew tap open-tempest-labs/zeashell

# Check formula info
brew info zeashell

# Install (or reinstall)
brew install zeashell
# OR for testing updates:
brew reinstall zeashell

# Verify version
zea --version

# Run basic tests
echo "name,value" > /tmp/test.csv
echo "Alice,100" >> /tmp/test.csv
echo "Bob,200" >> /tmp/test.csv

zea load /tmp/test.csv | zea filter "value > 150"
# Should only show Bob

# Clean up
rm /tmp/test.csv
```

### 9. Update Main Repository Documentation

If needed, update the main repository's README.md with any new features or changes:

```bash
cd /path/to/zeashell
# Edit README.md as needed
git add README.md
git commit -m "Update documentation for v0.X.Y"
git push origin main
```

### 10. Announce the Release

Consider announcing the release:

- [ ] Update project homepage/website
- [ ] Post to social media (Twitter/X, LinkedIn, etc.)
- [ ] Announce in relevant communities (Reddit, HN, etc.)
- [ ] Send email to users/subscribers
- [ ] Update project dependencies if this is a library

## Post-Release

### Verify Release

- [ ] Check GitHub release page is accessible
- [ ] Verify source tarball downloads correctly
- [ ] Test Homebrew installation on fresh machine
- [ ] Test `go install` installation
- [ ] Verify all download links work

### Monitor

- [ ] Watch for bug reports
- [ ] Monitor installation issues
- [ ] Check formula audit results (if any)

## Troubleshooting

### Formula Doesn't Install

**Problem**: Homebrew can't install the formula

**Solutions**:
1. Verify the SHA256 hash is correct
2. Check the tarball URL is accessible
3. Ensure Go dependency is available
4. Run `brew audit --strict zeashell` to check for issues

### Version Mismatch

**Problem**: `zea --version` shows wrong version

**Solution**: Ensure `internal/cli/root.go` was updated and committed before tagging

### SHA256 Mismatch

**Problem**: Homebrew reports SHA256 mismatch

**Solution**: Recalculate SHA256 from the actual GitHub release tarball:
```bash
curl -L https://github.com/open-tempest-labs/zeashell/archive/refs/tags/v0.X.Y.tar.gz | sha256sum
```

### Tag Already Exists

**Problem**: Tag already exists and needs to be recreated

**Solution**: Delete the tag locally and remotely:
```bash
git tag -d v0.X.Y
git push origin :refs/tags/v0.X.Y
# Then recreate the tag
```

## Release Template

Use this template for consistent release notes:

```markdown
# ZeaShell v0.X.Y - [Release Title]

**Brief description**

## 🎉 New Features
- Feature 1
- Feature 2

## 🐛 Bug Fixes
- Fix 1
- Fix 2

## ⚡ Performance Improvements
- Improvement 1

## 📝 Documentation
- Doc update 1

## 🔨 Breaking Changes (if any)
- Change 1 (how to migrate)

## 📦 Installation

[Standard installation instructions]

## 🙏 Contributors

Thanks to all contributors for this release!
```

## Version History

| Version | Release Date | Highlights |
|---------|-------------|------------|
| v0.1.0  | 2026-03-16  | Initial release with multi-format support, HTTP/HTTPS loading, path-based columns |

## Additional Resources

- [Semantic Versioning](https://semver.org/)
- [GitHub Release Documentation](https://docs.github.com/en/repositories/releasing-projects-on-github)
- [Homebrew Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
- [Keep a Changelog](https://keepachangelog.com/)
