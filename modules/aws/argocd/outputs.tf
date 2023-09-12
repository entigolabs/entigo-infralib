output "ssh-pub-key" {
  value = tls_private_key.argocd.public_key_openssh
  description = "The Public SSH key that argocd will use to access GIT repository."
}

output "ssh-pub-key-id" {
  value = aws_iam_user_ssh_key.argocd.ssh_public_key_id
  description = "The public key id that argocd will use to access GIT repository."
}
