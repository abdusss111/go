#!/bin/bash

# Migration script for Practice 3: Database Schema Migration
# This script provides easy commands for managing database migrations

MIGRATE_BIN="$HOME/go/bin/migrate"
MIGRATIONS_PATH="./internal/db/migrations"
DATABASE_URL="sqlite3://./expense.db"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# Check if migrate binary exists
if [ ! -f "$MIGRATE_BIN" ]; then
    print_error "golang-migrate binary not found at $MIGRATE_BIN"
    print_info "Please install it with: go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
    exit 1
fi

# Check if migrations directory exists
if [ ! -d "$MIGRATIONS_PATH" ]; then
    print_error "Migrations directory not found: $MIGRATIONS_PATH"
    exit 1
fi

case "$1" in
    "up")
        print_info "Applying all migrations..."
        $MIGRATE_BIN -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" up
        if [ $? -eq 0 ]; then
            print_status "All migrations applied successfully"
        else
            print_error "Failed to apply migrations"
            exit 1
        fi
        ;;
    "down")
        steps=${2:-1}
        print_info "Rolling back $steps migration(s)..."
        $MIGRATE_BIN -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" down "$steps"
        if [ $? -eq 0 ]; then
            print_status "Migration(s) rolled back successfully"
        else
            print_error "Failed to roll back migrations"
            exit 1
        fi
        ;;
    "version")
        print_info "Checking migration version..."
        $MIGRATE_BIN -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" version
        ;;
    "force")
        version=${2:-""}
        if [ -z "$version" ]; then
            print_error "Please specify a version number"
            echo "Usage: $0 force <version>"
            exit 1
        fi
        print_warning "Forcing migration version to $version..."
        $MIGRATE_BIN -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" force "$version"
        if [ $? -eq 0 ]; then
            print_status "Migration version forced to $version"
        else
            print_error "Failed to force migration version"
            exit 1
        fi
        ;;
    "reset")
        print_warning "Resetting database (removing expense.db)..."
        rm -f expense.db
        print_info "Applying all migrations..."
        $MIGRATE_BIN -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" up
        if [ $? -eq 0 ]; then
            print_status "Database reset and migrations applied successfully"
        else
            print_error "Failed to reset database"
            exit 1
        fi
        ;;
    "verify")
        print_info "Running schema verification..."
        go run cmd/verify/main.go
        ;;
    "help"|"--help"|"-h"|"")
        echo "Practice 3: Database Migration Management Script"
        echo ""
        echo "Usage: $0 <command> [options]"
        echo ""
        echo "Commands:"
        echo "  up                    Apply all pending migrations"
        echo "  down [steps]          Roll back migrations (default: 1 step)"
        echo "  version               Show current migration version"
        echo "  force <version>       Force migration version"
        echo "  reset                 Remove database and apply all migrations"
        echo "  verify                Run schema verification program"
        echo "  help                  Show this help message"
        echo ""
        echo "Examples:"
        echo "  $0 up                 # Apply all migrations"
        echo "  $0 down 2             # Roll back 2 migrations"
        echo "  $0 version            # Check current version"
        echo "  $0 force 3            # Force version to 3"
        echo "  $0 reset              # Reset and reapply all migrations"
        echo "  $0 verify             # Verify schema integrity"
        ;;
    *)
        print_error "Unknown command: $1"
        echo "Run '$0 help' for usage information"
        exit 1
        ;;
esac
