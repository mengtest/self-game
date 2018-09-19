package dao

import (
	"github.com/jinzhu/gorm"
	"self_game/compoments"
	"self_game/utils/logging"
)

var (
	logger      = logging.GetLogger()
	redisClient *compoments.RedisInstance
	db          *gorm.DB
)

func init() {
	db = compoments.GetDB()
	redisClient = compoments.GetRedisClient()
}
