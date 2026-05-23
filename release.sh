#!/usr/bin/env bash
set -euo pipefail

# Usage: ./release.sh [minor|patch]
# Defaults to patch. Use 'minor' when adding new commands.

BUMP="${1:-patch}"

if [[ "$BUMP" != "minor" && "$BUMP" != "patch" ]]; then
  echo "Error: bump type must be 'minor' or 'patch' (got '$BUMP')" >&2
  exit 1
fi

VERSION_FILE="pkg/version/get_version.go"

# Extract current version
CURRENT=$(grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' "$VERSION_FILE")
if [[ -z "$CURRENT" ]]; then
  echo "Error: could not parse version from $VERSION_FILE" >&2
  exit 1
fi

MAJOR=$(echo "$CURRENT" | cut -d. -f1 | tr -d 'v')
MINOR=$(echo "$CURRENT" | cut -d. -f2)
PATCH=$(echo "$CURRENT" | cut -d. -f3)

if [[ "$BUMP" == "minor" ]]; then
  MINOR=$((MINOR + 1))
  PATCH=0
else
  PATCH=$((PATCH + 1))
fi

NEW_VERSION="v${MAJOR}.${MINOR}.${PATCH}"
TAG="release ${NEW_VERSION}"

echo "Bumping $CURRENT → $NEW_VERSION ($BUMP)"

# Update version file
sed -i '' "s/${CURRENT}/${NEW_VERSION}/" "$VERSION_FILE"

# Commit and tag
git add "$VERSION_FILE"
git commit -m "release ${NEW_VERSION}"
git tag "$NEW_VERSION" -m "$TAG"
git push origin main
git push origin "refs/tags/${NEW_VERSION}"

echo "Released ${NEW_VERSION}"
