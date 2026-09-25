#!/usr/bin/env python3
"""GET /liquidation/tasks — 查询强平任务状态与成交汇总。

默认读同目录 payload.json（只需 liquidation_id；timestamp/sign 由脚本自动生成）。

成功时 data 字段见 docs/api.md §3.2，主要包括：
  status, symbol, total_qty, total_amount, total_fee, total_net_proceeds_amount,
  avg_price, remaining_debt_amount, remaining_collateral_qty, fills 等。
"""

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
    liquidation_id = str(raw["liquidation_id"]).strip()
    if not liquidation_id:
        print("payload.json 缺少 liquidation_id", file=sys.stderr)
        return 2

    client = BeTrustClient(BASE_URL, API_KEY)
    status, body = client.get_task(liquidation_id)
    print(f"HTTP {status}")
    print(json.dumps(body, ensure_ascii=False, indent=2))
    if status == 401 and isinstance(body, dict):
        msg = body.get("message", "")
        if msg == "invalid sign":
            print(
                "\n提示: 检查 python/common/config.py 与库 sign_key 是否一致。",
                file=sys.stderr,
            )
    return 0 if status == 200 else 1


if __name__ == "__main__":
    raise SystemExit(main())
