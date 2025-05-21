// Package config 提供了应用程序的配置管理功能封装，基于Viper实现
// 支持从配置文件、环境变量加载配置，并提供类型安全的访问接口
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Provider 定义了配置提供器的接口
type Provider interface {
	// GetString 获取字符串配置
	GetString(key string) string
	// GetInt 获取整数配置
	GetInt(key string) int
	// GetBool 获取布尔配置
	GetBool(key string) bool
	// GetFloat64 获取浮点数配置
	GetFloat64(key string) float64
	// GetStringSlice 获取字符串切片配置
	GetStringSlice(key string) []string
	// GetStringMap 获取字符串映射配置
	GetStringMap(key string) map[string]interface{}
	// GetDuration 获取时间段配置
	GetDuration(key string) time.Duration
	// Get 获取任意类型配置
	Get(key string) interface{}
	// GetOrDefault 获取配置，如果不存在则返回默认值
	GetOrDefault(key string, defaultVal interface{}) interface{}
	// IsSet 判断配置是否存在
	IsSet(key string) bool
	// AllSettings 获取所有配置
	AllSettings() map[string]interface{}
	// UnmarshalKey 将指定键的配置解析到结构体
	UnmarshalKey(key string, val interface{}) error
	// Unmarshal 将全部配置解析到结构体
	Unmarshal(val interface{}) error
}

// ViperProvider 是使用Viper实现的配置提供器
type ViperProvider struct {
	v *viper.Viper
}

// 确保ViperProvider实现了Provider接口
var _ Provider = (*ViperProvider)(nil)

// 全局配置实例及互斥锁
var (
	instance     Provider
	instanceLock sync.RWMutex
)

// Options 定义配置初始化选项
type Options struct {
	// ConfigName 配置文件名称（不含扩展名）
	ConfigName string
	// ConfigType 配置文件类型（扩展名，如yaml, json）
	ConfigType string
	// ConfigPaths 配置文件搜索路径
	ConfigPaths []string
	// EnvPrefix 环境变量前缀
	EnvPrefix string
	// AutomaticEnv 是否自动加载环境变量
	AutomaticEnv bool
	// WatchConfig 是否监视配置文件变化
	WatchConfig bool
	// WatchConfigCallback 配置文件变化回调函数
	WatchConfigCallback func()
}

// DefaultOptions 返回默认配置选项
func DefaultOptions() Options {
	return Options{
		ConfigName:   "config",
		ConfigType:   "yaml",
		ConfigPaths:  []string{".", "./config", "/etc/app", "$HOME/.app"},
		EnvPrefix:    "APP",
		AutomaticEnv: true,
		WatchConfig:  false,
		WatchConfigCallback: func() {
			fmt.Println("配置文件已更新")
		},
	}
}

// New 创建新的配置提供器
func New(opts Options) (Provider, error) {
	v := viper.New()

	// 设置配置文件信息
	v.SetConfigName(opts.ConfigName)
	v.SetConfigType(opts.ConfigType)

	// 添加配置文件搜索路径
	for _, path := range opts.ConfigPaths {
		// 处理路径中的环境变量
		if strings.HasPrefix(path, "$") {
			parts := strings.SplitN(path, "/", 2)
			envVar := strings.TrimPrefix(parts[0], "$")
			envVal := os.Getenv(envVar)
			if envVal != "" {
				if len(parts) > 1 {
					path = filepath.Join(envVal, parts[1])
				} else {
					path = envVal
				}
			}
		}
		v.AddConfigPath(path)
	}

	// 加载环境变量
	if opts.EnvPrefix != "" {
		v.SetEnvPrefix(opts.EnvPrefix)
	}
	if opts.AutomaticEnv {
		v.AutomaticEnv()
	}
	// 设置环境变量分隔符，例如APP_DATABASE_URL中的_
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 尝试加载配置文件
	if err := v.ReadInConfig(); err != nil {
		// 如果配置文件不存在，记录但不返回错误
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("读取配置文件错误: %w", err)
		}
		fmt.Printf("警告: 未找到配置文件，将使用环境变量和默认值\n")
	}

	// 监视配置文件变化
	if opts.WatchConfig {
		v.WatchConfig()
		if opts.WatchConfigCallback != nil {
			v.OnConfigChange(func(e fsnotify.Event) {
				opts.WatchConfigCallback()
			})
		}
	}

	return &ViperProvider{v: v}, nil
}

// GetInstance 获取全局配置实例，如果不存在则使用默认选项初始化
func GetInstance() Provider {
	instanceLock.RLock()
	if instance != nil {
		defer instanceLock.RUnlock()
		return instance
	}
	instanceLock.RUnlock()

	instanceLock.Lock()
	defer instanceLock.Unlock()
	if instance != nil {
		return instance
	}

	// 使用默认选项创建配置实例
	opts := DefaultOptions()
	provider, err := New(opts)
	if err != nil {
		// 记录错误但继续使用空配置实例
		fmt.Printf("初始化配置时发生错误: %v\n", err)
		provider = &ViperProvider{v: viper.New()}
	}
	instance = provider
	return instance
}

// SetInstance 设置全局配置实例
func SetInstance(provider Provider) {
	instanceLock.Lock()
	defer instanceLock.Unlock()
	instance = provider
}

// Reset 重置全局配置实例
func Reset() {
	instanceLock.Lock()
	defer instanceLock.Unlock()
	instance = nil
}

// Provider接口实现
func (p *ViperProvider) GetString(key string) string {
	return p.v.GetString(key)
}

func (p *ViperProvider) GetInt(key string) int {
	return p.v.GetInt(key)
}

func (p *ViperProvider) GetBool(key string) bool {
	return p.v.GetBool(key)
}

func (p *ViperProvider) GetFloat64(key string) float64 {
	return p.v.GetFloat64(key)
}

func (p *ViperProvider) GetStringSlice(key string) []string {
	return p.v.GetStringSlice(key)
}

func (p *ViperProvider) GetStringMap(key string) map[string]interface{} {
	return p.v.GetStringMap(key)
}

func (p *ViperProvider) GetDuration(key string) time.Duration {
	return p.v.GetDuration(key)
}

func (p *ViperProvider) Get(key string) interface{} {
	return p.v.Get(key)
}

func (p *ViperProvider) GetOrDefault(key string, defaultVal interface{}) interface{} {
	if !p.v.IsSet(key) {
		return defaultVal
	}
	return p.v.Get(key)
}

func (p *ViperProvider) IsSet(key string) bool {
	return p.v.IsSet(key)
}

func (p *ViperProvider) AllSettings() map[string]interface{} {
	return p.v.AllSettings()
}

func (p *ViperProvider) UnmarshalKey(key string, val interface{}) error {
	return p.v.UnmarshalKey(key, val)
}

func (p *ViperProvider) Unmarshal(val interface{}) error {
	return p.v.Unmarshal(val)
}

// 全局便捷函数
func GetString(key string) string {
	return GetInstance().GetString(key)
}

func GetInt(key string) int {
	return GetInstance().GetInt(key)
}

func GetBool(key string) bool {
	return GetInstance().GetBool(key)
}

func GetFloat64(key string) float64 {
	return GetInstance().GetFloat64(key)
}

func GetStringSlice(key string) []string {
	return GetInstance().GetStringSlice(key)
}

func GetStringMap(key string) map[string]interface{} {
	return GetInstance().GetStringMap(key)
}

func GetDuration(key string) time.Duration {
	return GetInstance().GetDuration(key)
}

func Get(key string) interface{} {
	return GetInstance().Get(key)
}

func GetOrDefault(key string, defaultVal interface{}) interface{} {
	return GetInstance().GetOrDefault(key, defaultVal)
}

func IsSet(key string) bool {
	return GetInstance().IsSet(key)
}

func AllSettings() map[string]interface{} {
	return GetInstance().AllSettings()
}

func UnmarshalKey(key string, val interface{}) error {
	return GetInstance().UnmarshalKey(key, val)
}

func Unmarshal(val interface{}) error {
	return GetInstance().Unmarshal(val)
}
