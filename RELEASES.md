# Releasing lmux

This project uses GoReleaser to build and publish multi-platform binaries and packages.

## Prerequisites

- GitHub repo: `sbcinnovation/lmux` (public)
- Homebrew tap: `sbcinnovation/homebrew-tap` with a `Formula/` folder (GoReleaser will open PRs/commits)
- Optional Scoop bucket: `sbcinnovation/scoop-bucket`

## Versioning

- Tag releases with semantic versioning: `vX.Y.Z`
- The `version` command is compiled with `-ldflags` to embed the tag and build metadata.

## How to cut a release

1. Update `CHANGELOG` (if used) and ensure `main` is green.
2. Create and push a tag:
   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```
3. GitHub Actions will run GoReleaser on tag push and publish:
   - Release assets (archives + checksums)
   - Linux packages (`.deb`, `.rpm`)
   - Homebrew formula update in `sbcinnovation/homebrew-tap`
   - Scoop manifest update in `sbcinnovation/scoop-bucket` (if configured)

## Local test (dry run)

```bash
# Install goreleaser locally if needed
# brew install goreleaser

goreleaser release --skip-publish --clean
```

## Updating taps/buckets

- Homebrew: GoReleaser pushes to `sbcinnovation/homebrew-tap` using `GITHUB_TOKEN`.
- Scoop: GoReleaser updates `sbcinnovation/scoop-bucket`.

## Update checks in lmux

- Users can run `lmux version --check` to see if a newer version is available. This queries `https://api.github.com/repos/sbcinnovation/lmux/releases/latest`.

## Troubleshooting

- If Homebrew tap commit fails, ensure repo exists and the workflow has `contents: write` permission.
- For RPM signing or custom repos, extend `.goreleaser.yaml` `nfpms` config.
