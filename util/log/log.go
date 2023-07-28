package log

import (
	"github.com/wonderivan/logger"
)

/*
 * 使用方法: 直接 import	"github.com/wonderivan/logger"
 * logger.Error()
 * logger.Debug()
 */

func InitLogger() {
	logger.SetLogger("./log.json")
	logger.Debug("debug init : %v ", logger.LevelMap)
}
