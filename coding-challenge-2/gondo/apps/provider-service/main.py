from generated import provider_pb2, provider_pb2_grpc
from grpc_server import ProviderGrpcServicer
from py_core.app import create_app
from py_core.server import grpc_lifespan

app = create_app(
    title="Provider Service",
    description="Provider registry and routing rules",
    lifespan=grpc_lifespan(
        servicers=[
            (provider_pb2_grpc.add_ProviderServiceServicer_to_server, ProviderGrpcServicer()),
        ],
        descriptors=[provider_pb2.DESCRIPTOR],
        bind="0.0.0.0:50051",
    ),
)
