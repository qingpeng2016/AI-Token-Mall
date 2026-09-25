#!/usr/bin/env python3
"""POST /liquidation/tasks/cancel — 取消强平。默认读同目录 payload.json。"""

from __future__ import annotations

import json
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))

from common.client import BeTrustClient
from common.config import API_KEY, BASE_URL


def main() -> int:
    payload_path = Path(__file__).resolve().parent / "payload.json"
    raw = json.loads(payload_path.read_text(encoding="utf-8"))
    liquidation_id = str(raw["liquidation_id"])
    cancel_reason = str(raw.get("cancel_reason", "user_repaid"))

    client = BeTrustClient(BASE_URL, API_KEY)
    status, body = client.cancel_task(liquidation_id, cancel_reason)
    print(f"HTTP {status}")
    print(json.dumps(body, ensure_ascii=False, indent=2))
    return 0 if status == 200 else 1


if __name__ == "__main__":
    raise SystemExit(main())
