#!/bin/bash
set -e

REGISTRY="ghcr.io/sirnacou/refract"
TAG="${1:-latest}"

echo "Building and pushing Refract images to GHCR with tag: $TAG"
echo ""

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

build_and_push() {
    local service=$1
    local dockerfile=$2
    local context=$3
    
    echo -e "${BLUE}Building $service...${NC}"
    docker build -f "$dockerfile" -t "$REGISTRY/$service:$TAG" -t "$REGISTRY/$service:latest" "$context"
    
    echo -e "${GREEN}✓ Pushing $service to GHCR...${NC}"
    docker push "$REGISTRY/$service:$TAG"
    docker push "$REGISTRY/$service:latest"
    echo ""
}

# Build and push each service
build_and_push "migrations" "Dockerfile.migrations" "."
build_and_push "migrations-drizzle" "Dockerfile.migrations.drizzle" "."
build_and_push "frontend" "frontend/Dockerfile" "."
build_and_push "api" "api/Dockerfile.api" "."
build_and_push "redirector" "api/Dockerfile.redirector" "."
build_and_push "worker" "api/Dockerfile.worker" "."

echo -e "${GREEN}✓ All images built and pushed successfully!${NC}"
echo ""
echo "Images pushed:"
docker images | grep "$REGISTRY" | awk '{print "  " $1 ":" $2}'
