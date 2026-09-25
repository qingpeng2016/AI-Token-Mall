#!/usr/bin/env python3
"""POST /liquidation/tasks — 触发强平。默认读同目录 payload.json。"""

from __future__ import annotations

import json
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))

from common.client import BeTrustClient
from common.config import API_KEY, BASE_URL

CREATE_FIELDS = (
    "liquidation_id",
    "loan_order_id",
    "uid",
    "original_ratio",
    "trigger_ratio",
    "trigger_mark_price",
    "trigger_collateral_asset",
    "trigger_collateral_qty",
    "trigger_collateral_due_amount",
    "symbol",
)


def main() -> int:
    payload_path = Path(__file__).resolve().parent / "payload.json"
    raw = json.loads(payload_path.read_text(encoding="utf-8"))
    payload = {k: str(raw[k]) for k in CREATE_FIELDS if k in raw}

    client = BeTrustClient(BASE_URL, API_KEY)
    status, body = client.create_task(payload)
    print(f"HTTP {status}")
    print(json.dumps(body, ensure_ascii=False, indent=2))
    if status == 401 and isinstance(body, dict):
        msg = body.get("message", "")
        if msg == "invalid sign":
            print(
                "\n提示: invalid sign 多为 Python config.API_KEY 与库 liquidation_platform_config.sign_key 不一致；"
                "改库后请重启 go 服务（local 环境已禁用 30s 缓存）。",
                file=sys.stderr,
            )
        elif msg == "request expired":
            print("\n提示: 本机与服务器时间差超过 10 秒。", file=sys.stderr)
    return 0 if status == 200 else 1


if __name__ == "__main__":
    raise SystemExit(main())
