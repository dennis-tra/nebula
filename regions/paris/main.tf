terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 3.5.0"
    }
  }
  required_version = ">= 0.13"
}

provider "aws" {
  region = "eu-west-3"
}

resource "aws_instance" "ipfs-crawler" {
  ami           = "ami-0f7cd40eac2214b37"
  instance_type = "t2.small"
  tags = {
    Name = "ipfs-crawler"
  }
  security_groups = ["security_ipfs_crawler"]
  user_data       = <<-EOF
    #!/bin/sh
    cd /home/ubuntu/
    sudo apt-get update
    sudo apt install -y unzip git make build-essential ca-certificates curl gnupg lsb-release
    curl -fsSL https://get.docker.com -o get-docker.sh
    chmod u+x ./get-docker.sh
    sudo curl -L "https://github.com/docker/compose/releases/download/1.29.2/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    sudo chmod +x /usr/local/bin/docker-compose
    sudo ./get-docker.sh
    git clone https://github.com/wcgcyx/nebula-crawler.git
    cd nebula-crawler
    cd controller
    make docker
    cd deploy
    nohup docker-compose up > /dev/null 2>&1 &
  EOF
}

resource "aws_security_group" "security_ipfs_crawler" {
  name        = "security_ipfs_crawler"
  description = "security group for ipfs crawler"

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 4001
    to_port     = 4001
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 4001
    to_port     = 4001
    protocol    = "udp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "security_ipfs_crawler"
  }
}
