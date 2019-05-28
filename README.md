# sebak go sdk 

## Create new transaction

### Prerequisite
Before creating new transaction, you should check these,

* 'secret seed' of source account
* 'public address' of target account
* 'network id'

You can simply check 'network id' from SEBAK node information. If the address of your sebak node is 'https://testnet-sebak.blockchainos.org',
```sh
$ curl -v https://testnet-sebak.blockchainos.org
...
  "policy": {
    "network-id": "sebak-test-network",
    "initial-balance": "10000000000000000000",
    "base-reserve": "1000000",
...
```

The value of `"network-id"`, `sebak-test-network` is the 'network id'.

### `CreateAccount`

* `target` address must new account, this means, it does not exist in the SEBAK network. You can check the account status thru account API of SEBAK. Please see http://devteam.blockchainos.org/docs/api/#accounts-account-details-get .
* `amount` for creating account must be bigger than base reserve, you can check the amount from SEBAK node information like 'network-id'

### `Payment`

* `target` address must exist in network.

With this transaction example, you can submit transactons using following cli command.
Operation type can be "create" or "payment"

```sh
$ go run sebak_go_sdk.go {source account secret seed} {target account public address} {operation type}
```
