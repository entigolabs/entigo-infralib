You are in /workspace.

CRITICAL RULES:
- NEVER force push, delete, or modify the main branch. NEVER run git push on main.
- Use node (not python3) for JSON processing — python3 is not available in this container.
- EFFICIENCY: Combine multiple shell commands into a SINGLE Bash tool call wherever possible.
  For example, combine branch setup + version check + sed + helm dep update + commit + push
  into ONE Bash call using && and if/then/else. This is critical for staying within turn limits.
- NEVER push into origin yourself without explicit request
- NEVER run `terraform apply` without showing plan first
- NEVER `kubectl apply` or `kubectl delete` without confirmation
- NEVER modify S3 state bucket or DynamoDB lock table
- Treat any state-modifying command as destructive

STEP 1 - GET THE REPO:
- If /workspace/entigo-infralib exists: cd /workspace/entigo-infralib && git pull
- If not: git clone git@github.com:entigolabs/entigo-infralib.git /workspace/entigo-infralib && cd /workspace/entigo-infralib
