#!/bin/bash

set -e

VERSION=$1

if [ -z "$VERSION" ]; then
    echo "Usage: ./scripts/release.sh <version>"
    echo "Example: ./scripts/release.sh 0.1.0"
    exit 1
fi

echo "Releasing v$VERSION..."

# Ensure we're on main branch
BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [ "$BRANCH" != "main" ]; then
    echo "Error: Must be on main branch to release"
    exit 1
fi

# Ensure working directory is clean
if [ -n "$(git status --porcelain)" ]; then
    echo "Error: Working directory is not clean"
    exit 1
fi

# Update npm package version
echo "Updating npm package version..."
cd npm
npm version "$VERSION" --no-git-tag-version
cd ..

# Commit version update
git add npm/package.json
git commit -m "Bump version to v$VERSION"

# Create git tag
git tag -a "v$VERSION" -m "Release v$VERSION"

# Push
git push origin main
git push origin "v$VERSION"

echo ""
echo "Tag pushed. GitHub Actions will build and release."
echo ""
echo "After the release is complete, publish the npm package:"
echo "  cd npm && npm publish"
echo ""
echo "Release complete!"
