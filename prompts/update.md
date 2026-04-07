You are in /workspace.

CRITICAL RULES:
- The update-report.txt file is ALREADY WRITTEN for you at /workspace/update-report.txt.
  It contains the EXACT versions to update to. DO NOT search for versions yourself.
  DO NOT run helm search, DO NOT query OCI registries, DO NOT query GitHub for latest tags.
  The report already has all the information you need.
- For each line in the report, simply update the version in the specified file.
- If a report line has an EMPTY 'newer version' field, skip that update
  but print: 'SKIPPED: <line> — newer version is empty (likely expired credentials)'
- NEVER force push, delete, or modify the main branch. NEVER run git push on main.
  Only push to auto-* branches.
- Use node (not python3) for JSON processing — python3 is not available in this container.
- Do not create a pull request when the new version is semver wise older than the current version.
- EFFICIENCY: Combine multiple shell commands into a SINGLE Bash tool call wherever possible.
  For example, combine branch setup + version check + sed + helm dep update + commit + push
  into ONE Bash call using && and if/then/else. This is critical for staying within turn limits.

STEP 1 - GET THE REPO:
- If /workspace/entigo-infralib exists: cd /workspace/entigo-infralib && git checkout main && git pull origin main
- If not: git clone git@github.com:entigolabs/entigo-infralib.git /workspace/entigo-infralib && cd /workspace/entigo-infralib

STEP 2 - READ THE REPORT:
Read /workspace/update-report.txt. It has three sections:

=== K8S HELM UPDATES ===
Lines like: <chart_name> newer version <NEW>, current <OLD> (<file_path>)
Each line means: open <file_path>, find version <OLD>, replace with <NEW>.

=== AWS TERRAFORM UPDATES ===
Lines like: hashicorp/aws newer version <NEW>, current <OLD>
Use grep -r to find <OLD> in modules/aws/ and replace with <NEW>.
EKS AL2023_* lines: find the AMI version string in modules/aws/ and replace.

=== GOOGLE TERRAFORM UPDATES ===
Lines like: hashicorp/google newer version <NEW>, current <OLD>
Use grep -r to find <OLD> in modules/google/ and replace with <NEW>.

STEP 3 - CREATE SEPARATE PRs:

FOR K8S HELM UPDATES - ONE PR PER CHART or PROVIDER:
  Two types of updates, first is Helm charts that have the file name Chart.yaml and the other is Crossplane Provider that has provider.yaml
  Extract the chart or provider directory name from the file_path. For example:
  modules/k8s/aws-alb/Chart.yaml → directory name is aws-alb
  modules/k8s/crossplane-google/templates/provider.yaml → directory name is crossplane-google
  Skip the platform-apis folder update in the K8S HELM UPDATES.

  For EACH chart or provider, run the ENTIRE branch+update+commit+push sequence in a SINGLE Bash call:

  For Helm charts (Chart.yaml), run this as ONE Bash command:
  ```
  cd /workspace/entigo-infralib && \
  git checkout main && \
  git fetch origin auto-k8s-<DIR> 2>/dev/null || true && \
  if git rev-parse --verify origin/auto-k8s-<DIR> >/dev/null 2>&1; then
    git checkout -B auto-k8s-<DIR> origin/auto-k8s-<DIR>
    git merge main --no-edit
  else
    git checkout -b auto-k8s-<DIR>
  fi && \
  # Reset Chart.yaml and Chart.lock to main's version before applying our change.
  # This prevents stale sub-dependency versions from persisting on long-lived branches.
  git checkout main -- modules/k8s/<DIR>/Chart.yaml modules/k8s/<DIR>/Chart.lock 2>/dev/null || true && \
  # Check if the target version is already in the file — skip everything if so
  if grep -q 'version: <NEW>' modules/k8s/<DIR>/Chart.yaml; then
    # Version already matches — check if branch has different file content than remote
    if git rev-parse --verify origin/auto-k8s-<DIR> >/dev/null 2>&1; then
      LOCAL_TREE=$(git rev-parse HEAD^{tree})
      REMOTE_TREE=$(git rev-parse origin/auto-k8s-<DIR>^{tree})
      if [ "$LOCAL_TREE" != "$REMOTE_TREE" ]; then
        echo "PUSH_MERGE_ONLY"
      else
        echo "ALREADY_DONE"
      fi
    else
      echo "ALREADY_DONE"
    fi
  else
    sed -i 's/version: <OLD>/version: <NEW>/' modules/k8s/<DIR>/Chart.yaml && \
    cd modules/k8s/<DIR> && helm dependency update && cd /workspace/entigo-infralib && \
    git add -A && \
    if git diff --cached --quiet; then
      echo "NO_CHANGES"
    else
      git commit -m 'chore(deps): update <CHART> to <NEW>' && \
      if git rev-parse --verify origin/auto-k8s-<DIR> >/dev/null 2>&1; then
        LOCAL_TREE=$(git rev-parse HEAD^{tree}) && \
        REMOTE_TREE=$(git rev-parse origin/auto-k8s-<DIR>^{tree}) && \
        if [ "$LOCAL_TREE" != "$REMOTE_TREE" ]; then
          git push origin auto-k8s-<DIR> && echo "PUSHED"
        else
          echo "SKIP_PUSH_IDENTICAL"
        fi
      else
        git push origin auto-k8s-<DIR> && echo "PUSHED"
      fi
    fi
  fi
  ```

  For Crossplane Providers (provider.yaml), run this as ONE Bash command:
  ```
  cd /workspace/entigo-infralib && \
  git checkout main && \
  git fetch origin auto-k8s-<DIR> 2>/dev/null || true && \
  if git rev-parse --verify origin/auto-k8s-<DIR> >/dev/null 2>&1; then
    git checkout -B auto-k8s-<DIR> origin/auto-k8s-<DIR>
    git merge main --no-edit
  else
    git checkout -b auto-k8s-<DIR>
  fi && \
  # Reset provider files to main's version before applying our change.
  # This prevents stale versions from persisting on long-lived branches.
  git checkout main -- modules/k8s/<DIR>/templates/provider.yaml modules/k8s/<DIR>/pullpush.sh 2>/dev/null || true && \
  # Check if the target version is already in provider.yaml — skip if so
  if grep -q '<NEW>' modules/k8s/<DIR>/templates/provider.yaml; then
    # Version already matches — check if branch has different file content than remote
    if git rev-parse --verify origin/auto-k8s-<DIR> >/dev/null 2>&1; then
      LOCAL_TREE=$(git rev-parse HEAD^{tree})
      REMOTE_TREE=$(git rev-parse origin/auto-k8s-<DIR>^{tree})
      if [ "$LOCAL_TREE" != "$REMOTE_TREE" ]; then
        echo "PUSH_MERGE_ONLY"
      else
        echo "ALREADY_DONE"
      fi
    else
      echo "ALREADY_DONE"
    fi
  else
    sed -i 's/<OLD>/<NEW>/g' modules/k8s/<DIR>/templates/provider.yaml && \
    sed -i 's/VERSION="<OLD>"/VERSION="<NEW>"/' modules/k8s/<DIR>/pullpush.sh && \
    git add -A && \
    if git diff --cached --quiet; then
      echo "NO_CHANGES"
    else
      git commit -m 'chore(deps): update <PROVIDER> to <NEW>' && \
      if git rev-parse --verify origin/auto-k8s-<DIR> >/dev/null 2>&1; then
        LOCAL_TREE=$(git rev-parse HEAD^{tree}) && \
        REMOTE_TREE=$(git rev-parse origin/auto-k8s-<DIR>^{tree}) && \
        if [ "$LOCAL_TREE" != "$REMOTE_TREE" ]; then
          git push origin auto-k8s-<DIR> && echo "PUSHED"
        else
          echo "SKIP_PUSH_IDENTICAL"
        fi
      else
        git push origin auto-k8s-<DIR> && echo "PUSHED"
      fi
    fi
  fi
  ```

  After the Bash call, check the output:
  - "ALREADY_DONE" or "NO_CHANGES" or "SKIP_PUSH_IDENTICAL" → skip to next chart, no PR needed.
  - "PUSH_MERGE_ONLY" → run git push in one call, then check/update PR.
  - "PUSHED" → check/update PR.

  For PR creation/update (only when PUSHED or PUSH_MERGE_ONLY), run in ONE Bash call:
  ```
  cd /workspace/entigo-infralib && \
  PR_JSON=$(gh pr list --head auto-k8s-<DIR> --json number,title,url) && \
  RELEASE_NOTES=$(curl -s https://api.github.com/repos/<ORG>/<REPO>/releases/latest | node -e "
    const d=require('fs').readFileSync('/dev/stdin','utf8');
    try { const r=JSON.parse(d); console.log('TAG:'+r.tag_name); console.log('BODY:'+r.body.substring(0,1500)); }
    catch(e) { console.log('UNAVAILABLE'); }
  ") && \
  echo "PR_JSON: $PR_JSON" && \
  echo "RELEASE: $RELEASE_NOTES"
  ```
  Then in ONE more Bash call, create or edit the PR based on the output (gh pr create or gh pr edit).

  PR title: 'chore(deps): update <chart_name> <old> → <new>'
  PR body MUST include:
     - Version change and file changed
     - A link to the project's GitHub releases page
     - From the release notes, extract and include in the PR body:
       * Breaking changes (if any, highlighted with ⚠️)
       * Summary of key changes (2-3 sentences max)
       * Link to full changelog
     - If release notes fetch fails, just note 'Release notes unavailable' and include the link.
     - Do NOT spend more than TWO curl calls per chart on release notes.

  Common repo mappings for release notes:
    aws-load-balancer-controller → kubernetes-sigs/aws-load-balancer-controller
    crossplane → crossplane/crossplane
    karpenter → aws/karpenter-provider-aws
    mimir-distributed → grafana/mimir
    promtail → grafana/loki (promtail is part of loki)
    wireguard → bryopsida/wireguard-chart
    upbound/provider-family-aws → crossplane-contrib/provider-upjet-aws
    upbound/provider-family-gcp → crossplane-contrib/provider-upjet-gcp
    crossplane-contrib/provider-kafka → crossplane-contrib/provider-kafka
    If the exact repo is unknown, try the obvious GitHub org/repo and if it fails, skip.

FOR AWS TERRAFORM UPDATES - SEPARATE PRs PER TYPE:
  There are two types of updates, first is named 'ami' whose Report lines start with EKS AL2023_*.
  and the second for provider and module versions named 'provider'.

  For EACH type, run the ENTIRE branch+update+commit+push sequence in a SINGLE Bash call:
  ```
  cd /workspace/entigo-infralib && \
  git checkout main && \
  git fetch origin auto-aws-<TYPE> 2>/dev/null || true && \
  if git rev-parse --verify origin/auto-aws-<TYPE> >/dev/null 2>&1; then
    git checkout -B auto-aws-<TYPE> origin/auto-aws-<TYPE>
    git rebase main
  else
    git checkout -b auto-aws-<TYPE>
  fi && \
  # For 'provider' type: apply ALL provider/module version changes
  # For 'ami' type: update ami_release_version blocks in eks/main.tf and eks-node-group/main.tf
  <APPLY SED COMMANDS HERE> && \
  # Also update images/test/cache/main.tf for provider type
  git add -A && \
  if git diff --cached --quiet; then
    echo "NO_CHANGES"
  else
    echo "BRANCH: $(git branch --show-current)" && \
    git commit -m 'chore(deps): update AWS <TYPE>' && \
    if git rev-parse --verify origin/auto-aws-<TYPE> >/dev/null 2>&1; then
      # Compare actual file tree content, not commit history.
      # After rebase, commits differ but content may be identical.
      LOCAL_TREE=$(git rev-parse HEAD^{tree}) && \
      REMOTE_TREE=$(git rev-parse origin/auto-aws-<TYPE>^{tree}) && \
      if [ "$LOCAL_TREE" != "$REMOTE_TREE" ]; then
        git push --force origin auto-aws-<TYPE> && echo "PUSHED"
      else
        echo "SKIP_PUSH_IDENTICAL"
      fi
    else
      git push origin auto-aws-<TYPE> && echo "PUSHED"
    fi
  fi
  ```

  After the Bash call, check the output:
  - "NO_CHANGES" or "SKIP_PUSH_IDENTICAL" → skip to next type, no PR needed.
  - "PUSHED" → check/update PR (same pattern as K8S: one Bash call for gh pr list + release notes, one for gh pr create/edit).

  PR title: 'chore(deps): update AWS Terraform <type>'
  PR body MUST include for each updated component:
     - Version change table
     - Link to GitHub releases:
       hashicorp/aws → https://github.com/hashicorp/terraform-provider-aws/releases
       EKS AMIs → https://github.com/awslabs/amazon-eks-ami/releases
     - Fetch release notes with curl from:
       https://api.github.com/repos/hashicorp/terraform-provider-aws/releases/latest
     - Extract: breaking changes (⚠️), key changes summary, link to full changelog
     - One curl call per provider max. If it fails, note 'Release notes unavailable'.

FOR GOOGLE TERRAFORM UPDATES - ONE COMBINED PR:
  There is only one type of update for GOOGLE TERRAFORM UPDATES.
  Run the ENTIRE branch+update+commit+push sequence in a SINGLE Bash call:
  ```
  cd /workspace/entigo-infralib && \
  git checkout main && \
  git fetch origin auto-gcptf 2>/dev/null || true && \
  if git rev-parse --verify origin/auto-gcptf >/dev/null 2>&1; then
    git checkout -B auto-gcptf origin/auto-gcptf
    git rebase main
  else
    git checkout -b auto-gcptf
  fi && \
  # Apply ALL Google version changes
  <APPLY SED COMMANDS HERE> && \
  # Also update images/test/cache/main.tf
  git add -A && \
  if git diff --cached --quiet; then
    echo "NO_CHANGES"
  else
    echo "BRANCH: $(git branch --show-current)" && \
    git commit -m 'chore(deps): update Google Terraform providers' && \
    if git rev-parse --verify origin/auto-gcptf >/dev/null 2>&1; then
      # Compare actual file tree content, not commit history.
      # After rebase, commits differ but content may be identical.
      LOCAL_TREE=$(git rev-parse HEAD^{tree}) && \
      REMOTE_TREE=$(git rev-parse origin/auto-gcptf^{tree}) && \
      if [ "$LOCAL_TREE" != "$REMOTE_TREE" ]; then
        git push --force origin auto-gcptf && echo "PUSHED"
      else
        echo "SKIP_PUSH_IDENTICAL"
      fi
    else
      git push origin auto-gcptf && echo "PUSHED"
    fi
  fi
  ```

  After the Bash call, check the output:
  - "NO_CHANGES" or "SKIP_PUSH_IDENTICAL" → skip to AFTER ALL PRs ARE DONE.
  - "PUSHED" → check/update PR.

  PR title: 'chore(deps): update Google Terraform providers'
  PR body MUST include for each updated component:
     - Version change table
     - Link to GitHub releases:
       hashicorp/google → https://github.com/hashicorp/terraform-provider-google/releases
       hashicorp/google-beta → https://github.com/hashicorp/terraform-provider-google-beta/releases
       terraform-google-modules/kubernetes-engine/google → https://github.com/terraform-google-modules/terraform-google-kubernetes-engine/releases
     - Fetch release notes with curl (one call per provider max)
     - Extract: breaking changes (⚠️), key changes summary, link to full changelog
     - If it fails, note 'Release notes unavailable'.

AFTER ALL PRs ARE DONE:
  git checkout main
  Print a summary of all actions taken:
    - Which PRs were created or updated (with PR numbers)
    - Which branches were skipped (already up to date)
    - Which report lines were skipped (empty newer version)
  Delete /workspace/update-report.txt
