#!/bin/bash

# Exit immediately if a command exits with a non-zero status
set -e

# --- CONFIGURATION ---
SERVER_IP=""
SERVER_PORT=""
SSH_KEY=""
REMOTE_DIR=""
APP_NAME=""

echo "📦 Packaging source code..."
# Creates a temporary archive without node_modules or local build artifacts
tar --exclude='node_modules' --exclude='dist' --exclude='backend/app' -czf source.tar.gz backend frontend Dockerfile

echo "🚀 Transferring source code to the server..."
ssh -p $SERVER_PORT -i $SSH_KEY root@$SERVER_IP "mkdir -p $REMOTE_DIR"
scp -P $SERVER_PORT -i $SSH_KEY source.tar.gz root@$SERVER_IP:$REMOTE_DIR/

# Remove the local archive after upload
rm source.tar.gz

echo "🏗️ Starting remote deployment pipeline..."
ssh -p $SERVER_PORT -i $SSH_KEY root@$SERVER_IP << EOF
  cd $REMOTE_DIR
  
  # Extract source code
  tar -xzf source.tar.gz
  rm source.tar.gz
  
  # Build the new Docker image on the server using the multi-stage Dockerfile
  echo "🐳 Building Docker image on the server..."
  docker build -t ${APP_NAME}:latest .
  
  # Stop and remove the old container (|| true prevents script failure if the container doesn't exist)
  echo "🛑 Stopping and removing old container..."
  docker kill ${APP_NAME} 2>/dev/null || true
  docker rm ${APP_NAME} 2>/dev/null || true
  
  # Run the new container
  echo "▶️ Launching the new container..."
  docker run -d --name ${APP_NAME} -p 8081:8080 ${APP_NAME}:latest
  
  # Optional: Clean up dangling images to save disk space
  docker image prune -f
  
  echo "✅ Deployment completed successfully!"
EOF