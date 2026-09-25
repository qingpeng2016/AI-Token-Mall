"""HMAC-SHA256 sign，与 common/betrustsign 及 docs/api.md 一致。"""

from __future__ import annotations

import hashlib
import hmac
import json
from typing import Any, Mapping


def canonical_value(v: Any) -> str:
    if v is None:
        return ""
    if isinstance(v, bool):
        return "true" if v else "false"
    if isinstance(v, int):
        return str(v)
    if isinstance(v, float):
        if v == int(v):
            return str(int(v))
        s = f"{v:.15f}".rstrip("0").rstrip(".")
        return s
    return str(v)


def build_sign_base(params: Mapping[str, Any]) -> str:
    keys = sorted(k for k in params if k != "sign")
    parts = [f"{k}={canonical_value(params[k])}" for k in keys]
    return "&".join(parts)


def compute_sign(secret: str, params: Mapping[str, Any]) -> str:
    base = build_sign_base(params)
    mac = hmac.new(secret.encode("utf-8"), base.encode("utf-8"), hashlib.sha256)
    return mac.hexdigest()


def attach_sign(secret: str, body: dict[str, Any]) -> dict[str, Any]:
    out = dict(body)
    out.pop("sign", None)
    out["sign"] = compute_sign(secret, out)
    return out


def sign_query_params(secret: str, params: Mapping[str, Any]) -> dict[str, str]:
    base = {k: canonical_value(v) for k, v in params.items() if k != "sign"}
    base["sign"] = compute_sign(secret, base)
    return base


def sign_json_body(secret: str, body: Mapping[str, Any]) -> bytes:
    signed = attach_sign(secret, dict(body))
    return json.dumps(signed, separators=(",", ":"), ensure_ascii=False).encode("utf-8")
