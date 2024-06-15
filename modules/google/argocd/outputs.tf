output "name" {
    value = resource.helm_release.argocd.name
}

output "hname" {
    value = local.hname
}