package gowebdav

import (
	"net/http"
	"net/url"
)

// ProxyConfig 表示代理配置
type ProxyConfig struct {
	HTTP      string // HTTP代理地址
	HTTPS     string // HTTPS代理地址
	Username  string // 代理用户名
	Password  string // 代理密码
}

// SetProxy 设置代理配置
func (c *Client) SetProxy(config ProxyConfig) error {
	// 创建一个新的HTTP客户端，设置代理
	transport := &http.Transport{}

	// 配置代理
	if config.HTTP != "" || config.HTTPS != "" {
		// 解析HTTP代理
		var httpProxy *url.URL
		if config.HTTP != "" {
			var err error
			httpProxy, err = url.Parse(config.HTTP)
			if err != nil {
				return err
			}
			// 设置HTTP代理认证
			if config.Username != "" || config.Password != "" {
				httpProxy.User = url.UserPassword(config.Username, config.Password)
			}
		}

		// 解析HTTPS代理
		var httpsProxy *url.URL
		if config.HTTPS != "" {
			var err error
			httpsProxy, err = url.Parse(config.HTTPS)
			if err != nil {
				return err
			}
			// 设置HTTPS代理认证
			if config.Username != "" || config.Password != "" {
				httpsProxy.User = url.UserPassword(config.Username, config.Password)
			}
		}

		// 设置代理函数
		transport.Proxy = func(req *http.Request) (*url.URL, error) {
			if req.URL.Scheme == "https" && httpsProxy != nil {
				return httpsProxy, nil
			}
			if httpProxy != nil {
				return httpProxy, nil
			}
			return nil, nil
		}
	}

	// 创建新的HTTP客户端
	c.http = &http.Client{
		Transport: transport,
	}

	return nil
}

// SetProxyFromEnvironment 从环境变量中设置代理
func (c *Client) SetProxyFromEnvironment() {
	// 从环境变量中获取代理配置
	c.http = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}
}
