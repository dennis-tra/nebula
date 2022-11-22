terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.40.0"
    }
  }
  required_version = ">= 1.0"
}

provider "aws" {
  # AMI: ami-0caef02b518350c8b
  region = "eu-central-1"
  alias  = "eu_central_1"
}

provider "aws" {
  # AMI: ami-02ea247e531eb3ce6
  region = "us-west-1"
  alias  = "us_west_1"
}

provider "aws" {
  # AMI: ami-097a2df4ac947655f
  region = "us-east-1"
  alias  = "us_east_1"
}

provider "aws" {
  # AMI: ami-04b3c23ec8efcc2d6
  region = "sa-east-1"
  alias  = "sa_east_1"
}

provider "aws" {
  # AMI: ami-07651f0c4c315a529
  region = "ap-southeast-1"
  alias  = "ap_southeast_1"
}

variable "ssh_key" {
  type = string
}

locals {
  default_tags = {
    managed_by = "terraform"
    project    = "nebula-hydra-dial-down"
  }
}

module "nebula_eu_central_1_offset_0" {
  source = "./node"

  ami          = "ami-0caef02b518350c8b"
  ssh_key      = var.ssh_key
  offset       = 0
  region       = "eu-central-1"
  default_tags = local.default_tags

  providers = {
    aws = aws.eu_central_1
  }
}

module "nebula_us_west_1_offset_1" {
  source = "./node"

  ami          = "ami-02ea247e531eb3ce6"
  ssh_key      = var.ssh_key
  offset       = 1
  region       = "us-west-1"
  default_tags = local.default_tags

  providers = {
    aws = aws.us_west_1
  }
}

module "nebula_us_east_1_offset_2" {
  source = "./node"

  ami          = "ami-097a2df4ac947655f"
  ssh_key      = var.ssh_key
  offset       = 2
  region       = "us-east-1"
  default_tags = local.default_tags

  providers = {
    aws = aws.us_east_1
  }
}

module "nebula_sa_east_1_offset_3" {
  source = "./node"

  ami          = "ami-04b3c23ec8efcc2d6"
  ssh_key      = var.ssh_key
  offset       = 3
  region       = "sa-east-1"
  default_tags = local.default_tags

  providers = {
    aws = aws.sa_east_1
  }
}

module "nebula_ap_southeast_1_offset_4" {
  source = "./node"

  ami          = "ami-07651f0c4c315a529"
  ssh_key      = var.ssh_key
  offset       = 4
  region       = "ap-southeast-1"
  default_tags = local.default_tags

  providers = {
    aws = aws.ap_southeast_1
  }
}

module "nebula_eu_central_1_offset_5" {
  source = "./node"

  ami          = "ami-0caef02b518350c8b"
  ssh_key      = var.ssh_key
  offset       = 5
  region       = "eu-central-1"
  default_tags = local.default_tags

  providers = {
    aws = aws.eu_central_1
  }
}

module "nebula_us_west_1_offset_6" {
  source = "./node"

  ami          = "ami-02ea247e531eb3ce6"
  ssh_key      = var.ssh_key
  offset       = 6
  region       = "us-west-1"
  default_tags = local.default_tags

  providers = {
    aws = aws.us_west_1
  }
}

module "nebula_us_east_1_offset_7" {
  source = "./node"

  ami          = "ami-097a2df4ac947655f"
  ssh_key      = var.ssh_key
  offset       = 7
  region       = "us-east-1"
  default_tags = local.default_tags

  providers = {
    aws = aws.us_east_1
  }
}

module "nebula_sa_east_1_offset_8" {
  source = "./node"

  ami          = "ami-04b3c23ec8efcc2d6"
  ssh_key      = var.ssh_key
  offset       = 8
  region       = "sa-east-1"
  default_tags = local.default_tags

  providers = {
    aws = aws.sa_east_1
  }
}

module "nebula_ap_southeast_1_offset_9" {
  source = "./node"

  ami          = "ami-07651f0c4c315a529"
  ssh_key      = var.ssh_key
  offset       = 9
  region       = "ap-southeast-1"
  default_tags = local.default_tags

  providers = {
    aws = aws.ap_southeast_1
  }
}