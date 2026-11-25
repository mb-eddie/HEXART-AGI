#!/usr/bin/env python3

import importlib.util
import os
import sys


MODULE_PATH = "/usr/share/hexstrike-ai/hexstrike_mcp.py"


def load_hexstrike():
    spec = importlib.util.spec_from_file_location("hexstrike_mcp", MODULE_PATH)
    if spec is None or spec.loader is None:
        raise RuntimeError("cannot load HexStrike MCP module")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def main():
    module = load_hexstrike()
    server_url = os.environ.get("HEXSTRIKE_SERVER_URL", "http://127.0.0.1:8888")
    timeout = int(os.environ.get("HEXSTRIKE_REQUEST_TIMEOUT", "300"))
    host = os.environ.get("HEXSTRIKE_MCP_HOST", "0.0.0.0")
    port = int(os.environ.get("HEXSTRIKE_MCP_PORT", "8890"))

    client = module.HexStrikeClient(server_url, timeout)
    health = client.check_health()
    if "error" in health:
        print(
            "WARNING: HexStrike API is unavailable at "
            f"{server_url}: {health['error']}",
            file=sys.stderr,
        )
    else:
        print(
            f"HexStrike API connected at {server_url}: "
            f"{health.get('status', 'unknown')}",
            file=sys.stderr,
        )

    mcp = module.setup_mcp_server(client)
    mcp.settings.host = host
    mcp.settings.port = port
    mcp.settings.streamable_http_path = "/mcp"
    if getattr(mcp.settings, "transport_security", None) is not None:
        mcp.settings.transport_security.enable_dns_rebinding_protection = False

    print(
        f"Starting HexStrike MCP Streamable HTTP on {host}:{port}/mcp",
        file=sys.stderr,
    )
    mcp.run(transport="streamable-http")


if __name__ == "__main__":
    main()
