#!/bin/bash
find ./modules -maxdepth 2 -mindepth 2 -type d -not -path '*/.*' | while read -r dir; do
  if [ "$dir" != "." ]; then
      echo "$(echo $dir | sed 's/^..//'): $(find "$dir" -type f -exec sha256sum {} + | awk '{print $1}' | sort | sha256sum | awk '{print $1}')"
  fi
done | sort > checksums.sha256
find ./providers -maxdepth 1 -mindepth 1 -type f -not -name 'go.*' -not -name 'README.*' -not -name 'test*' | while read -r dir; do
  echo "$(echo $dir | sed 's/^..//'): $(sha256sum $dir | awk '{print $1}')"
done | sort >> checksums.sha256
