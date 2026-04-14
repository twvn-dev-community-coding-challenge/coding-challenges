"""Shared gRPC Server Reflection plugin for all services in the monorepo.

Enables grpcurl / grpcui to discover services at runtime.
Usage in any grpc_server.py:

    from reflection import enable_reflection
    enable_reflection(server, [my_pb2.DESCRIPTOR])
"""

from __future__ import annotations

from typing import Sequence

from google.protobuf import descriptor as _descriptor
from grpc import aio
from grpc_reflection.v1alpha import reflection


def enable_reflection(
    server: aio.Server,
    file_descriptors: Sequence[_descriptor.FileDescriptor],
) -> None:
    """Register gRPC Server Reflection on *server*.

    Args:
        server: An ``aio.Server`` instance (before ``start()``).
        file_descriptors: Proto ``DESCRIPTOR`` objects whose services
            should be discoverable (e.g. ``[charging_pb2.DESCRIPTOR]``).
    """
    service_names = [
        svc.full_name
        for fd in file_descriptors
        for svc in fd.services_by_name.values()
    ]
    reflection.enable_server_reflection(
        [*service_names, reflection.SERVICE_NAME],
        server,
    )
