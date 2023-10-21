bin := "nv"

# Run Aqua Security’s Trivy to catch possible vulnerabilities in the codebase
@audit:
  docker run -it --rm -v /var/run/docker.sock:/var/run/docker.sock -v {{justfile_directory()}}:/path aquasec/trivy fs --scanners config,secret,vuln /path

# Build nv’s executable
@build $GO_ENABLED="0": clean
  go mod download && go mod verify
  go build -ldflags="-w -s" -v -x -o {{bin}} ./cmd/nv

# Delete nv’s executable
@clean:
  rm -f {{bin}}

# Update dependencies
@go-mod-update:
  go get -d -u ./...
  go mod tidy

# Dry-run GoReleaser
@release-dry-run:
  goreleaser --snapshot --skip=publish --clean

# Launch the executable with optional arguments
@run *ARGS:
  go run ./cmd/nv/nv.go {{ARGS}}

# Git tag a version
@tag VERSION:
  git tag -a {{VERSION}} -s -m "{{VERSION}}"

# Run unit tests
@test:
  go test -v -cover  ./... | sed ''/PASS/s//$(printf "\033[32mPASS\033[0m")/'' | sed ''/FAIL/s//$(printf "\033[31mFAIL\033[0m")/''
