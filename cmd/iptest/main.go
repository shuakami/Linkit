package main

import (
	"fmt"
	"linkit/pkg/utils"
)

func main() {
	// 初始化IP库
	if err := utils.InitIPSearcher("ip2region/ip2region.xdb"); err != nil {
		fmt.Printf("初始化IP库失败: %v\n", err)
		return
	}
	defer utils.CloseIPSearcher()

	// 测试不同类型的IP
	testIPs := []string{
		"127.0.0.1",       // 本地IP
		"192.168.1.1",     // 私有IP
		"114.114.114.114", // 中国电信DNS
		"8.8.8.8",         // Google DNS
		"220.181.38.148",  // baidu.com
		"104.244.42.193",  // twitter.com
		"31.13.75.35",     // facebook.com
	}

	fmt.Println("开始测试IP地区查询:")
	fmt.Println("----------------------------------------")
	for _, ip := range testIPs {
		region := utils.GetIPRegion(ip)
		fmt.Printf("IP: %-20s\n", ip)
		fmt.Printf("  国家: %-15s\n", region.Country)
		fmt.Printf("  区域: %-15s\n", region.Region)
		fmt.Printf("  省份: %-15s\n", region.Province)
		fmt.Printf("  城市: %-15s\n", region.City)
		fmt.Printf("  ISP:  %-15s\n", region.ISP)
		fmt.Println("----------------------------------------")
	}
}
