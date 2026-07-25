#!/usr/bin/env python3
"""Generate Python gRPC stubs from shared protobuf definitions.

Run from the document-extractor service directory:

    cd src/services/document-extractor
    python ../../scripts/generate_protos.py

Or from the repo root:

    python src/scripts/generate_protos.py
"""

from __future__ import annotations

import os
import sys
from pathlib import Path


def main() -> None:
    """Generate Python gRPC stubs in the document-extractor service."""
    # Determine script location
    script_dir = Path(__file__).resolve().parent

    # Navigate to repo root (scripts/ is one level down from src/ or at root)
    if script_dir.name == "scripts" and (script_dir.parent / "src").exists():
        repo_root = script_dir.parent
    else:
        repo_root = script_dir

    proto_dir = repo_root / "src" / "services" / "shared" / "proto"
    output_dir = (
        repo_root / "src" / "services" / "document-extractor" / "grpc_client" / "proto"
    )

    if not proto_dir.exists():
        print(f"Proto directory not found: {proto_dir}")
        sys.exit(1)

    # Ensure output directory exists
    os.makedirs(output_dir, exist_ok=True)

    # Collect all proto files
    proto_files = []
    for root, _dirs, files in os.walk(proto_dir):
        for f in files:
            if f.endswith(".proto"):
                proto_files.append(os.path.join(root, f))

    if not proto_files:
        print(f"No .proto files found in {proto_dir}")
        sys.exit(1)

    # Generate using grpc_tools.protoc
    try:
        from grpc_tools import protoc
    except ImportError:
        print("grpcio-tools not installed. Run: pip install grpcio-tools")
        print("Or: pip install -e src/services/document-extractor")
        sys.exit(1)

    args = [sys.argv[0], f"-I{proto_dir}", f"--python_out={output_dir}", f"--grpc_python_out={output_dir}"]
    args += proto_files

    print(f"Generating Python gRPC stubs from {proto_dir}")
    print(f"Output: {output_dir}")
    print(f"Proto files: {len(proto_files)}")
    print()

    result = protoc.main(args)
    if result != 0:
        print(f"Proto generation failed with exit code {result}")
        sys.exit(1)

    # Fix imports in generated files to use absolute package paths
    print("Fixing proto imports...")
    for root, _dirs, files in os.walk(output_dir):
        for f in files:
            if f.endswith(".py"):
                path = os.path.join(root, f)
                with open(path) as fh:
                    content = fh.read()

                # Fix imports to use absolute package paths
                content = content.replace(
                    "from common.v1", "from grpc_client.proto.common.v1"
                ).replace(
                    "from ontology.v1", "from grpc_client.proto.ontology.v1"
                )

                with open(path, "w") as fh:
                    fh.write(content)

    # Create __init__.py files for all subdirectories
    print("Creating __init__.py files...")
    for root, _dirs, files in os.walk(output_dir):
        init_path = os.path.join(root, "__init__.py")
        if not os.path.exists(init_path):
            with open(init_path, "w"):
                pass

    print()
    print("Done! gRPC stubs generated in:", output_dir)
    print()
    print("The document-extractor gRPC client will use these stubs")
    print("to communicate with the ontology-service via ApplySequence.")


if __name__ == "__main__":
    main()
