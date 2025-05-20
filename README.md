SortedStartup Stream
====================

# The Idea
We wanted a self hosted / cloud hosted open source tool to help us do screen recording, create a library of those recording and share it internally with out team for sharing knowledge, communications, training etc.
We found few tools which we liked like Loom, but they were closed source and paid per user even if we did not use it much. It felt we were held hostage to their monthly price. So we decided to build our own tool which is open source and can be self hosted.

We noticed that the web APIs available in every browser for recording and playing video make our effort very easy.

# The Tech Stack
- Frontend: React
- Backend: golang
- Database: sqlite
- Video Store : File System, Sqlite, S3 compatible storage

# License
As of now we feel GPL v2 is the right license for us.
We will review it based on community feedback as we go along.

# Generate binary
1. `GOOS=linux GOARCH=amd64 go build -ldflags "-X main.Version=1.0.0 -X main.BuildTime=$(shell date -u '+%Y-%m-%d_%H:%M:%S')" -o ../../bin/stream-server-linux`