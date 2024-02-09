package conf

import (
	"github.com/spf13/viper"
	"time"
)

// IVisitConf 访问配置
type IVisitConf interface {
	Sub(key string) IVisitConf
	Get(key string) any
	GetString(key string) string
	GetBool(key string) bool
	GetInt(key string) int
	GetInt32(key string) int32
	GetInt64(key string) int64
	GetUint(key string) uint
	GetUint16(key string) uint16
	GetUint32(key string) uint32
	GetUint64(key string) uint64
	GetFloat64(key string) float64
	GetTime(key string) time.Time
	GetDuration(key string) time.Duration
	GetIntSlice(key string) []int
	GetStringSlice(key string) []string
	GetStringMap(key string) map[string]any
	GetStringMapString(key string) map[string]string
	GetStringMapStringSlice(key string) map[string][]string
	GetSizeInBytes(key string) uint
	GetAllKeys() []string
	GetAllSettings() map[string]any
}

type _VisitConf struct {
	*viper.Viper
}

func (vc *_VisitConf) Sub(key string) IVisitConf {
	return &_VisitConf{
		Viper: vc.Viper.Sub(key),
	}
}

func (vc *_VisitConf) GetAllKeys() []string {
	return vc.AllKeys()
}

func (vc *_VisitConf) GetAllSettings() map[string]any {
	return vc.AllSettings()
}
