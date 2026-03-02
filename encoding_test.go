package gowebdav

import (
	"testing"
)

// TestBuildURLWithChineseChars 测试buildURL方法对中文字符的处理
func TestBuildURLWithChineseChars(t *testing.T) {
	// 创建一个客户端实例
	client, err := NewClient("http://example.com", "user", "pass")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// 测试包含中文字符的路径
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple Chinese path",
			input:    "/共享文件",
			expected: "http://example.com/%E5%85%B1%E4%BA%AB%E6%96%87%E4%BB%B6",
		},
		{
			name:     "nested Chinese path",
			input:    "/共享文件/中文目录",
			expected: "http://example.com/%E5%85%B1%E4%BA%AB%E6%96%87%E4%BB%B6/%E4%B8%AD%E6%96%87%E7%9B%AE%E5%BD%95",
		},
		{
			name:     "mixed path",
			input:    "/shared/中文文件.txt",
			expected: "http://example.com/shared/%E4%B8%AD%E6%96%87%E6%96%87%E4%BB%B6.txt",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := client.buildURL(tc.input)
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}
