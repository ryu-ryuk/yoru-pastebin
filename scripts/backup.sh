#!/bin/bash
# =============================================================================
# YORU PASTEBIN - DATABASE BACKUP SCRIPT
# =============================================================================
# This script creates database backups and uploads them to S3

set -euo pipefail

# Configuration
DB_HOST="${DB_HOST:-db}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${POSTGRES_DB}"
DB_USER="${POSTGRES_USER}"
BACKUP_DIR="/backups"
S3_BUCKET="${BACKUP_S3_BUCKET:-yoru-pastebin-backups}"
RETENTION_DAYS="${BACKUP_RETENTION_DAYS:-30}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}" >&2
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}" >&2
    exit 1
}

# Check if required tools are available
check_requirements() {
    log "Checking requirements..."
    
    if ! command -v pg_dump &> /dev/null; then
        error "pg_dump not found. Please install PostgreSQL client."
    fi
    
    if ! command -v aws &> /dev/null; then
        warn "AWS CLI not found. Installing..."
        apk add --no-cache aws-cli
    fi
    
    if [[ -z "${AWS_ACCESS_KEY_ID:-}" ]] || [[ -z "${AWS_SECRET_ACCESS_KEY:-}" ]]; then
        error "AWS credentials not configured. Set AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY."
    fi
}

# Create database backup
create_backup() {
    local timestamp=$(date +%Y%m%d_%H%M%S)
    local backup_file="${BACKUP_DIR}/yoru_backup_${timestamp}.sql"
    local compressed_file="${backup_file}.gz"
    
    log "Creating database backup..."
    
    # Create backup directory if it doesn't exist
    mkdir -p "${BACKUP_DIR}"
    
    # Set password for pg_dump
    export PGPASSWORD="${POSTGRES_PASSWORD}"
    
    # Create the backup
    if pg_dump -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -d "${DB_NAME}" \
        --verbose --clean --no-owner --no-privileges > "${backup_file}"; then
        log "Database backup created: ${backup_file}"
    else
        error "Failed to create database backup"
    fi
    
    # Compress the backup
    if gzip "${backup_file}"; then
        log "Backup compressed: ${compressed_file}"
        echo "${compressed_file}"
    else
        error "Failed to compress backup"
    fi
}

# Upload backup to S3
upload_to_s3() {
    local backup_file="$1"
    local s3_key="backups/$(basename "${backup_file}")"
    
    log "Uploading backup to S3: s3://${S3_BUCKET}/${s3_key}"
    
    if aws s3 cp "${backup_file}" "s3://${S3_BUCKET}/${s3_key}" \
        --storage-class STANDARD_IA \
        --metadata "created=$(date -Iseconds),retention=${RETENTION_DAYS}"; then
        log "Backup uploaded successfully"
    else
        error "Failed to upload backup to S3"
    fi
}

# Clean old local backups
cleanup_local() {
    log "Cleaning up local backups older than ${RETENTION_DAYS} days..."
    
    find "${BACKUP_DIR}" -name "yoru_backup_*.sql.gz" -mtime +${RETENTION_DAYS} -delete
    
    log "Local cleanup completed"
}

# Clean old S3 backups
cleanup_s3() {
    log "Cleaning up S3 backups older than ${RETENTION_DAYS} days..."
    
    local cutoff_date=$(date -d "${RETENTION_DAYS} days ago" +%Y-%m-%d)
    
    aws s3api list-objects-v2 \
        --bucket "${S3_BUCKET}" \
        --prefix "backups/" \
        --query "Contents[?LastModified<=\`${cutoff_date}\`].Key" \
        --output text | while read -r key; do
        if [[ -n "${key}" ]]; then
            log "Deleting old backup: s3://${S3_BUCKET}/${key}"
            aws s3 rm "s3://${S3_BUCKET}/${key}"
        fi
    done
    
    log "S3 cleanup completed"
}

# Health check function
health_check() {
    log "Performing health check..."
    
    # Check database connectivity
    export PGPASSWORD="${POSTGRES_PASSWORD}"
    if pg_isready -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}"; then
        log "Database is healthy"
    else
        error "Database health check failed"
    fi
    
    # Check S3 connectivity
    if aws s3 ls "s3://${S3_BUCKET}/" > /dev/null 2>&1; then
        log "S3 bucket is accessible"
    else
        error "S3 bucket is not accessible"
    fi
}

# Main execution
main() {
    log "Starting Yoru Pastebin backup process..."
    
    check_requirements
    health_check
    
    local backup_file
    backup_file=$(create_backup)
    
    upload_to_s3 "${backup_file}"
    
    cleanup_local
    cleanup_s3
    
    log "Backup process completed successfully!"
}

# Run main function
main "$@"
