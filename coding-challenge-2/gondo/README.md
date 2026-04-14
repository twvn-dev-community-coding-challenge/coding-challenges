# Go & Do - Membership Registration

## Team Members

- Hien Huynh (Team Lead / Kaluza Pod 4)
- Thien Nguyen (Developer / Kaluza Pod 3)

## Architecture

NX monorepo with 3 Python/FastAPI backend microservices and a React frontend.

| App | Type | HTTP | gRPC | Description |
|-----|------|------|------|-------------|
| `notification-service` | Python / FastAPI | 8001 | — | SMS lifecycle orchestration |
| `provider-service` | Python / FastAPI | 8002 | 50051 | Provider registry and routing rules |
| `charging-service` | Python / FastAPI | 8003 | 50052 | Cost estimation and recording |
| `frontend` | React / Vite | 4200 | — | Web UI |

## Prerequisites

- Node.js 20 (pinned via `.nvmrc` — run `nvm use` to activate)
- Python 3.13+
- Yarn 1.22+
- (optional) `grpcurl` / `grpcui` for exploring gRPC contracts — `brew install grpcurl grpcui`

## How to Run

### 1. Install JS dependencies

```bash
nvm use        # uses .nvmrc to select Node 20
yarn install
```

### 2. Set up Python virtual environment

```bash
python3 -m venv .venv
source .venv/bin/activate
pip install -r apps/notification-service/requirements.txt
```

All three Python services share the same dependency set.

### 3. Start all services (recommended)

```bash
yarn start
# or directly:
bash scripts/start-all.sh
```

This starts charging → provider → notification in the correct order with health checks. Press Ctrl+C to stop all.

### Start services individually

```bash
yarn nx run charging-service:serve      # port 8003 + gRPC 50052
yarn nx run provider-service:serve      # port 8002 + gRPC 50051
yarn nx run notification-service:serve  # port 8001 (main entry)
```

To stop all services:

```bash
yarn stop
```

### Frontend (port 4200)

```bash
yarn nx run frontend:serve
```

### 4. Run tests

```bash
# All projects
yarn nx run-many -t test

# Individual
yarn nx run notification-service:test
yarn nx run frontend:test
```

### 5. Lint

```bash
# All projects
yarn nx run-many -t lint

# Individual
yarn nx run notification-service:lint
yarn nx run frontend:lint
```

### 6. Build

```bash
# All projects
yarn nx run-many -t build

# Individual
yarn nx run frontend:build
```

### 7. Generate OpenAPI contract

```bash
yarn nx run notification-service:generate-openapi
# Outputs: apps/notification-service/openapi.json
```

### 8. Useful NX commands

```bash
yarn nx show projects          # list all projects
yarn nx graph                  # visualize project graph
yarn nx affected -t test,lint  # run only affected targets
```

## Docker (E2E)

Run all backend services in containers:

```bash
yarn docker:up       # build and start all services
yarn docker:logs     # tail logs from all containers
yarn docker:down     # stop and remove containers
```

Once running, all endpoints are available at the same ports:

- Notification: http://localhost:8001/docs
- Provider: http://localhost:8002/docs
- Charging: http://localhost:8003/docs

Test E2E connectivity:

```bash
curl http://localhost:8001/health
```

## gRPC Contracts

Internal service-to-service communication uses gRPC. Proto definitions live in `libs/grpc-contracts/protos/`.

| Proto | Service | RPCs | Consumers |
|-------|---------|------|-----------|
| `provider.proto` | `ProviderService` (:50051) | `ResolveRouting`, `SelectProvider` | notification-service |
| `charging.proto` | `ChargingService` (:50052) | `EstimateCost`, `EstimateCostBatch`, `RecordActualCost` | — |

### Reviewing contracts

**Read the proto files directly:**

```bash
cat libs/grpc-contracts/protos/provider.proto
cat libs/grpc-contracts/protos/charging.proto
```

### gRPC Reflection (Swagger-like UI for gRPC)

All gRPC servers enable [Server Reflection](https://grpc.io/docs/guides/reflection/) via the shared `libs/grpc-contracts/reflection.py` plugin. This lets tools like `grpcui` and `grpcurl` discover services at runtime — no proto files needed on the client side.

**grpcui — interactive web UI (like Swagger for gRPC):**

```bash
brew install grpcui

yarn start   # services must be running

grpcui -plaintext localhost:50051   # Provider Service
grpcui -plaintext localhost:50052   # Charging Service
```

This opens a browser with forms for every RPC, input fields for each message field, and live responses.

**grpcurl — CLI client (like curl for gRPC):**

```bash
brew install grpcurl

# List services
grpcurl -plaintext localhost:50051 list
grpcurl -plaintext localhost:50052 list

# Describe a service
grpcurl -plaintext localhost:50051 describe gondo.provider.ProviderService
grpcurl -plaintext localhost:50052 describe gondo.charging.ChargingService

# SelectProvider (provider-service :50051)
grpcurl -plaintext -d '{
  "country_code": "VN",
  "carrier": "VIETTEL",
  "as_of": "2026-04-11T00:00:00Z",
  "policy_context": {"policy": "lowest_cost_healthy", "message_id": "msg_001"}
}' localhost:50051 gondo.provider.ProviderService/SelectProvider

# EstimateCost (charging-service :50052)
grpcurl -plaintext -d '{
  "message_id": "msg_001",
  "provider_id": "prv_02",
  "country_code": "VN",
  "carrier": "VIETTEL",
  "as_of": "2026-04-11T00:00:00Z"
}' localhost:50052 gondo.charging.ChargingService/EstimateCost
```

**Adding reflection to a new gRPC service:**

```python
from reflection import enable_reflection
from generated import my_pb2

server = aio.server()
add_MyServiceServicer_to_server(MyServicer(), server)
enable_reflection(server, [my_pb2.DESCRIPTOR])  # one line
```

### Regenerating stubs

After editing any `.proto` file:

```bash
yarn nx run grpc-contracts:generate
```

Generated Python stubs are written to `libs/grpc-contracts/generated/`.

## Design Decisions

- **NX monorepo** for unified task orchestration across Python and JS projects
- **FastAPI + OpenAPI** for public HTTP contracts; gRPC planned for internal service calls
- **Vite** as the React bundler for fast dev experience
- **`nx:run-commands`** executor for Python services (no native NX Python plugin)

## Challenges Faced

[What was hard? How did you overcome it?]

## What We Learned

[New skills, technologies, or insights]

## With More Time, We Would...

[Nice-to-haves you didn't implement]

## AI Tools Used (if any)

[Which tools? How did they help?]
