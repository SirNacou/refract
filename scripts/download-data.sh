#!/usr/bin/env bash
#
# Downloads required data files for the redirector service:
# - MaxMind GeoLite2-City database
# - UA Parser regexes.yaml
#
# Usage:
#   ./scripts/download-data.sh
#
# Environment variables:
#   GEOIP_LICENSE_KEY - MaxMind license key (required for GeoIP)
#

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DATA_DIR="$PROJECT_ROOT/services/redirector/data"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Create data directories
mkdir -p "$DATA_DIR/geoip"
mkdir -p "$DATA_DIR/user_agent"

# Download UA Parser regexes.yaml
download_ua_regexes() {
    local UA_REGEXES_URL="https://raw.githubusercontent.com/ua-parser/uap-core/master/regexes.yaml"
    local UA_REGEXES_PATH="$DATA_DIR/user_agent/regexes.yaml"

    if [[ -f "$UA_REGEXES_PATH" ]]; then
        log_info "UA regexes.yaml already exists, skipping..."
        return 0
    fi

    log_info "Downloading UA Parser regexes.yaml..."
    if curl -fsSL "$UA_REGEXES_URL" -o "$UA_REGEXES_PATH"; then
        log_info "Downloaded: $UA_REGEXES_PATH"
    else
        log_error "Failed to download UA regexes.yaml"
        return 1
    fi
}

# Download MaxMind GeoLite2-City database
download_geoip() {
    local GEOIP_PATH="$DATA_DIR/geoip/GeoLite2-City.mmdb"

    if [[ -f "$GEOIP_PATH" ]]; then
        log_info "GeoLite2-City.mmdb already exists, skipping..."
        return 0
    fi

    if [[ -z "${GEOIP_LICENSE_KEY:-}" ]]; then
        log_warn "GEOIP_LICENSE_KEY not set, skipping GeoIP download"
        log_warn "Get a free license key at: https://www.maxmind.com/en/geolite2/signup"
        log_warn "Then run: GEOIP_LICENSE_KEY=your_key ./scripts/download-data.sh"
        return 0
    fi

    log_info "Downloading MaxMind GeoLite2-City database..."
    local GEOIP_URL="https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City&license_key=${GEOIP_LICENSE_KEY}&suffix=tar.gz"
    local TMP_DIR=$(mktemp -d)

    if curl -fsSL "$GEOIP_URL" -o "$TMP_DIR/geoip.tar.gz"; then
        tar -xzf "$TMP_DIR/geoip.tar.gz" -C "$TMP_DIR"
        mv "$TMP_DIR"/GeoLite2-City_*/GeoLite2-City.mmdb "$GEOIP_PATH"
        rm -rf "$TMP_DIR"
        log_info "Downloaded: $GEOIP_PATH"
    else
        log_error "Failed to download GeoLite2-City database"
        log_error "Check your GEOIP_LICENSE_KEY is valid"
        rm -rf "$TMP_DIR"
        return 1
    fi
}

main() {
    log_info "Downloading data files for redirector service..."
    echo

    download_ua_regexes
    download_geoip

    echo
    log_info "Done!"
}

main "$@"
