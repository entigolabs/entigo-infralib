<h1 align="center">Infralib Modules</h1>

<p align="center">
  <strong>Production-tested Terraform / OpenTofu modules and Helm charts for building a complete Kubernetes platform on AWS, Google Cloud or Azure.</strong>
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

> **In production since 2023.** Released and tested every weekday against live AWS, Google Cloud and Azure accounts. Used by Estonia's Information System Authority (RIA) and the Health and Welfare Information Systems Centre (TEHIK).

New here? Follow the [quickstart guide](https://infralib-quickstart.dev.entigo.dev/).

## What this is

This repository holds the building blocks — the opinionated Terraform / OpenTofu modules and Kubernetes Helm charts we repeatedly use to run real platforms: networking, EKS/GKE/AKS clusters, autoscaling, ArgoCD, ingress, observability, DNS and TLS, secrets, and more. The modules work with either Terraform or OpenTofu.

You can consume them in three ways:

- **With the [Infralib Agent](https://github.com/entigolabs/entigo-infralib-agent)** — describe the platform in one YAML file and let the agent provision and continuously update it. Most people want this.
- **Directly from Terraform or OpenTofu** — reference a released module by tag (see [example](#example-usage)).
- **Directly from ArgoCD** — point an Application at a Helm chart path (see [example](#example-usage)).

## Repository layout

| Folder | Contents |
|---|---|
| [`modules/`](modules/) | Terraform / OpenTofu modules and Kubernetes Helm charts |
| [`images/`](images/) | Runtime images for running infrastructure as code |
| [`providers/`](providers/) | Terraform / OpenTofu provider configurations |

## Modules

Modules live under [`modules/`](modules/), grouped by target:

| Group | What's inside |
|---|---|
| [`modules/aws`](modules/aws) | VPC, EKS, Karpenter, node groups, Route 53, KMS, ECR proxy, EFS, Transit Gateway, cost alerts, and more |
| [`modules/google`](modules/google) | VPC, GKE, node pools, Cloud DNS, KMS, GAR proxy, services |
| [`modules/azure`](modules/azure) | VPC, AKS, node pools, Azure DNS, KMS, ACR proxy, Crossplane |
| [`modules/k8s`](modules/k8s) | ArgoCD, Istio, external-dns, external-secrets, Prometheus, Grafana, Loki, Mimir, Alloy, Karpenter, cluster-autoscaler, Harbor, Trivy, Kyverno, SAML proxy, and more |

See [`modules/k8s/README.md`](modules/k8s/README.md) for chart-specific notes.

## Example usage

Reference a released module by tag. Releases are published to
[entigo-infralib-release](https://github.com/entigolabs/entigo-infralib-release).

**Terraform / OpenTofu**

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

## Release process

Releases are cut from `main` roughly once per day. Testing and releasing run
through [entigo-infralib-test](https://github.com/entigolabs/entigo-infralib-test),
the framework this repository's modules are tested with; the workflows under
`.github/workflows/` call its reusable ones.

1. **Nuke** — each evening the entigo-infralib AWS, Google Cloud, Oracle and Azure test accounts are torn down (Nuke action, one job per cloud, configured under [nuke/](nuke/README.md)).
2. **Stable** — in the morning the latest release is installed on every environment from [entigo-infralib-release](https://github.com/entigolabs/entigo-infralib-release) and its tests run (Stable action).
3. **Release** — after a green scheduled Stable run, `main` is applied to every environment and tested again; if the tests pass and `main` differs from the last release, a new release is created, published and the images are tagged with it (Release action).

Once a release is created, it is published to
[entigo-infralib-release](https://github.com/entigolabs/entigo-infralib-release),
where it can be used by the [Infralib Agent](https://github.com/entigolabs/entigo-infralib-agent)
or referenced directly from Terraform / OpenTofu and ArgoCD.

Releases are also published as OCI packages to the **AWS Public ECR**
and **GitHub Container Registry**, so modules can be consumed as an OCI source
without going through git.

Agent configuration examples:

```yaml
# OCI repository for AWS
sources:
  - url: oci://public.ecr.aws/entigolabs/entigo-infralib-release

# OCI repository for Google Cloud and Azure
sources:
  - url: oci://ghcr.io/entigolabs/entigo-infralib-release

# GIT repository for any
sources:
  - url: https://github.com/entigolabs/entigo-infralib-release

# GIT branch on our development repo
sources:
  - url: https://github.com/entigolabs/entigo-infralib
    force_version: true
    version: main
```

## Testing

Every module carries its tests next to it, as Go tests under
`modules/<type>/<name>/test/`, written against the packages of
[entigo-infralib-test](https://github.com/entigolabs/entigo-infralib-test).
The test environments are agent configurations under `environments/`: one
file per cloud and prefix (`aws_biz.yaml`, `google_pri.yaml`, ...), listing
the modules of each step. A module is part of an environment when a step
lists it; its inputs for that environment are `test/<environment>.yaml` in
the module, which hold only what the scenario changes from the module's
defaults. `test/module.yaml` marks the modules that exist once per
environment (`pin_step`) and lists kube-score checks to skip for charts whose
upstream parts fail them.

```
./test.sh                                 provision every environment you have credentials for, then test every module
./test.sh modules/k8s/argocd              test one module in a step of its own on the shared environments
modules/k8s/argocd/test.sh                the same, from the module's directory
./test.sh --env aws_biz test              only the tests, environments already provisioned
./test.sh static modules/aws/vpc          static checks only: tflint, or helm lint + template + kube-score
./test.sh envs                            the environments and their modules
./test.sh --help
```

`test.sh` needs bash and docker; it fetches the framework's orchestrator of
the version it pins and runs the agent and the tests from the framework's
per-cloud images with your credentials and kubeconfig. Credentials and regions
come from the variables the agent reads: `AWS_REGION` and AWS credentials,
`GOOGLE_PROJECT`, `GOOGLE_REGION`, `GOOGLE_ZONE` and a gcloud login or key,
`OCI_REGION`, `OCI_COMPARTMENT_ID` and `~/.oci/config`, `AZURE_LOCATION`,
`AZURE_SUBSCRIPTION_ID` and a service principal. `go.mod` requires the same
framework version `test.sh` pins; after bumping the pin, run the tests once
and commit the `go.mod` and `go.sum` that produces.

## License

Licensed under [AGPL-3.0](LICENSE.txt). For commercial licensing, contact [entigo.com](https://www.entigo.com).
