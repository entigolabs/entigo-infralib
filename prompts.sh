#!/bin/bash
# prompts.sh
# Runs Claude Code inside a Docker container as UID 1000 for full isolation.
# Prompts are read from .md files in the prompts/ directory next to this script.
#
# Usage:
#   ./prompts.sh update           # Run dependency updates (normal logging)
#   ./prompts.sh -v update        # Run with verbose/debug logging
#   ./prompts.sh fix              # Fix failed CI on auto-* PRs
#   ./prompts.sh login            # Re-authenticate only
#   ./prompts.sh <name>           # Run prompts/<name>.md
#


set -euo pipefail

# Parse flags
VERBOSE=false
while getopts "v" opt; do
  case $opt in
    v) VERBOSE=true ;;
    *) echo "Usage: $0 [-v] <prompt_name|login>" >&2; exit 1 ;;
  esac
done
shift $((OPTIND - 1))

WORK_DIR="${ENTIGO_INFRALIB_DIR:-$HOME/claude}"
IMAGE_NAME="claude-code"
REPO_URL="git@github.com:entigolabs/entigo-infralib.git"
CLAUDE_CONFIG_VOL="claude-code-config"
CLAUDE_DATA_VOL="claude-code-data"
DOCKER="sudo docker"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# Ensure workspace exists
mkdir -p "$WORK_DIR"

# Ensure claude config file exists on host for bind mount
CLAUDE_JSON="${HOME}/.claude-docker.json"
[ -f "$CLAUDE_JSON" ] || echo '{}' > "$CLAUDE_JSON"

# Build the image if needed
$DOCKER build -t "$IMAGE_NAME" -f "${SCRIPT_DIR}/images/claude/Dockerfile.claude" "$SCRIPT_DIR"

# Container home dir matches the user inside the container
C_HOME="/home/node"

# Generate a container-specific gitconfig (fixes path mismatches)
CONTAINER_GITCONFIG="${WORK_DIR}/.gitconfig-container"
cat > "$CONTAINER_GITCONFIG" <<EOF
[user]
	email = $(git config --global user.email)
	name = $(git config --global user.name)
	signingkey = ${C_HOME}/.ssh/id_ed25519.pub
[gpg]
	format = ssh
[commit]
	gpgsign = true
[pull]
	rebase = true
[rebase]
	autoStash = true
[safe]
	directory = /workspace/entigo-infralib
[branch "main"]
	pushRemote = no_push
[branch "master"]
	pushRemote = no_push
EOF

# GitHub token for gh CLI (create at https://github.com/settings/tokens with repo scope)
GH_TOKEN_FILE="${HOME}/.gh-token-claude"

# Function to run claude in the container as UID 1000
run_claude() {
  local gh_token_arg=""
  if [ -f "$GH_TOKEN_FILE" ]; then
    gh_token_arg="-e GH_TOKEN=$(cat "$GH_TOKEN_FILE")"
  fi

  # Use -it (interactive+TTY) for login/verbose, -i only when piping output
  local it_flag="-it"
  if [ "${PIPE_MODE:-false}" = true ]; then
    it_flag="-i"
  fi

  $DOCKER run --rm $it_flag \
    --user "$(id -u):$(id -g)" \
    -v "${WORK_DIR}:/workspace" \
    -v "${CLAUDE_CONFIG_VOL}:${C_HOME}/.claude" \
    -v "${CLAUDE_DATA_VOL}:${C_HOME}/.local/share/claude" \
    -v "${CLAUDE_JSON}:${C_HOME}/.claude.json" \
    -v "${HOME}/.config/gh:${C_HOME}/.config/gh:ro" \
    -v "${HOME}/.ssh:${C_HOME}/.ssh:ro" \
    -v "${CONTAINER_GITCONFIG}:${C_HOME}/.gitconfig:ro" \
    -e "HOME=${C_HOME}" \
    -e "GIT_SSH_COMMAND=ssh -i ${C_HOME}/.ssh/id_ed25519 -o IdentitiesOnly=yes" \
    $gh_token_arg \
    -w /workspace \
    "${IMAGE_NAME}" "$@"
}

# --- Login mode ---
if [ "${1:-}" = "login" ]; then
  echo "Starting interactive login..."
  run_claude
  exit 0
fi

# --- Require a prompt name ---
PROMPT_NAME="${1:-}"
if [ -z "$PROMPT_NAME" ]; then
  echo "Usage: $0 [-v] <prompt_name|login>"
  echo "Available prompts:"
  ls "${SCRIPT_DIR}/prompts/"*.md 2>/dev/null | xargs -I{} basename {} .md | sed 's/^/  /'
  exit 1
fi

PROMPT_FILE="${SCRIPT_DIR}/prompts/${PROMPT_NAME}.md"
if [ ! -f "$PROMPT_FILE" ]; then
  echo "ERROR: Prompt file not found: $PROMPT_FILE"
  echo "Available prompts:"
  ls "${SCRIPT_DIR}/prompts/"*.md 2>/dev/null | xargs -I{} basename {} .md | sed 's/^/  /'
  exit 1
fi

# --- Pre-steps for 'update' prompt: generate the report ---
if [ "$PROMPT_NAME" = "update" ]; then
  # Ensure repo is cloned
  if [ ! -d "$WORK_DIR/entigo-infralib" ]; then
    echo "Cloning repo..."
    git clone "$REPO_URL" "$WORK_DIR/entigo-infralib"
  fi

  # Update to latest main
  cd "$WORK_DIR/entigo-infralib"
  git checkout main
  git pull origin main

  # Run report scripts on the host
  LF=$'\n'
  REPORT=""

  if [ -x ./common/k8s/report.sh ]; then
    helm repo update
    K8S=$(./common/k8s/report.sh 2>&1) || true
    if [ -n "$K8S" ]; then
      REPORT+="=== K8S HELM UPDATES ===${LF}${K8S}${LF}${LF}"
    fi
  fi

  if [ -x ./common/aws/report.sh ]; then
    AWS=$(./common/aws/report.sh 2>&1) || true
    if [ -n "$AWS" ]; then
      REPORT+="=== AWS TERRAFORM UPDATES ===${LF}${AWS}${LF}${LF}"
    fi
  fi

  if [ -x ./common/google/report.sh ]; then
    GCP=$(./common/google/report.sh 2>&1) || true
    if [ -n "$GCP" ]; then
      REPORT+="=== GOOGLE TERRAFORM UPDATES ===${LF}${GCP}${LF}${LF}"
    fi
  fi

  cd - > /dev/null

  if [ -z "$REPORT" ]; then
    echo "$(date): No updates found."
    exit 0
  fi

  echo "$(date): Updates found:"
  echo "$REPORT"
  echo "$REPORT" > "$WORK_DIR/update-report.txt"
fi

# --- Run Claude Code with the selected prompt ---
echo "$(date): Running prompt '${PROMPT_NAME}'..."

CLAUDE_ARGS=(
  --max-turns 160
  --allowedTools "Read,Write,Edit,Bash,WebFetch"
  --verbose
  --output-format stream-json
)

if [ "$VERBOSE" = true ]; then
  # Full debug output: raw stream-json to stdout
  run_claude "${CLAUDE_ARGS[@]}" -p "$(cat "$PROMPT_FILE")"
else
  # Normal mode: stream-json filtered through jq for readable progress
  PIPE_MODE=true run_claude "${CLAUDE_ARGS[@]}" -p "$(cat "$PROMPT_FILE")" 2>&1 \
    | grep --line-buffered '^\{' \
    | jq -r --unbuffered '
        if .type == "assistant" and .message.content then
          .message.content[] |
            if .type == "text" and .text then
              .text
            elif .type == "tool_use" then
              "  → \(.name): \(.input.description // (.input.command // "" | split("\n")[0] | .[0:120]))"
            else empty end
        elif .type == "result" then
          "\nDone: \(.subtype // "success") | turns: \(.num_turns // "?") | cost: $\(.total_cost_usd // "?" | tostring | .[0:6]) | duration: \((.duration_ms // 0) / 1000 | floor)s"
        else empty end
      '
fi

echo "$(date): Done."
