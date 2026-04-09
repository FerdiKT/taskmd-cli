# Release Workflow

## Versioned release

```bash
make test
git tag v0.1.0
git push origin main --tags
```

The GitHub Actions release workflow will:

1. Build multi-arch tarballs with `make brew-dist`
2. Create a GitHub release and upload `dist/*`
3. Update `FerdiKT/homebrew-tap` formula `taskmd`

## Local release dry run

```bash
make brew-dist VERSION=0.1.0
ls dist/
cat dist/checksums.txt
```

## Homebrew install

```bash
brew tap FerdiKT/tap
brew install taskmd
```

