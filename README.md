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


# fly.io

## Setup fly.io from your local machine
1. In your local terminal run ```fly launch``` command and follow the instructions to launch your app.
2. Once you have launched your app, a fly.toml file will be created.
3. Add the following to the fly.toml file:
```
[mounts]
  source = "sortedstream_data"
  destination = "/data"
```
4. create a new volume in fly.io
```
fly volumes create sortedstream_data -r <region> --size 10
```
6. We have added sample fly.toml file in deploy folder, you can use it to deploy the app.
7. As we are using firebase authentication, convert the firebase secret json into base64 and create a secret in fly.io. Don't use online tools to convert the json into base64. Use your local machine to convert the json into base64.
```
cat firebase-secret.json | base64 -w0
fly secrets set GOOGLE_APPLICATION_CREDENTIALS_BASE64=<base64_encoded_json>
```
8. (Optional) To restrict login access to specific emails, set the ALLOWED_EMAILS environment variable with comma-separated email addresses:
```
fly secrets set ALLOWED_EMAILS=user1@example.com,user2@example.com,admin@yourcompany.com
```
If this variable is not set, all authenticated users will be allowed to access the app.

## Deploy in fly.io
1. In .github/workflows folder, we do have build yaml files, which will build the binary, generate docker image and push to github container registry.
2. To trigger the build, you have to create a tag in the repo. For example, if you want to deploy version 1.0.0, you have to create a tag with name v1.0.0.
3. Once github action upload docker image, then you just have to deploy this docker image to fly.io.
```
fly deploy --image ghcr.io/sortedstartup/stream/sortedstream:binary-production-latest

```



