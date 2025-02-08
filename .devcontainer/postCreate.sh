#!/bin/bash

ROOT_DIR=$(pwd)
echo "Root dir = $ROOT_DIR"

# function to download go dependencies
download_go_dependencies() {
    echo "Downloading go dependencies"
    cd $ROOT_DIR/backend
    # # GOMAXPROCS is supposed to download all dependencies in parallel, need to verify
    # GOMAXPROCS=10 go mod download -x

    # This does a build therefore populated build cache and downloads dependencies
    # Due to this the go run will be extremely fast
    GOMAXPROCS=10 go mod download -x
    if [ $? -ne 0 ]; then
        echo "Ignoring Failure: Failed to download go dependencies"
        return 0
    fi
}

# function to download node dependencies
download_node_dependencies() {
    echo "Downloading node dependencies"
    cd $ROOT_DIR/frontend/webapp
    npm install -g pnpm
    pnpm install
    if [ $? -ne 0 ]; then
        echo "Ignoring Failure: Failed to download node dependencies"
        return 0
    fi
}

get_firebase_credentials() {
    # The env variables $FIREBASE_ADMIN_CREDENTIALS_BASE64 supplied by Github codespaces secret automatically to any codespace
    if [ -z "$FIREBASE_ADMIN_CREDENTIALS_BASE64" ]; then
    echo "ERROR: FIREBASE_ADMIN_CREDENTIALS_BASE64 is not set"
    return 1
    fi

    echo $FIREBASE_ADMIN_CREDENTIALS_BASE64 | base64 -d > .secrets/firebase-admin-credentials.json
    if [ $? -ne 0 ]; then
    echo "ERROR: Failed to decode FIREBASE_ADMIN_CREDENTIALS_BASE64 or write it to file [./firebase-admin-credentials.json]"
    return 1
    fi

    echo "Successfully decoded FIREBASE_ADMIN_CREDENTIALS_BASE64 to [./firebase-admin-credentials.json]"
    return 0
}

# --- Script starts here ---
get_firebase_credentials
download_go_dependencies &
download_node_dependencies &
# --- Script ends here ---

# Wait for background function to end
wait

echo "Script execution ends"

