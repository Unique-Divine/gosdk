# gonibi

A Golang client for interacting with the Nibiru blockchain.

--- 

## Dev Notes

### Mini Sprint

- [ ] feat: add in transaction broadcasting
  - [ ] initialize with mnemonic
  - [ ] initialize with priv key
  - [ ] write a test that sends a bank transfer.
  - [ ] write a test that submits a text gov proposal.
- [ ] refactor: DRY improvements on the QueryClient initialization
- [x] ci: Add go tests to CI
- [x] ci: Add code coverage to CI
- [ ] ci: Add linting to CI
- [ ] docs: for grpc.go
- [ ] docs: for clients.go
- [ ] epic: storing transaction history storage 

### Brain-dumped Questions

- Q: Which functionality warrants localnet testing?
- Q: Should there be a way to run queries with JSON-RPC 2 instead of GRPC?
