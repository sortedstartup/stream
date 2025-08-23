---
description: 
globs: backend/**,frontend/**
alwaysApply: false
---
This a mono repo containing many microservices, 
Every microservice has its own proto file
you MUST follow microservice best practices including 
1. every microservice has its own db
2. microservice talk only using API layers to each other. In this project api is expressed in using proto

We have 4 microservices:
1. videoservice
2. userservice  
3. commentservice
4. paymentservice

Tech stack: Go, React, nanostores (state management) sqlite, firebase auth

**Response Guidelines: Keep responses concise. Avoid detailed explanations unless specifically requested by the user.**

## Architecture Patterns

### Mono Service Pattern:
- All microservices run under a single mono service (backend/mono/main.go)
- Mono service handles server ports, not individual microservices
- Individual services only define business logic and database operations
- Each service has its own SQLite database

### Configuration Pattern:
- All services follow same config pattern: ServiceConfig + DBConfig
- Uses mapstructure tags for config file loading (not environment variables)
- Database uses connection URLs (DBConfig.Url), not file paths
- Service-specific configs only contain business logic settings
- Example structure:
  ```go
  type ServiceConfig struct {
      DB DBConfig `json:"db" mapstructure:"db"`
      // service-specific fields only
  }
  type DBConfig struct {
      Driver string `json:"driver" mapstructure:"driver"`
      Url    string `json:"url" mapstructure:"url"`
  }
  ```

### Proto Integration Pattern:
- Proto files stored in root /proto folder
- Each service generates own proto stubs in service/proto/ folder  
- Run `go generate` in service folder to create: proto stubs + SQLC queries
- Cross-service communication ONLY via gRPC APIs
- Never bypass API layer between services

### Payment Service Architecture:
- Generic design: application-agnostic, reusable across different apps
- User-based subscriptions: user pays → quota applies across all their workspaces/tenants
- Integration pattern: Application services call payment service to check access before actions
- Provider-agnostic: supports multiple payment providers (Stripe, Razorpay)
- Environment-specific settings (like price IDs) stored in config, not database

### Stream App Integration Flow:
- Access check pattern: Get tenant owner (userservice) → Check owner's subscription (paymentservice) → Allow/deny action
- Payment service doesn't know about tenants/workspaces (handled by userservice)
- Payment service tracks: user subscriptions, usage limits, plan definitions
- Usage types: storage limits (bytes), user limits (count)
- Real-time usage tracking: file uploads, user additions

Repository structure
.
|-- LICENSE
|-- README.md
|-- TODO.md
|-- backend
|   |-- Dockerfile
|   |-- commentservice
|   |   |-- api
|   |   |   |-- api.go
|   |   |   `-- api_test.go
|   |   |-- autogenerate.go
|   |   |-- config
|   |   |   `-- config.go
|   |   |-- db
|   |   |   |-- db.go
|   |   |   |-- migrate.go
|   |   |   |-- migrations
|   |   |   |   |-- 1_....sql
                    ....
|   |   |   |-- models.go
|   |   |   |-- querier.go
|   |   |   |-- queries.sql.go
|   |   |   `-- scripts
|   |   |       |-- queries.sql
|   |   |       `-- sqlc.yaml
|   |   |-- go.mod
|   |   `-- proto
|   |       |-- commentservice.pb.go
|   |       `-- commentservice_grpc.pb.go
|   |-- common
|   |   |-- auth
|   |   |   |-- auth.go
|   |   |   `-- firebase.go
|   |   |-- constants
|   |   |   `-- constants.go
|   |   |-- go.mod
|   |   |-- go.sum
|   |   |-- interceptors
|   |   |   |-- auth.go
|   |   |   `-- panic_recovery.go
|   |   `-- validation
|   |       |-- email_validation.go
|   |       |-- email_validation_test.go
|   |       `-- string_validations.go
|   |-- db.sqlite
|   |-- go.mod
|   |-- go.sum
|   |-- go.work
|   |-- go.work.sum
|   |-- mono
|   |   |-- config
|   |   |   `-- config.go
|   |   |-- db.sqlite
|   |   |-- functions
|   |   |-- go.mod
|   |   |-- go.sum
|   |   |-- main.go
|   |   |-- version
|   |   |   |-- version.githash
|   |   |   |-- version.go
|   |   |   `-- version.number
|   |   `-- webapp
|   |       `-- dist
|   |           `-- README.md
|   |-- uploads
|   |-- userservice
|   |   |-- api
|   |   |   `-- api.go
|   |   |-- autogenerate.go
|   |   |-- config
|   |   |   `-- config.go
|   |   |-- db
|   |   |   |-- db.go
|   |   |   |-- migrate.go
|   |   |   |-- migrations
|   |   |   |   `-- 1_....sql
|   |   |   |-- models.go
|   |   |   |-- queries.sql.go
|   |   |   `-- scripts
|   |   |       |-- queries.sql
|   |   |       `-- sqlc.yaml
|   |   |-- go.mod
|   |   `-- proto
|   |       |-- userservice.pb.go
|   |       `-- userservice_grpc.pb.go
|   |-- paymentservice
|   |   |-- api
|   |   |   |-- api.go
|   |   |   `-- http.go
|   |   |-- autogenerate.go
|   |   |-- config
|   |   |   `-- config.go
|   |   |-- db
|   |   |   |-- db.go
|   |   |   |-- migrate.go
|   |   |   |-- migrations
|   |   |   |-- models.go
|   |   |   |-- queries.sql.go
|   |   |   `-- scripts
|   |   |       |-- queries.sql
|   |   |       `-- sqlc.yaml
|   |   |-- go.mod
|   |   |-- providers
|   |   |   `-- stripe.go
|   |   `-- proto
|   |       |-- paymentservice.pb.go
|   |       `-- paymentservice_grpc.pb.go
|   `-- videoservice
|       |-- api
|       |   |-- api.go
|       |   |-- http.go
|       |   `-- http_test.go
|       |-- autogenerate.go
|       |-- config
|       |   `-- config.go
|       |-- db
|       |   |-- db.go
|       |   |-- migrate.go
|       |   |-- migrations
|       |   |   |-- 1_init.up.sql
|       |   |   `-- 2_....sql
|       |   |-- models.go
|       |   |-- queries.sql.go
|       |   `-- scripts
|       |       |-- queries.sql
|       |       `-- sqlc.yaml
|       |-- go.mod
|       |-- go.sum
|       |-- processing
|       `-- proto
|           |-- videoservice.pb.go
|           `-- videoservice_grpc.pb.go
|-- deploy
|   |-- Dockerfile
|   |-- prod
|   |   `-- fly.toml
|   `-- staging
|       `-- fly.toml
|-- docs
|   |-- html5-mediasource.md
|   `-- info-blob
|       |-- about-video-streaming-in-browser.md
|       `-- video-html5.excalidraw.png
|-- frontend
|   `-- webapp
|       |-- README.md
|       |-- src
|       |   |-- App.css
|       |   |-- App.jsx
|       |   |-- assets
|       |   |   `-- react.svg
|       |   |-- auth
|       |   |   |-- components
|       |   |   |   `-- ProtectedRoute.jsx
|       |   |   |-- pages
|       |   |   |   `-- LoginPage.jsx
|       |   |   |-- providers
|       |   |   |   `-- firebase-auth.ts
|       |   |   |-- store
|       |   |   |   `-- auth.ts
|       |   |   `-- utils
|       |   |-- components
|       |   |   |-- CommentSection.tsx
|       |   |   |-- ListOfVideos.jsx
|       |   |   |-- ScreenRecorder.jsx
|       |   |   `-- layout
|       |   |       |-- Footer.jsx
|       |   |       |-- Header.jsx
|       |   |       |-- Layout.jsx
|       |   |       `-- Sidebar.jsx
|       |   |-- index.css
|       |   |-- main.jsx
|       |   |-- pages
|       |   |   |-- HomePage.jsx
|       |   |   |--.....
|       |   |-- proto
|       |   |   |-- commentservice.ts
|       |   |   |-- paymentservice.ts
|       |   |   |-- google
|       |   |   |   `-- protobuf
|       |   |   |       `-- timestamp.ts
|       |   |   `-- videoservice.ts
|       |   |-- stores
|       |   |   |-- comments.ts
|       |   |   `-- videos.ts
|       |   |-- version
|       |   |   |-- get-git-hash.cjs
|       |   |   |-- version.number
|       |   |   `-- versionInfo.ts
|       |   `-- vite-env.d.ts
|       |-- tailwind.config.js
|       |-- tsconfig.json
|       |-- tsconfig.node.json
|       `-- vite.config.js
|-- package-lock.json
|-- proto
|   |-- commentservice.proto
|   |-- paymentservice.proto
|   |-- userservice.proto
|   `-- videoservice.proto

Features can be read from docs/features.md file

How to implement a new API:
- Identify which microservice(s) to make changes in 
- Update proto files in root level "proto" folder. Once we made proto changes, then we can make changes in respective backend api folder, by adding migrations, scripts/queries.sql. 
Then ask user to run 'go generate' in respective service folder in backend, they will run, it will auto generate few files.
`go generate` will generate:
0. go generate works in the context of a given microservice
1. server stubs for proto file 
2. go code for queries.sql under db/queries.sql.go

Note: Make sure we never bypass tenant security. No cross tenant leaks. This is important. 