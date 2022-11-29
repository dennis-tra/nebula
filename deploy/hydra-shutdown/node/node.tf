terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.40.0"
    }
  }
  required_version = ">= 1.0"
}
variable "ami" {
  type = string
}

variable "ssh_key" {
  type = string
}

variable "region" {
  type = string
}

variable "offset" {
  type = number
}

variable "frequency" {
  type    = number
  default = 10
}

variable "default_tags" {
  type = map(any)
}

output "ip_address" {
  value = aws_instance.nebula_node.public_ip
}

resource "aws_instance" "nebula_node" {
  ami           = var.ami
  instance_type = "t3.medium"
  key_name      = aws_key_pair.ssh_key.id

  security_groups = [aws_security_group.nebula_instance.name]

  root_block_device {
    volume_size = 20 # 20 GiB
  }

  user_data = <<-EOF
    #!/bin/sh
    cd /home/ubuntu/
    sudo apt-get update

    # Install docker
    sudo apt-get install -y ca-certificates curl gnupg lsb-release git
    sudo mkdir -p /etc/apt/keyrings
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
    sudo apt-get update
    sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

    # Download Nebula code
    git clone https://github.com/dennis-tra/nebula.git
    cd nebula/deploy

    # Start Nebula services
    sudo docker compose up -d

    # Configure cronjob
    echo "${var.offset}-59/${var.frequency} * * * * docker run --rm --network nebula --name nebula_crawler --hostname nebula_crawler dennistra/nebula-crawler:sha-ce5c756 nebula --db-host=postgres crawl --neighbors" > cronjobs
    sudo crontab cronjobs
    rm cronjobs
  EOF

  tags = merge(var.default_tags, {
    Name = "nebula-instance-${var.region}-offset-${var.offset}"
  })
}

resource "aws_key_pair" "ssh_key" {
  key_name   = "nebula-hydra-dial-down-ssh-key-offset-${var.offset}"
  public_key = var.ssh_key
}

resource "aws_security_group" "nebula_instance" {
  name        = "Nebula Security Group (Offset ${var.offset})"
  description = "Opens Ports for all service running on a Nebula instance"

  ingress {
    description      = "SSH Access"
    from_port        = 22
    to_port          = 22
    protocol         = "tcp"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }

  ingress {
    description      = "Grafana Access"
    from_port        = 3000
    to_port          = 3000
    protocol         = "tcp"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }

  ingress {
    description      = "Prometheus Access"
    from_port        = 9090
    to_port          = 9090
    protocol         = "tcp"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }

  ingress {
    description      = "Metrics Endpoint Access"
    from_port        = 6667
    to_port          = 6667
    protocol         = "tcp"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }

  egress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }

  tags = var.default_tags
}