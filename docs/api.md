# 质押清算 HTTP 接口说明

| 项 | 说明 |
| --- | --- |
| 文档版本 | 1.0 |
| 提供方 | Mojo 清算执行服务 |
| Base URL | 联调/生产环境由 Mojo 提供（HTTPS） |
| Content-Type | `application/json`（POST 请求体） |

---

## 一、通用约定

### 1.1 成功响应

HTTP 状态码一般为 **200**：

```json
{
  "code": 200,
  "message": "success",
  "data": {}
}
```

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| code | number | 业务码，200 表示成功 |
| message | string | 提示文案 |
| data | object | 业务数据 |

### 1.2 鉴权失败

HTTP **401**：

```json
{
  "code": 401,
  "message": "invalid sign",
  "data": null
}
```

| message | 含义 |
| --- | --- |
| invalid json | 请求体非合法 JSON |
| missing sign | 缺少 sign |
| invalid timestamp | timestamp 非整数 |
| request expired | 请求时间与服务器偏差超过 10 秒 |
| invalid sign | 签名不匹配 |

---

## 二、鉴权（sign + timestamp）

### 2.1 适用接口

| 方法 | 路径 | 调用方向 |
| --- | --- | --- |
| POST | `/liquidation/tasks` | BeTrust → Mojo |
| GET | `/liquidation/tasks` | BeTrust → Mojo |
| POST | `/liquidation/tasks/cancel` | BeTrust → Mojo |
| POST | BeTrust 提供的清算结果回调 URL | Mojo → BeTrust |

### 2.2 密钥与回调地址

- **sign_key**：存于 `liquidation_platform_config.sign_key`，BeTrust 与 Mojo 按环境线下约定，用于 HMAC 签名；不在 HTTP 请求中传输。
- **回调 URL**：BeTrust 提供清算终态接收地址，在环境联调时交给 Mojo 配置；**不在**触发/查询/取消接口中传递。

### 2.3 签名算法

1. 收集参与签名的参数  
   - **POST**：JSON 请求体顶层字段  
   - **GET**：URL Query 全部参数（含 `liquidation_id`、`timestamp` 等）
2. 字段 **sign 不参与**拼接
3. 键名按 ASCII 升序，拼接 `key1=value1&key2=value2&...`  
   - 整数：十进制字符串  
   - 字符串：原样
4. `sign = hex_lower(HMAC-SHA256(sign_key, 待签名字符串))`（64 位小写 hex）
5. **timestamp** 为毫秒 Unix 时间，与接收方服务器偏差不超过 **10 秒**

### 2.4 签名示例（POST）

```json
{
  "liquidation_id": "LQ-20260915-000001",
  "loan_order_id": "LO-987654",
  "timestamp": 1757923200123
}
```

待签名字符串示例：

```text
liquidation_id=LQ-20260915-000001&loan_order_id=LO-987654&timestamp=1757923200123
```

计算 sign 后写入请求再发送。

---

## 三、BeTrust 调用 Mojo

### 3.1 触发强平任务

**POST** `/liquidation/tasks`

| 说明 | |
| --- | --- |
| 鉴权 | sign + timestamp |
| 幂等 | 相同 `liquidation_id` 重复请求按业务规则返回 accepted 或说明已存在 |

**请求体**

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| timestamp | int64 | 是 | 毫秒时间戳 |
| sign | string | 是 | 签名 |
| liquidation_id | string | 是 | 强平任务 ID，生成的全局唯一 ID |
| loan_order_id | string | 是 | 借据 ID，用户在 BeTrust 质押记录的唯一 ID |
| uid | string | 是 | 用户 ID |
| original_ratio | string | 是 | 原始质押率 |
| trigger_ratio | string | 是 | 触发时质押率 |
| trigger_mark_price | string | 是 | 触发时标记价格 |
| trigger_collateral_asset | string | 是 | 触发时质押资产，如 BTC |
| trigger_collateral_qty | string | 是 | 触发时质押数量 |
| trigger_collateral_due_amount | string | 是 | 触发时应还总额 |
| symbol | string | 是 | 现货交易对，如 BTCUSDT |

**请求示例**

```json
{
  "timestamp": 1757923200123,
  "sign": "a1b2c3...64位hex...",
  "liquidation_id": "LQ-20260915-000001",
  "loan_order_id": "LO-987654",
  "uid": "U10001",
  "original_ratio": "0.75",
  "trigger_ratio": "0.85",
  "trigger_mark_price": "65000.12",
  "trigger_collateral_asset": "BTC",
  "trigger_collateral_qty": "1.5",
  "trigger_collateral_due_amount": "97500.00",
  "symbol": "BTCUSDT"
}
```

**响应 data**

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| result | string | 受理结果，见下表 |
| result_message | string | 结果说明（如首次受理、幂等已存在、拒绝原因等） |

**`result` 取值**

| 取值 | 含义 |
| --- | --- |
| accepted | 请求被接受（新建成功，或相同 `liquidation_id` 幂等命中） |
| rejected | 请求被拒绝（参数或业务校验未通过） |

**典型 `result` 与 `result_message`**

| result | result_message | 含义 |
| --- | --- | --- |
| accepted | 强平任务已受理 | 首次创建成功 |
| accepted | 强平任务已存在 | 相同 `liquidation_id` 幂等 |
| rejected | 该借款订单已有进行中的强平任务 | 同借据另有进行中任务 |
| rejected | （各字段校验文案） | 参数无效或平台配置问题 |

**响应示例（首次受理）**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "result": "accepted",
    "result_message": "强平任务已受理"
  }
}
```

**响应示例（幂等已存在）**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "result": "accepted",
    "result_message": "强平任务已存在"
  }
}
```

---

### 3.2 查询强平任务

**GET** `/liquidation/tasks`

**Query 参数**

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| liquidation_id | string | 是 | 强平任务 ID，生成的全局唯一 ID |
| timestamp | int64 | 是 | 毫秒时间戳 |
| sign | string | 是 | 签名 |

**请求示例**

```http
GET /liquidation/tasks?liquidation_id=LQ-20260915-000001&timestamp=1757923200123&sign=a1b2c3...64位hex...
```

**响应 data（顶层）**

下列字段在成功响应的 `data` 中**均必返**（无成交时 `avg_price` 为 `"0"`）。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| liquidation_id | string | 强平任务 ID，生成的全局唯一 ID |
| loan_order_id | string | 借据 ID，用户在 BeTrust 质押记录的唯一 ID |
| status | string | 任务状态，见下表 |
| symbol | string | 交易对 |
| total_qty | string | 累计成交数量 |
| total_amount | string | 累计成交金额 |
| total_fee | string | 累计手续费 |
| total_net_proceeds_amount | string | 净所得金额（累计成交金额扣减累计手续费后的净值，用于冲抵债务等） |
| avg_price | string | 成交均价；尚无成交时为 `"0"` |
| remaining_debt_amount | string | 剩余债务（`trigger_collateral_due_amount` 扣减已成交净回款，按 `liquidation_traders` 实时汇总，不足 0 时返回 `0`） |
| remaining_collateral_qty | string | 剩余质押数量（`trigger_collateral_qty` 扣减已成交数量，按 `liquidation_traders` 实时汇总，不足 0 时返回 `0`） |
| fills | array | 成交明细（无成交时为 `[]`） |

**`status` 取值**

| 取值 | 含义 | 是否终态 |
| --- | --- | --- |
| liquidating | 清算执行中（正常卖出流程） | 否 |
| paused | 已暂停（如风控/告警，待恢复或取消） | 否 |
| settled | 已结清（债务已覆盖） | 是 |
| bad_debt | 坏账（质押卖尽仍未能清偿） | 是 |
| cancelled | 已取消（BeTrust 发起取消且 Mojo 受理） | 是 |
| failed | 失败终态（如硬边界触发） | 是 |

**fills[]**

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| exchange_trade_id | string | 成交 ID |
| batch_no | string | 批次号 |
| price | string | 成交价 |
| quantity | string | 成交量 |
| amount | string | 成交额 |
| fee | string | 手续费 |
| trade_time_ms | int64 | 成交时间（毫秒） |
| fee_asset | string | 手续费币种；未知时为空字符串 |

**响应示例**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "liquidation_id": "LQ-20260915-000001",
    "loan_order_id": "LO-987654",
    "status": "liquidating",
    "symbol": "BTCUSDT",
    "total_qty": "0.5",
    "total_amount": "32500.00",
    "total_fee": "32.50",
    "total_net_proceeds_amount": "32467.50",
    "avg_price": "65000.00",
    "remaining_debt_amount": "65032.50",
    "remaining_collateral_qty": "1.0",
    "fills": []
  }
}
```

---

### 3.3 取消强平任务

**POST** `/liquidation/tasks/cancel`

| 说明 | |
| --- | --- |
| 鉴权 | sign + timestamp |
| 行为 | 停止后续卖出；已在途订单按 Mojo 规则处理；**`liquidating` 与 `paused` 均可取消** |

**请求体**

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| liquidation_id | string | 是 | 强平任务 ID，生成的全局唯一 ID |
| timestamp | int64 | 是 | 毫秒时间戳 |
| sign | string | 是 | 签名 |
| cancel_reason | string | 是 | 取消原因 |
| requested_at | int64 | 是 | 发起取消时间（毫秒） |

**请求示例**

```json
{
  "liquidation_id": "LQ-20260915-000001",
  "timestamp": 1757923300000,
  "sign": "d4e5f6...64位hex...",
  "cancel_reason": "user_repaid",
  "requested_at": 1757923299000
}
```

**响应 data**

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| cancel_result | string | 取消受理结果，见下表 |
| message | string | 结果说明 |

**`cancel_result` 取值**

| 取值 | 含义 |
| --- | --- |
| accepted | 取消已受理，Mojo 停止后续卖出 |
| already_finished | 任务已终态，无法再取消 |
| rejected | 取消被拒绝（如任务不存在） |

**`cancel_result` 与典型 `message`**

| cancel_result | message | 含义 |
| --- | --- | --- |
| accepted | 已受理，停止后续卖出 | 取消成功 |
| already_finished | 任务已终态，无法取消 | 任务已结束 |
| rejected | 强平任务不存在 | 无对应 `liquidation_id` |

**响应示例**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "cancel_result": "accepted",
    "message": "已受理，停止后续卖出"
  }
}
```

---

## 四、Mojo 回调 BeTrust（BeTrust 需实现）

清算任务达到**终态**后，Mojo 向 **联调时 BeTrust 提供的 HTTPS 地址** 发起 **POST**，Content-Type 为 `application/json`。

请求体结构体与 **[3.2 查询强平任务](#32-查询强平任务)** 响应 `data` 完全一模一样。

区别仅为：回调时任务已结束，`status` 为终态取值（**`status` 枚举见 [3.2](#32-查询强平任务)**）。

**请求示例（终态 settled）**

```json
{
  "timestamp": 1757924000123,
  "sign": "9abc01...64位hex...",
  "liquidation_id": "LQ-20260915-000001",
  "loan_order_id": "LO-987654",
  "status": "settled",
  "symbol": "BTCUSDT",
  "total_qty": "1.0",
  "total_amount": "65000.00",
  "total_fee": "65.00",
  "total_net_proceeds_amount": "64935.00",
  "remaining_debt_amount": "0",
  "remaining_collateral_qty": "0",
  "avg_price": "65000.00",
  "fills": []
}
```

**BeTrust 响应建议**

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| code | int | 200 表示接收成功 |
| message | string | 说明 |

```json
{
  "code": 200,
  "message": "ok"
}
```
