"""Compile protos into generated/ with grpc_tools well-known types on the include path."""

from __future__ import annotations

import re
import subprocess
import sys
from pathlib import Path


def _patch_grpc_pb2_imports(generated_dir: Path) -> None:
    """Make *_pb2_grpc.py use package-relative imports for sibling *_pb2 modules."""
    for path in generated_dir.glob("*_pb2_grpc.py"):
        text = path.read_text(encoding="utf-8")
        new_text, count = re.subn(
            r"^import ([a-z][a-z0-9_]*_pb2) as ",
            r"from . import \1 as ",
            text,
            count=1,
            flags=re.MULTILINE,
        )
        if count:
            path.write_text(new_text, encoding="utf-8")


def main() -> None:
    root = Path(__file__).resolve().parent.parent
    protos_dir = root / "protos"
    out_dir = root / "generated"
    out_dir.mkdir(parents=True, exist_ok=True)

    import grpc_tools

    well_known = Path(grpc_tools.__file__).resolve().parent / "_proto"
    proto_files = [protos_dir / "provider.proto", protos_dir / "charging.proto"]

    cmd = [
        sys.executable,
        "-m",
        "grpc_tools.protoc",
        f"-I{protos_dir}",
        f"-I{well_known}",
        f"--python_out={out_dir}",
        f"--grpc_python_out={out_dir}",
        *[str(p) for p in proto_files],
    ]
    subprocess.run(cmd, check=True)
    _patch_grpc_pb2_imports(out_dir)


if __name__ == "__main__":
    main()
