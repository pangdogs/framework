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
	GetStringOrDefault(key, def string) string
	GetBoolOrDefault(key string, def bool) bool
	GetIntOrDefault(key string, def int) int
	GetInt32OrDefault(key string, def int32) int32
	GetInt64OrDefault(key string, def int64) int64
	GetUintOrDefault(key string, def uint) uint
	GetUint16OrDefault(key string, def uint16) uint16
	GetUint32OrDefault(key string, def uint32) uint32
	GetUint64OrDefault(key string, def uint64) uint64
	GetFloat64OrDefault(key string, def float64) float64
	GetTimeOrDefault(key string, def time.Time) time.Time
	GetDurationOrDefault(key string, def time.Duration) time.Duration
	GetIntSliceOrDefault(key string, def []int) []int
	GetStringSliceOrDefault(key string, def []string) []string
	GetStringMapOrDefault(key string, def map[string]any) map[string]any
	GetStringMapStringOrDefault(key string, def map[string]string) map[string]string
	GetStringMapStringSliceOrDefault(key string, def map[string][]string) map[string][]string
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

func (vc *_VisitConf) GetStringOrDefault(key, def string) string {
	if !vc.IsSet(key) {
		return def
	}
	return vc.GetString(key)
}

func (vc *_VisitConf) GetBoolOrDefault(key string, def bool) bool {
	if !vc.IsSet(key) {
		return def
	}
	return vc.GetBool(key)
}

func (vc *_VisitConf) GetIntOrDefault(key string, def int) int {
	if !vc.IsSet(key) {
		return def
	}
	return vc.GetInt(key)
}

func (vc *_VisitConf) GetInt32OrDefault(key string, def int32) int32 {
	if !vc.IsSet(key) {
		return def
	}
	return vc.GetInt32(key)
}

func (vc *_VisitConf) GetInt64OrDefault(key string, def int64) int64 {
	if !vc.IsSet(key) {
		return def
	}
	return vc.GetInt64(key)
}

func (vc *_VisitConf) GetUintOrDefault(key string, def uint) uint {
	if !vc.IsSet(key) {
		return def
	}
	return vc.GetUint(key)
}

func (vc *_VisitConf) GetUint16OrDefault(key string, def uint16) uint16 {
	if !vc.IsSet(key) {
		return def
	}
	return vc.GetUint16(key)
}

func (vc *_VisitConf) GetUint32OrDefault(key string, def uint32) uint32 {
	if !vc.IsSet(key) {
		return def
	}
	return vc.GetUint32(key)
}

func (vc *_VisitConf) GetUint64OrDefault(key string, def uint64) uint64 {
	if !vc.IsSet(key) {
		return def
	}
	return vc.GetUint64(key)
}

func (vc *_VisitConf) GetFloat64OrDefault(key string, def float64) float64 {
	if !vc.IsSet(key) {
		return def
	}
	return vc.GetFloat64(key)
}

func (vc *_VisitConf) GetTimeOrDefault(key string, def time.Time) time.Time {
	if !vc.IsSet(key) {
		return def
	}
	return vc.GetTime(key)
}

func (vc *_VisitConf) GetDurationOrDefault(key string, def time.Duration) time.Duration {
	if !vc.IsSet(key) {
		return def
	}
	return vc.GetDuration(key)
}

func (vc *_VisitConf) GetIntSliceOrDefault(key string, def []int) []int {
	if !vc.IsSet(key) {
		return def
	}
	return vc.GetIntSlice(key)
}

func (vc *_VisitConf) GetStringSliceOrDefault(key string, def []string) []string {
	if !vc.IsSet(key) {
		return def
	}
	return vc.GetStringSlice(key)
}

func (vc *_VisitConf) GetStringMapOrDefault(key string, def map[string]any) map[string]any {
	if !vc.IsSet(key) {
		return def
	}
	return vc.GetStringMap(key)
}

func (vc *_VisitConf) GetStringMapStringOrDefault(key string, def map[string]string) map[string]string {
	if !vc.IsSet(key) {
		return def
	}
	return vc.GetStringMapString(key)
}

func (vc *_VisitConf) GetStringMapStringSliceOrDefault(key string, def map[string][]string) map[string][]string {
	if !vc.IsSet(key) {
		return def
	}
	return vc.GetStringMapStringSlice(key)
}

func (vc *_VisitConf) GetAllKeys() []string {
	return vc.AllKeys()
}

func (vc *_VisitConf) GetAllSettings() map[string]any {
	return vc.AllSettings()
}
