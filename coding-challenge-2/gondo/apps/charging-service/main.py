from contextlib import asynccontextmanager

from fastapi import FastAPI

from grpc_server import create_and_start_charging_grpc_server


@asynccontextmanager
async def lifespan(_app: FastAPI):
    grpc_server = await create_and_start_charging_grpc_server()
    yield
    await grpc_server.stop(5.0)


app = FastAPI(
    title="Charging Service",
    description="Cost estimation and recording",
    lifespan=lifespan,
)


@app.get("/health")
def health() -> dict[str, str]:
    return {"status": "ok"}
