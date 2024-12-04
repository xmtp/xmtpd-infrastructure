# Contributing Guidelines

Pull requests, bug reports, and all other forms of contribution are welcomed and highly encouraged! :octocat:

## Testing

To run the test suite you will have to provide it with a set of private keys for the node.
The private key needs to be registered with the blockchain smart contract and a NodeID has to be issued. For more info see [Onboarding](https://github.com/xmtp/xmtpd/blob/main/doc/onboarding.md)

Create a file in the top level directory called `secrets.yaml` and set your node private keys such as:
```
env.secret.XMTPD_SIGNER_PRIVATE_KEY: "<key>"
env.secret.XMTPD_PAYER_PRIVATE_KEY: "<key>"
```