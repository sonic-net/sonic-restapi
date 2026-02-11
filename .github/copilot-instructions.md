# Copilot Instructions for sonic-restapi

## Project Overview

sonic-restapi provides a RESTful API server for SONiC switches, exposing HTTPS endpoints for dynamic network configuration. The server (`go-server-server`) allows external systems to configure switch features like VxLAN tunnels, routes, and ARP entries programmatically.

## Architecture

```
sonic-restapi/
├── go-server-server/       # Go REST API server implementation
├── arp_responder/          # ARP responder module
├── arpthrift/              # ARP Thrift interface
├── arpthrifttest/          # ARP Thrift tests
├── Makefile                # Build system
├── Dockerfile              # Production Docker image
├── Dockerfile.build        # Build environment Docker image
├── Dockerfile.test         # Test Docker image
├── Dockerfile.*.prod       # Platform-specific production images
├── build.sh                # Build script
├── CreateMockPort.sh       # Test helper script
├── azurepipeline.yml       # CI pipeline
└── README.md
```

### Key Concepts
- **HTTPS REST API**: Exposes configuration endpoints over HTTPS
- **Redis backend**: Reads from and writes to SONiC Redis databases
- **Docker deployment**: Runs as a Docker container on SONiC switches
- **go-server-server**: The main Go binary implementing REST endpoints

## Language & Style

- **Primary languages**: Go (server), Python (tests), Shell (scripts)
- **Go conventions**: Standard Go formatting (`gofmt`), idiomatic error handling
- **Indentation**: Go: tabs (gofmt default), Python: 4 spaces
- **API design**: RESTful conventions — proper HTTP methods, status codes, JSON payloads
- **Naming**: Go exported names in `PascalCase`, unexported in `camelCase`

## Build Instructions

```bash
# Build using Docker
./build.sh

# This generates two Docker images:
# - rest-api-image: Production deployment
# - rest-api-image-test_local: Local testing with embedded Redis

# The production image is also saved as rest-api-image.gz
```

## Testing

```bash
# Run container locally with embedded Redis
docker run -d --rm -p8090:8090 -p6379:6379 \
  --name rest-api --cap-add NET_ADMIN --privileged \
  -t rest-api-image-test_local:latest

# Run tests
cd test
pytest -v

# Check logs
docker exec -it rest-api bash
cat /tmp/rest-api.err.log
```

### Deploying on a Switch
```bash
# Copy rest-api-image.gz to the switch
docker load < rest-api-image.gz
docker run -d -p=8090:8090/tcp \
  -v /var/run/redis/redis.sock:/var/run/redis/redis.sock \
  --name rest-api --cap-add NET_ADMIN --privileged \
  -t rest-api-image:latest
```

## PR Guidelines

- **Signed-off-by**: Required on all commits
- **CLA**: Sign Linux Foundation EasyCLA
- **Testing**: Include API tests for new endpoints
- **Security**: HTTPS endpoints must maintain proper authentication and authorization
- **CI**: Azure pipeline checks must pass

## Gotchas

- **TLS certificates**: HTTPS requires proper certificate configuration
- **Redis socket**: Production deployment mounts the SONiC Redis Unix socket
- **Network namespace**: Container needs `NET_ADMIN` capability
- **API versioning**: Maintain backward compatibility for existing API consumers
- **Concurrent access**: Handle concurrent Redis operations safely
- **Error responses**: Return proper HTTP status codes and error messages
