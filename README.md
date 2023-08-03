# gonibi

A Golang client for interacting with the Nibiru blockchain.

--- 

## Dev Notes

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

- Q: Should gonibi run as a binary? -> Yes. What should it be able to do?
- Q: Which functionality warrants localnet testing?
- Q: Should there be a way to run queries with JSON-RPC 2 instead of GRPC?
