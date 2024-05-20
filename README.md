# Nibiru Go SDK - NibiruChain/nibiru/gosdk

A Golang client for interacting with the Nibiru blockchain.

--- 

## Dev Notes

### Finalizing "v1"

- [ ] Write usage examples
- [ ] Create a quick start guide
- [ ] Migrate to the [Nibiru repo](https://github.com/NibiruChain/nibiru) and archive this one.

### Mini Sprint

- [ ] feat: add in transaction broadcasting
  - [x] initialize keyring obj with mnemonic
  - [x] initialize keyring obj with priv key
  - [ ] write a test that sends a bank transfer.
  - [ ] write a test that submits a text gov proposal.
- [x] impl Tendermint RPC client
- [x] refactor: DRY improvements on the QueryClient initialization
- [x] ci: Add go tests to CI
- [x] ci: Add code coverage to CI
- [ ] ci: Add linting to CI
- [ ] docs: for grpc.go
- [ ] docs: for clients.go
- [ ] impl wallet abstraction for the keyring
- [ ] epic: storing transaction history storage 

### Question Brain-dump

- Q: Should gosdk run as a binary? -> No.
- Q: Which functionality warrants localnet testing?
- Q: Should there be a way to run queries with JSON-RPC 2 instead of GRPC?
