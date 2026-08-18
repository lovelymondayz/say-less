#!/bin/bash
set -e

cd /root/say-less

echo "Pulling latest code..."
git pull origin main

echo "Building and deploying..."
docker compose build --no-cache && docker compose up -d --force-recreate

echo "Say Less deployed successfully!"
