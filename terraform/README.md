# xmtpd Terraform modules

> [!WARNING]
> This software is **experimental** and in early development. Expect frequent changes and unresolved issues.

`xmtpd` (XMTP daemon) is an experimental version of XMTP node software. It is **not** the node software that currently forms the XMTP network.

This repository includes Terraform modules that can be used as part of a new or existing Terraform plan to add `xmtpd` nodes to your infrastructure.

Currently only AWS is supported with plans for future support for Google Cloud and other hosting providers.

## Prerequisites

1. An AWS account
2. An IAM user in that account with privileges to add and remove a wide range of infrastructure (the built-in Administrator policy will work, or you can make a more tailored policy if desired)
3. AWS credentials for that user (`AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`)
4. An Alchemy account with an active subscription, and an app in that account with access to the following blockchains:
   - Ethereum
   - Base
   - Arbitrum One
   - OP Mainnet
   - Polygon Mainnet
   - ZKSync Mainnet
   - Linea Mainnet
   - XMTP Ropsten
5. A private key for your node, which can be generated using the instructions in the [onboarding guide](https://github.com/xmtp/xmtpd/blob/main/doc/onboarding.md) in the `xmtpd` repo.

## Register a domain

1. Choose a new domain name for your node
2. Register the domain in [Route53](https://us-east-1.console.aws.amazon.com/route53/v2/home?region=us-east-2#Dashboard)
3. Wait for the registration to complete (may take 15-20 minutes)

## Deploying into a new VPC

### Set up terraform variables

#### With Terraform Cloud

1. Create a workspace named `xmtp-testnet`
2. Add the following variables to the workspace

```terraform
verifier_chain_rpc_urls = {
    chain_rpc_1 = $ETHEREUM_RPC_URL
    chain_rpc_8453 = $BASE_RPC_URL
    chain_rpc_42161 = $ARBITRUM_ONE_RPC_URL
    chain_rpc_10 = $OP_MAINNET_RPC_URL
    chain_rpc_324 = $POLYGON_MAINNET_RPC_URL
    chain_rpc_59144 = $LINEA_MAINNET_RPC_URL
}
contracts = $CONTRACTS_JSON_CONFIG
app_chain_wss_url = $XMTP_SEPOLIA_WSS_URL
settlement_chain_wss_url = $BASE_SEPOLIA_WSS_URL
app_chain_rpc_url = $XMTP_SEPOLIA_RPC_URL
settlement_chain_rpc_url = $BASE_SEPOLIA_RPC_URL
signer_private_key = $GENERATED_PRIVATE_KEY
domain_name = $YOUR_NEW_DOMAIN_NAME
```

#### Without Terraform Cloud

1. Create a file named `terraform.tfvars` in the `examples/aws-complete` folder
2. Add the following secrets to the file, with any variables substituted with the appropriate values from your accounts

```terraform
verifier_chain_rpc_urls = {
    chain_rpc_1 = $ETHEREUM_RPC_URL
    chain_rpc_8453 = $BASE_RPC_URL
    chain_rpc_42161 = $ARBITRUM_ONE_RPC_URL
    chain_rpc_10 = $OP_MAINNET_RPC_URL
    chain_rpc_324 = $POLYGON_MAINNET_RPC_URL
    chain_rpc_59144 = $LINEA_MAINNET_RPC_URL
}
contracts = $CONTRACTS_JSON_CONFIG
app_chain_wss_url = $XMTP_SEPOLIA_WSS_URL
settlement_chain_wss_url = $BASE_SEPOLIA_WSS_URL
app_chain_rpc_url = $XMTP_SEPOLIA_RPC_URL
settlement_chain_rpc_url = $BASE_SEPOLIA_RPC_URL
signer_private_key = $GENERATED_PRIVATE_KEY
domain_name = $YOUR_NEW_DOMAIN_NAME
```

### Deploy the plan

#### With Terraform Cloud

1. `cd ./examples/aws-complete`
2. Modify the file `_providers.tf` and set the organization to your Terraform Cloud organization and the workspace to a new workspace in your organization.
3. In Terraform Cloud make sure the `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` environment variables are set to the credentials for your Terraform IAM user
4. `terraform init` (you may need to run `terraform login` first if you have never used the Terraform CLI before)
5. `terraform plan` and review the plan
6. `terraform apply`

#### Without Terraform Cloud

1. `cd ./examples/aws-complete`
2. In the `_providers.tf` file, remove the `cloud` block
3. `terraform init`
4. `terraform apply`
