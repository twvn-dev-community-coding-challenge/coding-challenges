from generated import charging_pb2, charging_pb2_grpc
from py_core.app import create_app
from py_core.server import grpc_lifespan

from grpc_server import ChargingGrpcServicer

app = create_app(
    title="Charging Service",
    description="Cost estimation and recording",
    lifespan=grpc_lifespan(
        servicers=[
            (charging_pb2_grpc.add_ChargingServiceServicer_to_server, ChargingGrpcServicer()),
        ],
        descriptors=[charging_pb2.DESCRIPTOR],
        bind="0.0.0.0:50052",
    ),
)
