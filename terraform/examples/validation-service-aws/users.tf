locals {
  # sort alphabetically
  developer_emails = toset([
    "martin@ephemerahq.com",
    "nicholas@ephemerahq.com",
    "rich@ephemerahq.com",
  ])
}


module "grant_user_access" {
  source = "./modules/grant_user_access"

  developer_emails = local.developer_emails

  providers = {
    aws = aws
  }
}