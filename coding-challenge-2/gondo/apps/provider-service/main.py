from contextlib import asynccontextmanager

from fastapi import FastAPI
from google.protobuf.json_format import MessageToDict

from grpc_client import estimate_cost_via_grpc
from grpc_server import create_and_start_provider_grpc_server


@asynccontextmanager
async def lifespan(_app: FastAPI):
    grpc_server = await create_and_start_provider_grpc_server()
    yield
    await grpc_server.stop(5.0)


app = FastAPI(
    title="Provider Service",
    description="Provider registry and routing rules",
    lifespan=lifespan,
)


@app.get("/health")
def health() -> dict[str, str]:
    return {"status": "ok"}


@app.get("/test-grpc/charging")
async def test_grpc_charging() -> dict[str, object]:
    response = await estimate_cost_via_grpc()
    return MessageToDict(response, preserving_proto_field_name=True)
