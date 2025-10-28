A simple blockchain-based library checkout system written in Go.
Each book checkout is stored as a block in a blockchain, ensuring immutability and traceability of transactions.

Features

Custom blockchain implementation using SHA-256 hashing

REST API for adding and retrieving book checkout records

Book creation with MD5-based unique ID generation

Genesis block creation at startup

Proof of work concept of blockchain 

Basic block validation (hash integrity, position check, chain linkage)


🚀 Run the Application
1. Clone the repository
git clone https://github.com/ianmolbisht/simple-blockchain-using-Go.git
cd go-library-blockchain

2. Install dependencies
go mod init go-library-blockchain
go get github.com/gorilla/mux

3. Run the server
go run main.go


Server runs on:

http://localhost:3000
