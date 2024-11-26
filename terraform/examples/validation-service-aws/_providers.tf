terraform {
  required_version = ">= 1.9"

  cloud {
    organization = "xmtp"

    workspaces {
      name = "public-testnet-playground"
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
