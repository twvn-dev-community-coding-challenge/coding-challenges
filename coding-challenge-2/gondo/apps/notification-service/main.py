from fastapi import FastAPI
from google.protobuf.json_format import MessageToDict

from grpc_client import estimate_cost_via_grpc, select_provider_via_grpc

app = FastAPI(
    title="Notification Service",
    description="SMS lifecycle orchestration",
)


@app.get("/health")
def health() -> dict[str, str]:
    return {"status": "ok"}


@app.get("/test-grpc/provider")
async def test_grpc_provider() -> dict[str, object]:
    response = await select_provider_via_grpc()
    return MessageToDict(response, preserving_proto_field_name=True)


@app.get("/test-grpc/charging")
async def test_grpc_charging() -> dict[str, object]:
    response = await estimate_cost_via_grpc()
    return MessageToDict(response, preserving_proto_field_name=True)
