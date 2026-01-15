# High-level justfile for wsh workspace
# Run with: just <recipe>

# Show all available recipes
default:
    @just --list

# Build warg CLI and install to bin
build_warg:
    go build -C ./warg/cmd/warg -o ../../../bin/warg/warg

# Run all warg tests (delegates to warg/justfile)
test_warg:
    cd warg && just test-all

# Build all projects
build_all:
    @just build_warg

# Clean all build artifacts
clean:
    rm -rf bin/
    cd warg && just clean
