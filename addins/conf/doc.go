// Package conf 提供基于 Viper 的服务配置 add-in。
//
// 可通过 A 访问合并后的应用配置，通过 S 访问当前服务的配置子树，
// 并通过 With 自定义配置加载行为。此实现会保留到服务终止后，因此仍可在
// OnTerminated 回调中使用。
package conf
