#!/bin/bash
# retry <command...>
# Runs the command up to RETRY_ATTEMPTS times (default 3) with a doubling delay that
# starts at RETRY_DELAY seconds (default 10). Meant for registry pushes and copies,
# which fail transiently with 5xx / connection resets; every push here is
# content-addressed, so repeating a partially completed one is harmless.
# The command's own stdout is left alone so callers can still capture a digest from it.
retry() {
  local attempts="${RETRY_ATTEMPTS:-3}" delay="${RETRY_DELAY:-10}" n=1
  until "$@"; do
    if [ "${n}" -ge "${attempts}" ]; then
      echo "ERROR: giving up after ${n} attempts: $*" >&2
      return 1
    fi
    echo "WARN: attempt ${n}/${attempts} failed, retrying in ${delay}s: $*" >&2
    sleep "${delay}"
    n=$((n + 1))
    delay=$((delay * 2))
  done
}
