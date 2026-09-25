"""BeTrust 联调 HTTP 客户端。"""

from __future__ import annotations

import json
import time
import urllib.error
import urllib.parse
import urllib.request
from typing import Any

from common.betrust_sign import sign_json_body, sign_query_params


class BeTrustClient:
    def __init__(self, base_url: str, api_key: str) -> None:
        self.base_url = base_url.rstrip("/")
        self.api_key = api_key

    def _request_json(
        self,
        method: str,
        path: str,
        *,
        body: bytes | None = None,
        query: dict[str, str] | None = None,
    ) -> tuple[int, Any]:
        url = f"{self.base_url}{path}"
        if query:
            url = f"{url}?{urllib.parse.urlencode(query)}"
        req = urllib.request.Request(url, data=body, method=method)
        if body is not None:
            req.add_header("Content-Type", "application/json")
        try:
            with urllib.request.urlopen(req, timeout=60) as resp:
                raw = resp.read().decode("utf-8")
                return resp.status, json.loads(raw) if raw else None
        except urllib.error.HTTPError as e:
            raw = e.read().decode("utf-8")
            try:
                return e.code, json.loads(raw)
            except json.JSONDecodeError:
                return e.code, raw
        except urllib.error.URLError as e:
            if isinstance(e.reason, ConnectionRefusedError) or "Connection refused" in str(e.reason):
                raise SystemExit(
                    f"无法连接 {url}\n"
                    "请先启动 Mojo HTTP 服务，例如在项目根目录：\n"
                    "  go run main.go --run_conf=local --run_bot=false\n"
                    f"（默认端口见 python/common/config.py BASE_URL={self.base_url!r}）"
                ) from e
            raise

    def create_task(self, payload: dict[str, Any]) -> tuple[int, Any]:
        payload = dict(payload)
        payload.setdefault("timestamp", int(time.time() * 1000))
        body = sign_json_body(self.api_key, payload)
        return self._request_json("POST", "/liquidation/tasks", body=body)

    def get_task(self, liquidation_id: str, timestamp: int | None = None) -> tuple[int, Any]:
        ts = timestamp if timestamp is not None else int(time.time() * 1000)
        query = sign_query_params(
            self.api_key,
            {"liquidation_id": liquidation_id, "timestamp": ts},
        )
        return self._request_json("GET", "/liquidation/tasks", query=query)

    def cancel_task(
        self,
        liquidation_id: str,
        cancel_reason: str,
        *,
        requested_at: int | None = None,
        timestamp: int | None = None,
    ) -> tuple[int, Any]:
        now = int(time.time() * 1000)
        payload = {
            "liquidation_id": liquidation_id,
            "timestamp": timestamp if timestamp is not None else now,
            "cancel_reason": cancel_reason,
            "requested_at": requested_at if requested_at is not None else now,
        }
        body = sign_json_body(self.api_key, payload)
        return self._request_json("POST", "/liquidation/tasks/cancel", body=body)
