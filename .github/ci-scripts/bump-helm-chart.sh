#!/bin/bash

# Script to bump Helm chart version and appVersion in Chart.yaml
# Usage: ./bump-helm-chart.sh <version>
# Examples: ./bump-helm-chart.sh v1.2.3 or ./bump-helm-chart.sh 1.2.3

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_error() {
    echo -e "${RED}Error: $1${NC}" >&2
}

print_success() {
    echo -e "${GREEN}$1${NC}"
}

print_warning() {
    echo -e "${YELLOW}$1${NC}"
}

# Function to show usage
show_usage() {
    echo "Usage: $0 <version>"
    echo ""
    echo "Examples:"
    echo "  $0 v1.2.3"
    echo "  $0 1.2.3"
    echo ""
    echo "The script will update both 'version' and 'appVersion' in Chart.yaml"
    echo "  - version: semver without 'v' prefix (e.g., 1.2.3)"
    echo "  - appVersion: string with 'v' prefix (e.g., \"v1.2.3\")"
}

# Function to validate semver format
validate_semver() {
    local version="$1"
    # Check if version matches semver pattern (major.minor.patch with optional pre-release and build)
    if [[ ! "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.-]+)?(\+[a-zA-Z0-9.-]+)?$ ]]; then
        return 1
    fi
    return 0
}

# Function to parse version input
parse_version() {
    local input_version="$1"
    local version
    
    # Remove 'v' prefix if present
    if [[ "$input_version" =~ ^v(.+)$ ]]; then
        version="${BASH_REMATCH[1]}"
    else
        version="$input_version"
    fi
    
    # Validate the version format
    if ! validate_semver "$version"; then
        print_error "Invalid version format: '$input_version'. Expected format: v1.2.3 or 1.2.3"
        return 1
    fi
    
    echo "$version"
}

# Function to update Chart.yaml
update_chart_yaml() {
    local chart_file="$1"
    local version="$2"
    local app_version="v$version"
    
    # Check if Chart.yaml exists
    if [[ ! -f "$chart_file" ]]; then
        print_error "Chart.yaml not found at: $chart_file"
        return 1
    fi
    
    # Update version and appVersion using sed
    # Update version (semver without v)
    sed -i.tmp "s/^version: .*$/version: $version/" "$chart_file"
    
    # Update appVersion (string with v)
    sed -i.tmp "s/^appVersion: .*$/appVersion: \"$app_version\"/" "$chart_file"
    
    # Remove temporary file created by sed
    rm -f "${chart_file}.tmp"
    
    print_success "Updated Chart.yaml:"
    print_success "  version: $version"
    print_success "  appVersion: \"$app_version\""
}

# Main script logic
main() {
    # Check if version argument is provided
    if [[ $# -eq 0 ]]; then
        print_error "No version provided"
        show_usage
        exit 1
    fi
    
    # Check for help flag
    if [[ "$1" == "-h" || "$1" == "--help" ]]; then
        show_usage
        exit 0
    fi
    
    local input_version="$1"
    local chart_file="cmd/relayproxy/helm-charts/relay-proxy/Chart.yaml"
    
    # Parse and validate version
    local parsed_version
    if ! parsed_version=$(parse_version "$input_version"); then
        exit 1
    fi
    
    print_success "Parsed version: $parsed_version"
    
    # Update Chart.yaml
    if ! update_chart_yaml "$chart_file" "$parsed_version"; then
        exit 1
    fi
    
    print_success "Helm chart version bump completed successfully! 🦁"
}

# Run main function with all arguments
main "$@"
