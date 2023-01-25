#!/usr/bin/env bash
git fetch --tags -q
version=$(git tag -l --sort=-version:refname v* | head -n 1)
echo "From: $version"
a=( ${version//./ } )
b=( ${a[0]//v/ } )

major=${b[0]}
minor=${a[1]}
patch=${a[2]}

case $1 in
  patch)
    ((patch++))
  ;;
  minor)
    patch=0
    ((minor++))
  ;;
  major)
    patch=0
    minor=0
    ((major++))
  ;;
  *)
    echo "Invalid level passed"
    return 2
esac
echo "To: v${major}.${minor}.${patch}"
git tag -a "v${major}.${minor}.${patch}" -m "v${major}.${minor}.${patch}"