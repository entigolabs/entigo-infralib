instance_type = "t3.micro"
eip = true
key_name = "martivo_x220"
user_data = "apt-get update -y && apt-get install -y curl python3-pip bash-completion && pip3 install --system awscli --upgrade"
