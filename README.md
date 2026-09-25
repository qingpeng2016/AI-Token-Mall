# mojo-strategy-liquidation

质押清算策略服务（BeTrust 触发 / 执行 / 回调）。

## 启动

接口侧（HTTP）：

```
go run main.go --run_conf=local --run_bot=false
```

脚本侧（定时任务）：

```
go run main.go --run_conf=local --run_bot=true
```

## HTTP

- 强平 API：`/liquidation/tasks` 等，见 `docs/质押清算 (3).md`
- Swagger：`http://<host>:8886/swagger/index.html`

## 目录

```
main.go                 入口：--run_bot 区分脚本 / 接口
boot/                   dig 依赖注入
conf/                   yaml 配置
application/
  core-service/              业务逻辑（如 task_service.go）
  common/                    应用层枚举与常量
  strategy/scripts/
    common/                  共用 Binance 同步、档位、撤单等
    taskexec/                bot task_exec
    callback/                bot result_callback（调度名 task_name）
    orderreconcile/          bot order_updater（调度名 task_name）
    taskcancel/              bot task_cancel
domain/
  rest/request/              入参 DTO（如 task.go）
  rest/response/             出参 DTO（如 task.go）
  persistent/                仓储接口与实体
infrastructure/         mysql / redis / http 实现
interfaces/
  handler/                   HTTP 入口（如 task_handler.go）
  rest/                      gin 路由
common/                 日志、错误码、通知、中间件
docs/                   表结构、bot 调度、PRD
```

## 新增业务时

1. 接口：`domain/rest/request|response` 加 DTO → `application/core-service` 写 service → `interfaces/handler` 写 handler → `interfaces/rest` 挂路由 → `boot` 里 `Provide`
2. 脚本：在 `application/strategy/scripts/<task_name>/` 新建包 → `Scheduler.getHandleFunc` 注册 → `docs/liquidation_init.sql` 中 `bot_schedule_config` 加调度 → `boot` 里 `Provide`
