#!/bin/bash
SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
cd $SCRIPTPATH/../..
kubectl get vulnerabilityreports -A -o json | jq -r '
  # Group by namespace and image
  [.items[] | {
    namespace: .metadata.namespace,
    image: (.report.artifact.repository + ":" + .report.artifact.tag),
    vulnerabilities: [.report.vulnerabilities[]? | {id: .vulnerabilityID, severity: .severity}]
  }] |
  group_by(.namespace, .image) |
  map({
    namespace: .[0].namespace,
    image: .[0].image,
    vulnerabilities: [.[].vulnerabilities[]] | group_by(.severity) | map({
      severity: .[0].severity,
      count: (map(.id) | unique | length)
    }) | INDEX(.severity)
  }) |
  map(select((.vulnerabilities.CRITICAL.count // 0) > 0)) |
  sort_by(-(.vulnerabilities.CRITICAL.count // 0)) |
  ["NAMESPACE", "IMAGE", "CRITICAL", "HIGH", "MEDIUM", "LOW", "UNKNOWN"],
  (.[] | [
    .namespace,
    .image,
    (.vulnerabilities.CRITICAL.count // 0),
    (.vulnerabilities.HIGH.count // 0),
    (.vulnerabilities.MEDIUM.count // 0),
    (.vulnerabilities.LOW.count // 0),
    (.vulnerabilities.UNKNOWN.count // 0)
  ]) |
  @tsv
' | column -t


kubectl get vulnerabilityreports -A -o json | jq -r '
  # Group by namespace
  [.items[] | {
    namespace: .metadata.namespace,
    vulnerabilities: [.report.vulnerabilities[]? | {id: .vulnerabilityID, severity: .severity}]
  }] |
  group_by(.namespace) |
  map({
    namespace: .[0].namespace,
    vulnerabilities: [.[].vulnerabilities[]] | group_by(.severity) | map({
      severity: .[0].severity,
      count: (map(.id) | unique | length)
    }) | INDEX(.severity)
  }) |
  sort_by(-(.vulnerabilities.CRITICAL.count // 0)) |
  ["NAMESPACE", "CRITICAL", "HIGH", "MEDIUM", "LOW", "UNKNOWN"],
  (.[] | [
    .namespace,
    (.vulnerabilities.CRITICAL.count // 0),
    (.vulnerabilities.HIGH.count // 0),
    (.vulnerabilities.MEDIUM.count // 0),
    (.vulnerabilities.LOW.count // 0),
    (.vulnerabilities.UNKNOWN.count // 0)
  ]) |
  @tsv
' | column -t

kubectl get vulnerabilityreports -A -o json | jq -r '
  # Collect all vulnerabilities from all reports
  [.items[].report.vulnerabilities[]? | {id: .vulnerabilityID, severity: .severity}] |

  # Group by severity and get unique CVE IDs
  group_by(.severity) |
  map({
    severity: .[0].severity,
    count: (map(.id) | unique | length)
  }) |

  # Convert to object for easier access
  INDEX(.severity) |

  # Format output
  "Cluster-wide Unique CVE Statistics\n" +
  "===================================\n" +
  "Critical: \(.CRITICAL.count // 0)\n" +
  "High:     \(.HIGH.count // 0)\n" +
  "Medium:   \(.MEDIUM.count // 0)\n" +
  "Low:      \(.LOW.count // 0)\n" +
  "Unknown:  \(.UNKNOWN.count // 0)\n" +
  "-----------------------------------\n" +
  "Total:    \(((.CRITICAL.count // 0) + (.HIGH.count // 0) + (.MEDIUM.count // 0) + (.LOW.count // 0) + (.UNKNOWN.count // 0)))"
'
