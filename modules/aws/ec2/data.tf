data "aws_kms_key" "alias" {
  var.kms_key_id != "" ? 1 : 0
  key_id = var.kms_key_id
}
