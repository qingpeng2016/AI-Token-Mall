package entity

import (
	"time"
)

// BotScheduleConfig 脚本调度配置表
type BotScheduleConfig struct {
	ID                    uint      `gorm:"primaryKey;column:id;type:bigint unsigned;not null;autoIncrement;comment:主键ID"`
	Module                string    `gorm:"column:module;type:varchar(50);not null;comment:业务模块"`
	TaskName              string    `gorm:"column:task_name;type:varchar(100);not null;comment:任务名称"`
	IntervalSeconds       float64   `gorm:"column:interval_seconds;type:float;not null;default:60;comment:执行间隔（秒）"`
	ExeSort               *int      `gorm:"column:exe_sort;type:int;comment:统一流水线顺序，仅>0参与；0或NULL不参与"`
	Concurrency           *int      `gorm:"column:concurrency;type:int;comment:并发数，NULL表示不限制或使用默认值"`
	Description           string    `gorm:"column:description;type:varchar(500);comment:任务描述"`
	IsEnabled             int       `gorm:"column:is_enabled;type:tinyint(1);not null;default:1;comment:是否启用：1-启用，0-禁用"`
	IsStrategyEnabled     *int      `gorm:"column:is_strategy_enabled;type:tinyint(1);default:1"`
	LastLiveTime          string    `gorm:"column:last_live_time;type:varchar(255)"`
	LastLiveTimeByMachine string    `gorm:"column:last_live_time_by_machine;type:json;comment:各机器最后活跃时间 {\"machine_name\":\"datetime\"}"`
	IsPrimaryMachineRun   int       `gorm:"column:is_primary_machine_run;type:int;not null;default:0;comment:是否主机器运行"`
	CreatedAt             time.Time `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:创建时间"`
	UpdatedAt             time.Time `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间"`
}

func (BotScheduleConfig) TableName() string {
	return "bot_schedule_config"
}
