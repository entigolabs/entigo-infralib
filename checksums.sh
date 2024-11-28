#!/bin/bash
find ./modules -maxdepth 2 -mindepth 2 -type d -not -path '*/.*' | while read -r dir; do
  if [ "$dir" != "." ]; then
      echo "$(echo $dir | sed 's/^..//'): $(find "$dir" -type f -exec sha256sum {} + | awk '{print $1}' | sort | sha256sum | awk '{print $1}')"
  fi
done > checksums.sha256
