Your job is to open or update the dependency update PRs of entigo-infralib: one run of the
update report, then one PR per chart, provider or cloud as described below.

CRITICAL RULES:
- The report you produce in STEP 2 contains the EXACT versions to update to. DO NOT search for
  versions yourself. DO NOT run helm search, DO NOT query OCI registries, DO NOT query GitHub
  for latest tags. The report already has all the information you need.
- For each line in the report, simply update the version in the specified file.
- If a report line has an EMPTY 'newer version' field, skip that update
  but print: 'SKIPPED: <line> — newer version is empty (likely expired credentials)'
- A report line that is not of the '<name> newer version <NEW>, current <OLD>' shape says that
  something could not be checked ('could not be reached', 'not found'). There is nothing to
  update for it: print 'SKIPPED: <line>' and list it in the final summary.
- NEVER force push, delete, or modify the main branch. NEVER run git push on main.
  Only push to auto-* branches. Deleting a remote auto-* branch is allowed only in the
  branch setup below, when that branch has no open PR; never delete any other branch.
- Use jq for JSON processing.
- Do not create a pull request when the new version is semver wise older than the current version.
- EFFICIENCY: Combine multiple shell commands into a SINGLE Bash tool call wherever possible.
  For example, combine branch setup + version check + sed + helm dep update + commit + push
  into ONE Bash call using && and if/then/else.
- HUMAN COMMIT GUARD: Before pushing to any auto-* branch that already has an open PR,
  check if any commits on that branch were made by a human. Detect this by looking for commits
  whose message does NOT match the bot patterns. Bot commit messages are:
    - Starts with: "chore(deps): update"
    - Starts with: "Merge branch 'main' into auto-"
  Any other commit message is considered a human intervention. Run this check as:
    PR_NUMBER=$(gh pr list --head <BRANCH> --json number --jq '.[0].number' 2>/dev/null)
    NON_BOT=$(gh api repos/entigolabs/entigo-infralib/pulls/${PR_NUMBER}/commits \
      --jq '[.[] | select(
        (.commit.message | test("^chore\\(deps\\): update")) or
        (.commit.message | test("^Merge branch .main. into auto-"))
        | not
      )] | length')
  If NON_BOT > 0:
    - Do NOT git push
    - Do NOT gh pr create or gh pr edit
    - Print: "SKIPPED: <BRANCH> — human commits detected, not overwriting manual work"
    - Move on to the next item
  This check must run for every branch that already has an open PR, before any git push
  or PR create/edit operation.

STEP 1 - GET THE REPO:
Work in a clone of its own, never in a checkout somebody develops in: this run switches
branches and commits with git add -A. The clone is ${ENTIGO_INFRALIB_DIR:-$HOME/claude}/entigo-infralib,
called <CLONE> below; write its absolute path wherever <CLONE> appears.
- If <CLONE> exists: cd <CLONE> && git checkout main && git pull origin main
- If not: gh repo clone entigolabs/entigo-infralib <CLONE> && cd <CLONE>
  Make sure commits made there are signed the way this machine signs commits elsewhere.

Tools this run needs on the machine: git, gh (logged in), curl, jq, docker (with buildx, for
the report and for pullpush.sh digests) and helm (for helm dependency update). If one is
missing, stop and say which.

STEP 2 - RUN THE REPORT:
The report is the framework's scripts/report.sh, which ./test.sh runs in the pinned test
image, so nothing but docker is needed for it. From <CLONE>, on main:
  GITHUB_TOKEN=$(gh auth token) ./test.sh report k8s aws google azure oracle 2>&1
- The EKS AMI lines of the aws kind (report/aws.sh) read SSM and need AWS credentials in the
  shell. Without them the report prints 'EKS ... recommended release version not found in SSM'
  lines: skip those as described in the rules and carry on.
- The exit status is 1 when something could not be checked; the lines that were printed are
  still valid.
- If the report prints no 'newer version' line at all, print "No updates found." and stop.

The output has one section per kind, each introduced by a '==> Updates for <kind>' line:

==> Updates for k8s (K8S HELM UPDATES)
Lines like: <chart_name> newer version <NEW>, current <OLD> (<file_path>)
Each line means: open <file_path>, find version <OLD>, replace with <NEW>.

VENDORED MODULES (applies to every TERRAFORM section below):
Community modules are vendored into the module that calls them: the version is in
modules/<cloud>/<module>/vendor.yaml, the module's files are under
modules/<cloud>/<module>/vendors/ and vendor.lock records what was vendored.
- A report line whose name has three parts (<namespace>/<name>/<provider>, e.g.
  terraform-aws-modules/eks/aws) is a module: replace <OLD> with <NEW> for that source in
  every vendor.yaml that lists it (grep -rl --include=vendor.yaml '<source>' modules/), then
  for each of those module directories run
  ./test.sh vendor modules/<cloud>/<module>
  from <CLONE>. It rewrites vendors/ and vendor.lock; commit all of it.
- NEVER edit vendor.lock or any file under a vendors/ directory by hand, and exclude them from
  every grep/sed replacement (grep -r --exclude-dir=vendors --exclude=vendor.lock).
- If vendor.sh fails, leave that module out of the PR and say so in the final summary.

==> Updates for aws (AWS TERRAFORM UPDATES)
Lines like: hashicorp/aws newer version <NEW>, current <OLD>
Use grep -r --exclude-dir=vendors --exclude=vendor.lock to find <OLD> in modules/aws/ and in modules/aws-v2 and replace with <NEW>.
EKS AL2023_* lines: find the AMI version string in modules/aws/ and replace.

==> Updates for google (GOOGLE TERRAFORM UPDATES)
Lines like: hashicorp/google newer version <NEW>, current <OLD>
Use grep -r --exclude-dir=vendors --exclude=vendor.lock to find <OLD> in modules/google/ and replace with <NEW>.

==> Updates for azure (AZURE TERRAFORM UPDATES)
Lines like: hashicorp/azurerm newer version <NEW>, current <OLD>
Use grep -r --exclude-dir=vendors --exclude=vendor.lock to find <OLD> in modules/azure/ and replace with <NEW>.

==> Updates for oracle (ORACLE TERRAFORM UPDATES)
Lines like: oracle/oci newer version <NEW>, current <OLD>
Use grep -r --exclude-dir=vendors --exclude=vendor.lock to find <OLD> in modules/oracle/ and replace with <NEW>.

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
  cd <CLONE> && \
  git checkout main && \
  git fetch --prune origin && \
  # Reuse the branch only while its PR is open. Once that PR was merged or
  # closed, the branch and its history go, so the next PR shows only its own
  # commits; a fresh branch of the same name starts from main.
  OPEN_PR=$(gh pr list --head auto-k8s-<DIR> --state open --json number --jq '.[0].number' 2>/dev/null) && \
  if [ -n "$OPEN_PR" ] && git rev-parse --verify origin/auto-k8s-<DIR> >/dev/null 2>&1; then
    git checkout -B auto-k8s-<DIR> origin/auto-k8s-<DIR>
    git merge main --no-edit
  else
    if git rev-parse --verify origin/auto-k8s-<DIR> >/dev/null 2>&1; then
      git push origin --delete auto-k8s-<DIR> && git fetch --prune origin
    fi
    git checkout -B auto-k8s-<DIR> main
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
        # Human commit guard
        PR_NUMBER=$(gh pr list --head auto-k8s-<DIR> --json number --jq '.[0].number' 2>/dev/null) && \
        if [ -n "$PR_NUMBER" ]; then
          NON_BOT=$(gh api repos/entigolabs/entigo-infralib/pulls/${PR_NUMBER}/commits \
            --jq '[.[] | select(
              (.commit.message | test("^chore\\(deps\\): update")) or
              (.commit.message | test("^Merge branch .main. into auto-"))
              | not
            )] | length') && \
          if [ "$NON_BOT" -gt 0 ]; then
            echo "SKIPPED: auto-k8s-<DIR> — human commits detected, not overwriting manual work"
            exit 0
          fi
        fi
        echo "PUSH_MERGE_ONLY"
      else
        echo "ALREADY_DONE"
      fi
    else
      echo "ALREADY_DONE"
    fi
  else
    sed -i 's/version: <OLD>/version: <NEW>/' modules/k8s/<DIR>/Chart.yaml && \
    cd modules/k8s/<DIR> && helm dependency update && cd <CLONE> && \
    # Refresh the pod image digests of a module that pins them (a tag may have moved with the chart).
    if grep -qvE '^(#|imageDigests:|[[:space:]]|$)' modules/k8s/<DIR>/values-images.yaml 2>/dev/null; then modules/k8s/pod-images.sh modules/k8s/<DIR>; fi && \
    git add -A && \
    if git diff --cached --quiet; then
      echo "NO_CHANGES"
    else
      git commit -m 'chore(deps): update <CHART> to <NEW>' && \
      if git rev-parse --verify origin/auto-k8s-<DIR> >/dev/null 2>&1; then
        LOCAL_TREE=$(git rev-parse HEAD^{tree}) && \
        REMOTE_TREE=$(git rev-parse origin/auto-k8s-<DIR>^{tree}) && \
        if [ "$LOCAL_TREE" != "$REMOTE_TREE" ]; then
          # Human commit guard
          PR_NUMBER=$(gh pr list --head auto-k8s-<DIR> --json number --jq '.[0].number' 2>/dev/null) && \
          if [ -n "$PR_NUMBER" ]; then
            NON_BOT=$(gh api repos/entigolabs/entigo-infralib/pulls/${PR_NUMBER}/commits \
              --jq '[.[] | select(
                (.commit.message | test("^chore\\(deps\\): update")) or
                (.commit.message | test("^Merge branch .main. into auto-"))
                | not
              )] | length') && \
            if [ "$NON_BOT" -gt 0 ]; then
              echo "SKIPPED: auto-k8s-<DIR> — human commits detected, not overwriting manual work"
              exit 0
            fi
          fi
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
  cd <CLONE> && \
  git checkout main && \
  git fetch --prune origin && \
  # Reuse the branch only while its PR is open. Once that PR was merged or
  # closed, the branch and its history go, so the next PR shows only its own
  # commits; a fresh branch of the same name starts from main.
  OPEN_PR=$(gh pr list --head auto-k8s-<DIR> --state open --json number --jq '.[0].number' 2>/dev/null) && \
  if [ -n "$OPEN_PR" ] && git rev-parse --verify origin/auto-k8s-<DIR> >/dev/null 2>&1; then
    git checkout -B auto-k8s-<DIR> origin/auto-k8s-<DIR>
    git merge main --no-edit
  else
    if git rev-parse --verify origin/auto-k8s-<DIR> >/dev/null 2>&1; then
      git push origin --delete auto-k8s-<DIR> && git fetch --prune origin
    fi
    git checkout -B auto-k8s-<DIR> main
  fi && \
  # Reset provider files to main's version before applying our change.
  # This prevents stale versions from persisting on long-lived branches.
  git checkout main -- modules/k8s/<DIR>/templates/provider.yaml modules/k8s/<DIR>/pullpush.sh modules/k8s/<DIR>/values-images.yaml 2>/dev/null || true && \
  # Check if the target version is already in provider.yaml — skip if so
  if grep -q '<NEW>' modules/k8s/<DIR>/templates/provider.yaml; then
    # Version already matches — check if branch has different file content than remote
    if git rev-parse --verify origin/auto-k8s-<DIR> >/dev/null 2>&1; then
      LOCAL_TREE=$(git rev-parse HEAD^{tree})
      REMOTE_TREE=$(git rev-parse origin/auto-k8s-<DIR>^{tree})
      if [ "$LOCAL_TREE" != "$REMOTE_TREE" ]; then
        # Human commit guard
        PR_NUMBER=$(gh pr list --head auto-k8s-<DIR> --json number --jq '.[0].number' 2>/dev/null) && \
        if [ -n "$PR_NUMBER" ]; then
          NON_BOT=$(gh api repos/entigolabs/entigo-infralib/pulls/${PR_NUMBER}/commits \
            --jq '[.[] | select(
              (.commit.message | test("^chore\\(deps\\): update")) or
              (.commit.message | test("^Merge branch .main. into auto-"))
              | not
            )] | length') && \
          if [ "$NON_BOT" -gt 0 ]; then
            echo "SKIPPED: auto-k8s-<DIR> — human commits detected, not overwriting manual work"
            exit 0
          fi
        fi
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
    # Pin the new version's packages by digest (resolved upstream, no login needed).
    modules/k8s/<DIR>/pullpush.sh digests && \
    git add -A && \
    if git diff --cached --quiet; then
      echo "NO_CHANGES"
    else
      git commit -m 'chore(deps): update <PROVIDER> to <NEW>' && \
      if git rev-parse --verify origin/auto-k8s-<DIR> >/dev/null 2>&1; then
        LOCAL_TREE=$(git rev-parse HEAD^{tree}) && \
        REMOTE_TREE=$(git rev-parse origin/auto-k8s-<DIR>^{tree}) && \
        if [ "$LOCAL_TREE" != "$REMOTE_TREE" ]; then
          # Human commit guard
          PR_NUMBER=$(gh pr list --head auto-k8s-<DIR> --json number --jq '.[0].number' 2>/dev/null) && \
          if [ -n "$PR_NUMBER" ]; then
            NON_BOT=$(gh api repos/entigolabs/entigo-infralib/pulls/${PR_NUMBER}/commits \
              --jq '[.[] | select(
                (.commit.message | test("^chore\\(deps\\): update")) or
                (.commit.message | test("^Merge branch .main. into auto-"))
                | not
              )] | length') && \
            if [ "$NON_BOT" -gt 0 ]; then
              echo "SKIPPED: auto-k8s-<DIR> — human commits detected, not overwriting manual work"
              exit 0
            fi
          fi
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
  - "SKIPPED: ... human commits detected" → skip to next chart, no PR needed, log the skip in final summary.
  - "PUSH_MERGE_ONLY" → run git push in one call, then check/update PR.
  - "PUSHED" → check/update PR.

  For PR creation/update (only when PUSHED or PUSH_MERGE_ONLY), run in ONE Bash call:
  ```
  cd <CLONE> && \
  PR_JSON=$(gh pr list --head auto-k8s-<DIR> --json number,title,url) && \
  RELEASE_NOTES=$(curl -s https://api.github.com/repos/<ORG>/<REPO>/releases/latest | jq -r '"TAG:\(.tag_name)\nBODY:\(.body[0:1500])"' 2>/dev/null || echo UNAVAILABLE) && \
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
     - Do NOT spend more than FIVE curl calls per chart on release notes.
     - Try to find the application version from the chart version and provide the application release notes too.

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
  cd <CLONE> && \
  git checkout main && \
  git fetch --prune origin && \
  # Reuse the branch only while its PR is open. Once that PR was merged or
  # closed, the branch and its history go, so the next PR shows only its own
  # commits; a fresh branch of the same name starts from main.
  OPEN_PR=$(gh pr list --head auto-aws-<TYPE> --state open --json number --jq '.[0].number' 2>/dev/null) && \
  if [ -n "$OPEN_PR" ] && git rev-parse --verify origin/auto-aws-<TYPE> >/dev/null 2>&1; then
    git checkout -B auto-aws-<TYPE> origin/auto-aws-<TYPE>
    git rebase main
  else
    if git rev-parse --verify origin/auto-aws-<TYPE> >/dev/null 2>&1; then
      git push origin --delete auto-aws-<TYPE> && git fetch --prune origin
    fi
    git checkout -B auto-aws-<TYPE> main
  fi && \
  # For 'provider' type: apply ALL provider/module version changes
  # For 'ami' type: update ami_release_version blocks in eks/main.tf and eks-node-group/main.tf
  <APPLY SED COMMANDS HERE> && \
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
        # Human commit guard
        PR_NUMBER=$(gh pr list --head auto-aws-<TYPE> --json number --jq '.[0].number' 2>/dev/null) && \
        if [ -n "$PR_NUMBER" ]; then
          NON_BOT=$(gh api repos/entigolabs/entigo-infralib/pulls/${PR_NUMBER}/commits \
            --jq '[.[] | select(
              (.commit.message | test("^chore\\(deps\\): update")) or
              (.commit.message | test("^Merge branch .main. into auto-"))
              | not
            )] | length') && \
          if [ "$NON_BOT" -gt 0 ]; then
            echo "SKIPPED: auto-aws-<TYPE> — human commits detected, not overwriting manual work"
            exit 0
          fi
        fi
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
  - "SKIPPED: ... human commits detected" → skip to next type, no PR needed, log the skip in final summary.
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
  cd <CLONE> && \
  git checkout main && \
  git fetch --prune origin && \
  # Reuse the branch only while its PR is open. Once that PR was merged or
  # closed, the branch and its history go, so the next PR shows only its own
  # commits; a fresh branch of the same name starts from main.
  OPEN_PR=$(gh pr list --head auto-gcptf --state open --json number --jq '.[0].number' 2>/dev/null) && \
  if [ -n "$OPEN_PR" ] && git rev-parse --verify origin/auto-gcptf >/dev/null 2>&1; then
    git checkout -B auto-gcptf origin/auto-gcptf
    git rebase main
  else
    if git rev-parse --verify origin/auto-gcptf >/dev/null 2>&1; then
      git push origin --delete auto-gcptf && git fetch --prune origin
    fi
    git checkout -B auto-gcptf main
  fi && \
  # Apply ALL Google version changes
  <APPLY SED COMMANDS HERE> && \
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
        # Human commit guard
        PR_NUMBER=$(gh pr list --head auto-gcptf --json number --jq '.[0].number' 2>/dev/null) && \
        if [ -n "$PR_NUMBER" ]; then
          NON_BOT=$(gh api repos/entigolabs/entigo-infralib/pulls/${PR_NUMBER}/commits \
            --jq '[.[] | select(
              (.commit.message | test("^chore\\(deps\\): update")) or
              (.commit.message | test("^Merge branch .main. into auto-"))
              | not
            )] | length') && \
          if [ "$NON_BOT" -gt 0 ]; then
            echo "SKIPPED: auto-gcptf — human commits detected, not overwriting manual work"
            exit 0
          fi
        fi
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
  - "SKIPPED: ... human commits detected" → skip to AFTER ALL PRs ARE DONE, log the skip in final summary.
  - "PUSHED" → check/update PR.

  PR title: 'chore(deps): update Google Terraform providers'
  PR body MUST include for each updated component:
     - Version change table
     - Link to GitHub releases:
       hashicorp/google → https://github.com/hashicorp/terraform-provider-google/releases
       hashicorp/google-beta → https://github.com/hashicorp/terraform-provider-google-beta/releases
       terraform-google-modules/kubernetes-engine/google → https://github.com/terraform-google-modules/terraform-google-kubernetes-engine/releases
     - Fetch release notes with curl (five calls per provider max)
     - Extract: breaking changes (⚠️), key changes summary, link to full changelog
     - If it fails, note 'Release notes unavailable'.

FOR AZURE TERRAFORM UPDATES - ONE COMBINED PR:
  There is only one type of update for AZURE TERRAFORM UPDATES.
  Run the ENTIRE branch+update+commit+push sequence in a SINGLE Bash call:
  ```
  cd <CLONE> && \
  git checkout main && \
  git fetch --prune origin && \
  # Reuse the branch only while its PR is open. Once that PR was merged or
  # closed, the branch and its history go, so the next PR shows only its own
  # commits; a fresh branch of the same name starts from main.
  OPEN_PR=$(gh pr list --head auto-azuretf --state open --json number --jq '.[0].number' 2>/dev/null) && \
  if [ -n "$OPEN_PR" ] && git rev-parse --verify origin/auto-azuretf >/dev/null 2>&1; then
    git checkout -B auto-azuretf origin/auto-azuretf
    git rebase main
  else
    if git rev-parse --verify origin/auto-azuretf >/dev/null 2>&1; then
      git push origin --delete auto-azuretf && git fetch --prune origin
    fi
    git checkout -B auto-azuretf main
  fi && \
  # Apply ALL Azure version changes
  <APPLY SED COMMANDS HERE> && \
  git add -A && \
  if git diff --cached --quiet; then
    echo "NO_CHANGES"
  else
    echo "BRANCH: $(git branch --show-current)" && \
    git commit -m 'chore(deps): update Azure Terraform providers' && \
    if git rev-parse --verify origin/auto-azuretf >/dev/null 2>&1; then
      # Compare actual file tree content, not commit history.
      # After rebase, commits differ but content may be identical.
      LOCAL_TREE=$(git rev-parse HEAD^{tree}) && \
      REMOTE_TREE=$(git rev-parse origin/auto-azuretf^{tree}) && \
      if [ "$LOCAL_TREE" != "$REMOTE_TREE" ]; then
        # Human commit guard
        PR_NUMBER=$(gh pr list --head auto-azuretf --json number --jq '.[0].number' 2>/dev/null) && \
        if [ -n "$PR_NUMBER" ]; then
          NON_BOT=$(gh api repos/entigolabs/entigo-infralib/pulls/${PR_NUMBER}/commits \
            --jq '[.[] | select(
              (.commit.message | test("^chore\\(deps\\): update")) or
              (.commit.message | test("^Merge branch .main. into auto-"))
              | not
            )] | length') && \
          if [ "$NON_BOT" -gt 0 ]; then
            echo "SKIPPED: auto-azuretf — human commits detected, not overwriting manual work"
            exit 0
          fi
        fi
        git push --force origin auto-azuretf && echo "PUSHED"
      else
        echo "SKIP_PUSH_IDENTICAL"
      fi
    else
      git push origin auto-azuretf && echo "PUSHED"
    fi
  fi
  ```

  After the Bash call, check the output:
  - "NO_CHANGES" or "SKIP_PUSH_IDENTICAL" → skip to AFTER ALL PRs ARE DONE.
  - "SKIPPED: ... human commits detected" → skip to AFTER ALL PRs ARE DONE, log the skip in final summary.
  - "PUSHED" → check/update PR.

  PR title: 'chore(deps): update Azure Terraform providers'
  PR body MUST include for each updated component:
     - Version change table
     - Link to GitHub releases:
       hashicorp/azurerm → https://github.com/hashicorp/terraform-provider-azurerm/releases
       Azure/avm-res-containerservice-managedcluster/azurerm → https://github.com/Azure/terraform-azurerm-avm-res-containerservice-managedcluster/releases
     - Fetch release notes with curl (five calls per provider max)
     - Extract: breaking changes (⚠️), key changes summary, link to full changelog
     - If it fails, note 'Release notes unavailable'.

FOR ORACLE TERRAFORM UPDATES - ONE COMBINED PR:
  There is only one type of update for ORACLE TERRAFORM UPDATES.
  Run the ENTIRE branch+update+commit+push sequence in a SINGLE Bash call:
  ```
  cd <CLONE> && \
  git checkout main && \
  git fetch --prune origin && \
  # Reuse the branch only while its PR is open. Once that PR was merged or
  # closed, the branch and its history go, so the next PR shows only its own
  # commits; a fresh branch of the same name starts from main.
  OPEN_PR=$(gh pr list --head auto-oracletf --state open --json number --jq '.[0].number' 2>/dev/null) && \
  if [ -n "$OPEN_PR" ] && git rev-parse --verify origin/auto-oracletf >/dev/null 2>&1; then
    git checkout -B auto-oracletf origin/auto-oracletf
    git rebase main
  else
    if git rev-parse --verify origin/auto-oracletf >/dev/null 2>&1; then
      git push origin --delete auto-oracletf && git fetch --prune origin
    fi
    git checkout -B auto-oracletf main
  fi && \
  # Apply ALL Oracle version changes
  <APPLY SED COMMANDS HERE> && \
  git add -A && \
  if git diff --cached --quiet; then
    echo "NO_CHANGES"
  else
    echo "BRANCH: $(git branch --show-current)" && \
    git commit -m 'chore(deps): update Oracle Terraform providers' && \
    if git rev-parse --verify origin/auto-oracletf >/dev/null 2>&1; then
      # Compare actual file tree content, not commit history.
      # After rebase, commits differ but content may be identical.
      LOCAL_TREE=$(git rev-parse HEAD^{tree}) && \
      REMOTE_TREE=$(git rev-parse origin/auto-oracletf^{tree}) && \
      if [ "$LOCAL_TREE" != "$REMOTE_TREE" ]; then
        # Human commit guard
        PR_NUMBER=$(gh pr list --head auto-oracletf --json number --jq '.[0].number' 2>/dev/null) && \
        if [ -n "$PR_NUMBER" ]; then
          NON_BOT=$(gh api repos/entigolabs/entigo-infralib/pulls/${PR_NUMBER}/commits \
            --jq '[.[] | select(
              (.commit.message | test("^chore\\(deps\\): update")) or
              (.commit.message | test("^Merge branch .main. into auto-"))
              | not
            )] | length') && \
          if [ "$NON_BOT" -gt 0 ]; then
            echo "SKIPPED: auto-oracletf — human commits detected, not overwriting manual work"
            exit 0
          fi
        fi
        git push --force origin auto-oracletf && echo "PUSHED"
      else
        echo "SKIP_PUSH_IDENTICAL"
      fi
    else
      git push origin auto-oracletf && echo "PUSHED"
    fi
  fi
  ```

  After the Bash call, check the output:
  - "NO_CHANGES" or "SKIP_PUSH_IDENTICAL" → skip to AFTER ALL PRs ARE DONE.
  - "SKIPPED: ... human commits detected" → skip to AFTER ALL PRs ARE DONE, log the skip in final summary.
  - "PUSHED" → check/update PR.

  PR title: 'chore(deps): update Oracle Terraform providers'
  PR body MUST include for each updated component:
     - Version change table
     - Link to GitHub releases:
       oracle/oci → https://github.com/oracle/terraform-provider-oci/releases
       hashicorp/random → https://github.com/hashicorp/terraform-provider-random/releases
       hashicorp/time → https://github.com/hashicorp/terraform-provider-time/releases
     - Fetch release notes with curl (five calls per provider max)
     - Extract: breaking changes (⚠️), key changes summary, link to full changelog
     - If it fails, note 'Release notes unavailable'.

AFTER ALL PRs ARE DONE:
  git checkout main
  Print a summary of all actions taken:
    - Which PRs were created or updated (with PR numbers)
    - Which branches were skipped (already up to date)
    - Which branches were skipped due to human commits (list branch name and PR number)
    - Which report lines were skipped (empty newer version, or something that could not be checked)
