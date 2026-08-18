<h1 align="center">Infralib Modules</h1>

<p align="center">
  <strong>Production-tested Terraform modules and Helm charts for building a complete Kubernetes platform on AWS or Google Cloud.</strong>
</p>

<p align="center">
  <a href="LICENSE.txt"><img src="https://img.shields.io/github/license/entigolabs/entigo-infralib" alt="License"></a>
  <a href="https://www.entigo.com/infralib"><img src="https://img.shields.io/badge/website-entigo.com%2Finfralib-blue" alt="Website"></a>
  <a href="https://github.com/entigolabs/entigo-infralib-agent"><img src="https://img.shields.io/badge/run%20with-Infralib%20Agent-1f6feb" alt="Infralib Agent"></a>
</p>

<p align="center">
  <a href="https://docs.entigo.com">Documentation</a> ·
  <a href="https://www.entigo.com/infralib">Website</a> ·
  <a href="https://github.com/entigolabs/entigo-infralib-agent">Agent</a> ·
  <a href="https://github.com/entigolabs/entigo-infralib-release">Releases</a>
</p>

---

> **Looking to provision a platform? You want the [agent](https://github.com/entigolabs/entigo-infralib-agent).** It's the tool you install and run. This Infralib repo is where modules are developed and tested — useful if you're contributing a module or want to see how one works.

> **In production since 2023.** Released and tested every weekday against live AWS and Google Cloud accounts. Used by Estonia's Information System Authority (RIA) and the Health and Welfare Information Systems Centre (TEHIK).

New here? Follow the [quickstart guide](https://infralib-quickstart.dev.entigo.dev/).

## What this is

This repository holds the building blocks — the opinionated Terraform modules and Kubernetes Helm charts we repeatedly use to run real platforms: networking, EKS/GKE clusters, autoscaling, ArgoCD, ingress, observability, DNS and TLS, secrets, and more.

You can consume them in three ways:

- **With the [Infralib Agent](https://github.com/entigolabs/entigo-infralib-agent)** — describe the platform in one YAML file and let the agent provision and continuously update it. Most people want this.
- **Directly from Terraform** — reference a released module by tag (see [example](#example-usage)).
- **Directly from ArgoCD** — point an Application at a Helm chart path (see [example](#example-usage)).

## Modules

Modules live under [`modules/`](modules/), grouped by target:

| Group | What's inside |
|---|---|
| [`modules/aws`](modules/aws) | VPC, EKS, Karpenter, node groups, Route 53, KMS, ECR proxy, EFS, Transit Gateway, cost alerts, and more |
| [`modules/google`](modules/google) | VPC, GKE, node pools, Cloud DNS, KMS, GAR proxy, services |
| [`modules/k8s`](modules/k8s) | ArgoCD, Istio, external-dns, external-secrets, Prometheus, Grafana, Loki, Mimir, Alloy, Karpenter, cluster-autoscaler, Harbor, Trivy, Kyverno, SAML proxy, and more |

See [`modules/k8s/README.md`](modules/k8s/README.md) for chart-specific notes.

## Example usage

Reference a released module by tag. Releases are published to
[entigo-infralib-release](https://github.com/entigolabs/entigo-infralib-release).

**Terraform**

```hcl
module "main" {
  source                 = "git::https://github.com/entigolabs/entigo-infralib-release.git//modules/aws/vpc?ref=v1.0.14"
  prefix                 = "dev-net-main"
  elasticache_subnets    = []
  intra_subnets          = []
  one_nat_gateway_per_az = false
  vpc_cidr               = "10.112.0.0/16"
}
```

**ArgoCD Application**

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: 'external-dns-dev'
spec:
  destination:
    server: https://kubernetes.default.svc
    namespace: 'external-dns-dev'
  project: default
  sources:
    - repoURL: 'https://github.com/entigolabs/entigo-infralib-release.git'
      targetRevision: 'v1.0.14'
      path: "modules/k8s/external-dns"
      helm:
        ignoreMissingValueFiles: true
        valueFiles:
          - 'values.yaml'
          - 'values-aws.yaml'
        values: |
          global:
              aws:
                  account: "XXXX"
                  clusterOIDC: oidc.eks.eu-north-1.amazonaws.com/id/XXXX
  syncPolicy:
    syncOptions:
      - CreateNamespace=true
      - RespectIgnoreDifferences=true
```

## Repository layout

| Folder | Contents |
|---|---|
| [`modules/`](modules/) | Terraform modules and Kubernetes Helm charts |
| [`images/`](images/) | Runtime images for running infrastructure as code |
| [`providers/`](providers/) | Terraform provider configurations |

## Release process

Releases are cut from `main` roughly once per day:

1. **Nuke** — each evening the entigo-infralib AWS and Google Cloud accounts are torn down (Nuke action).
2. **Stable** — in the morning the latest release is installed and its tests run (Stable action).
3. **Upgrade** — the accounts are upgraded to the `main` branch and tests run again.
4. **Release** — if the tests pass and `main` differs from the last release, a new release is created (Release action).

Once a release is created, it is published to
[entigo-infralib-release](https://github.com/entigolabs/entigo-infralib-release),
where it can be used by the [Infralib Agent](https://github.com/entigolabs/entigo-infralib-agent)
or referenced directly from Terraform and ArgoCD.

## License

Licensed under [AGPL-3.0](LICENSE.txt). For commercial licensing, contact [entigo.com](https://www.entigo.com).
