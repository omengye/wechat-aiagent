package server

import (
	"github.com/cloudwego/hertz/pkg/app/server"
)

// HertzHttp Hertz HTTP服务器封装
type HertzHttp struct {
	Server *server.Hertz
}

// NewHertzHttp 创建Hertz HTTP服务器实例
// address: 监听地址，格式为 ":port" 或 "host:port"
func NewHertzHttp(address string) *HertzHttp {
	h := &HertzHttp{
		Server: server.Default(server.WithHostPorts(address)),
	}

	return h
}

// Run 启动HTTP服务器
func (h *HertzHttp) Run() {
	h.Server.Spin()
}
