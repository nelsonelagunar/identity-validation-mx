#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
KUBECTL_DIR="${SCRIPT_DIR}/../deployments/kubernetes"

NAMESPACE="identity-validation"
DRY_RUN="${DRY_RUN:-false}"
KUBE_CONTEXT="${KUBE_CONTEXT:-}"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

error() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] ERROR: $1" >&2
    exit 1
}

check_prerequisites() {
    log "Checking prerequisites..."
    
    if ! command -v kubectl &> /dev/null; then
        error "kubectl is not installed"
    fi
    
    if ! command -v docker &> /dev/null; then
        error "docker is not installed"
    fi
    
    if [ -n "$KUBE_CONTEXT" ]; then
        kubectl config use-context "$KUBE_CONTEXT" || error "Failed to switch to context $KUBE_CONTEXT"
    fi
    
    log "Prerequisites check passed"
}

build_image() {
    local tag="${1:-latest}"
    log "Building Docker image with tag: $tag"
    docker build -t identity-validation-mx:"$tag" -f "${SCRIPT_DIR}/../deployments/docker/Dockerfile" "${SCRIPT_DIR}/.."
    log "Docker image built successfully"
}

push_image() {
    local tag="${1:-latest}"
    local registry="${REGISTRY:-docker.io}"
    log "Pushing Docker image to registry: $registry"
    docker tag identity-validation-mx:"$tag" "$registry/identity-validation-mx:$tag"
    docker push "$registry/identity-validation-mx:$tag"
    log "Docker image pushed successfully"
}

deploy_namespace() {
    log "Deploying namespace..."
    kubectl apply -f "${KUBECTL_DIR}/namespace.yaml"
}

deploy_configmaps() {
    log "Deploying ConfigMaps..."
    kubectl apply -f "${KUBECTL_DIR}/configmap.yaml"
}

deploy_secrets() {
    log "Deploying Secrets..."
    
    if kubectl get secret identity-validation-secrets -n "$NAMESPACE" &> /dev/null; then
        log "Secrets already exist, skipping creation"
    else
        kubectl apply -f "${KUBECTL_DIR}/secrets.yaml"
    fi
}

deploy_postgres() {
    log "Deploying PostgreSQL..."
    kubectl apply -f "${KUBECTL_DIR}/postgres.yaml"
    
    if [ "$DRY_RUN" = "false" ]; then
        log "Waiting for PostgreSQL to be ready..."
        kubectl rollout status statefulset/postgres -n "$NAMESPACE" --timeout=300s
    fi
}

deploy_redis() {
    log "Deploying Redis..."
    kubectl apply -f "${KUBECTL_DIR}/redis.yaml"
    
    if [ "$DRRY_RUN" = "false" ]; then
        log "Waiting for Redis to be ready..."
        kubectl rollout status deployment/redis -n "$NAMESPACE" --timeout=120s
    fi
}

deploy_api() {
    log "Deploying API..."
    kubectl apply -f "${KUBECTL_DIR}/deployment.yaml"
    kubectl apply -f "${KUBECTL_DIR}/service.yaml"
    kubectl apply -f "${KUBECTL_DIR}/hpa.yaml"
    
    if [ "$DRY_RUN" = "false" ]; then
        log "Waiting for API deployment to be ready..."
        kubectl rollout status deployment/identity-validation-api -n "$NAMESPACE" --timeout=180s
    fi
}

deploy_ingress() {
    log "Deploying Ingress..."
    kubectl apply -f "${KUBECTL_DIR}/ingress.yaml"
}

verify_deployment() {
    log "Verifying deployment..."
    
    log "Pods:"
    kubectl get pods -n "$NAMESPACE"
    
    log "Services:"
    kubectl get services -n "$NAMESPACE"
    
    log "Deployments:"
    kubectl get deployments -n "$NAMESPACE"
    
    log "StatefulSets:"
    kubectl get statefulsets -n "$NAMESPACE"
    
    log "HPA status:"
    kubectl get hpa -n "$NAMESPACE"
}

deploy() {
    log "Starting deployment..."
    
    check_prerequisites
    
    if [ "$DRY_RUN" = "true" ]; then
        log "DRY RUN MODE - Changes will not be applied"
    fi
    
    deploy_namespace
    deploy_configmaps
    deploy_secrets
    deploy_postgres
    deploy_redis
    deploy_api
    deploy_ingress
    
    if [ "$DRY_RUN" = "false" ]; then
        verify_deployment
    fi
    
    log "Deployment completed successfully!"
}

usage() {
    echo "Usage: $0 [command]"
    echo "Commands:"
    echo "  deploy          Full deployment of all resources"
    echo "  build [tag]     Build Docker image"
    echo "  push [tag]      Build and push Docker image to registry"
    echo "  api            Deploy API only"
    echo "  postgres       Deploy PostgreSQL only"
    echo "  redis          Deploy Redis only"
    echo ""
    echo "Environment variables:"
    echo "  DRY_RUN         Set to 'true' for dry run mode"
    echo "  KUBE_CONTEXT    Kubernetes context to use"
    echo "  REGISTRY        Docker registry (default: docker.io)"
}

case "${1:-deploy}" in
    deploy)
        deploy
        ;;
    build)
        build_image "${2:-latest}"
        ;;
    push)
        build_image "${2:-latest}"
        push_image "${2:-latest}"
        ;;
    api)
        check_prerequisites
        deploy_api
        ;;
    postgres)
        check_prerequisites
        deploy_postgres
        ;;
    redis)
        check_prerequisites
        deploy_redis
        ;;
    *)
        usage
        exit 1
        ;;
esac