terraform {
  required_version = ">= 1.9"

  cloud {
    organization = "xmtp"

    workspaces {
      tags = ["testnet"]
    }
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.62"
    }
  }
}

provider "aws" {
  region = "us-east-2"
  default_tags {
    tags = {
      Environment = "testnet"
    }
  }
}

provider "aws" {
  alias  = "eu-north-1"
  region = "eu-north-1"
  default_tags {
    tags = {
      Environment = "testnet"
    }
  }
}
