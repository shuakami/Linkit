package utils

import (
	"net"
	"strings"
	"sync"

	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
)

var (
	searcher *xdb.Searcher
	once     sync.Once
)

// RegionInfo 表示地区信息
type RegionInfo struct {
	Country  string // 国家
	Region   string // 区域
	Province string // 省份
	City     string // 城市
	ISP      string // ISP
}

// InitIPSearcher 初始化IP搜索器
func InitIPSearcher(dbPath string) error {
	var err error
	once.Do(func() {
		// 创建searcher对象
		searcher, err = xdb.NewWithFileOnly(dbPath)
	})
	return err
}

// parseRegion 解析地区字符串
func parseRegion(region string) *RegionInfo {
	parts := strings.Split(region, "|")
	if len(parts) != 5 {
		return &RegionInfo{
			Country:  "UNKNOWN",
			Region:   "UNKNOWN",
			Province: "UNKNOWN",
			City:     "UNKNOWN",
			ISP:      "UNKNOWN",
		}
	}

	// 处理"0"值
	for i := range parts {
		if parts[i] == "0" {
			parts[i] = ""
		}
	}

	return &RegionInfo{
		Country:  parts[0],
		Region:   parts[1],
		Province: parts[2],
		City:     parts[3],
		ISP:      parts[4],
	}
}

// GetIPRegion 获取IP所属地区
func GetIPRegion(ip string) *RegionInfo {
	// 处理本地回环地址
	if ip == "127.0.0.1" || ip == "::1" || ip == "localhost" {
		return &RegionInfo{
			Country:  "LOCAL",
			Region:   "LOCAL",
			Province: "LOCAL",
			City:     "LOCAL",
			ISP:      "LOCAL",
		}
	}

	// 检查是否是私有IP
	if ipAddr := net.ParseIP(ip); ipAddr != nil {
		if ipAddr.IsPrivate() || ipAddr.IsLoopback() {
			return &RegionInfo{
				Country:  "LOCAL",
				Region:   "LOCAL",
				Province: "LOCAL",
				City:     "LOCAL",
				ISP:      "LOCAL",
			}
		}
	}

	// 如果searcher未初始化，返回UNKNOWN
	if searcher == nil {
		return &RegionInfo{
			Country:  "UNKNOWN",
			Region:   "UNKNOWN",
			Province: "UNKNOWN",
			City:     "UNKNOWN",
			ISP:      "UNKNOWN",
		}
	}

	// 搜索IP地区
	region, err := searcher.SearchByStr(ip)
	if err != nil {
		return &RegionInfo{
			Country:  "UNKNOWN",
			Region:   "UNKNOWN",
			Province: "UNKNOWN",
			City:     "UNKNOWN",
			ISP:      "UNKNOWN",
		}
	}

	return parseRegion(region)
}

// CloseIPSearcher 关闭IP搜索器
func CloseIPSearcher() {
	if searcher != nil {
		searcher.Close()
	}
}
