package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/emper0r/val-store-server/internal/repositories"
)

func main() {
	// 定义命令行参数
	cookiesPtr := flag.String("cookies", "", "Valorant cookies字符串，格式为：\"cookie1=value1; cookie2=value2\"")
	regionPtr := flag.String("region", "ap", "区域代码，例如: ap, na, eu, kr")
	outputFilePtr := flag.String("output", "", "输出文件路径（可选，默认输出到标准输出）")
	helpPtr := flag.Bool("help", false, "显示帮助信息")
	debugPtr := flag.Bool("debug", false, "启用调试模式")

	// 解析命令行参数
	flag.Parse()

	// 显示帮助信息
	if *helpPtr || *cookiesPtr == "" {
		printUsage()
		return
	}

	// 启用调试模式
	if *debugPtr {
		log.Println("调试模式已启用")
		// 禁用可能泄露敏感信息的日志输出
		log.SetFlags(log.Ldate | log.Ltime) // 仅保留日期和时间，移除文件和行号
		// 这里避免设置过于详细的日志级别
	}

	// 初始化ValorantAPI
	api, err := repositories.NewValorantAPI()
	if err != nil {
		log.Fatalf("初始化API失败: %v", err)
	}

	// 设置区域
	region := strings.ToLower(*regionPtr)
	api.SetRegion(region)
	fmt.Printf("使用区域: %s\n", region)

	// 解析Cookie
	cookies := repositories.ParseCookieString(*cookiesPtr)
	if len(cookies) == 0 {
		log.Fatalf("无法解析Cookie字符串，请确保格式正确")
	}

	// 输出关键Cookie检查
	if *debugPtr {
		checkKeyCookies(cookies)
	}

	// 使用Cookie进行认证
	session, err := api.AuthenticateWithCookies(cookies)
	if err != nil {
		log.Fatalf("认证失败: %v", err)
	}

	fmt.Printf("认证成功! 用户: %s#%s (ID: %s)\n", session.RiotUsername, session.RiotTagline, session.UserID)

	// 获取商店原始数据
	rawData, err := api.GetStoreOffersRaw(session.UserID, session.AccessToken, session.Entitlement)
	if err != nil {
		log.Fatalf("获取商店数据失败: %v", err)
	}

	// 输出数据
	if *outputFilePtr != "" {
		// 写入文件
		err = os.WriteFile(*outputFilePtr, rawData, 0644)
		if err != nil {
			log.Fatalf("写入文件失败: %v", err)
		}
		fmt.Printf("商店数据已保存到: %s\n", *outputFilePtr)
	} else {
		// 直接输出原始JSON，确保它是有效的JSON格式
		// 首先检查是否为有效的JSON（这一步是可选的）
		var jsonObj interface{}
		if err := json.Unmarshal(rawData, &jsonObj); err != nil {
			log.Fatalf("获取到的数据不是有效的JSON: %v", err)
		}

		// 为了格式美观，重新格式化JSON（可选步骤）
		prettyJSON, err := json.MarshalIndent(jsonObj, "", "  ")
		if err != nil {
			// 如果格式化失败，就使用原始数据
			fmt.Println(string(rawData))
		} else {
			fmt.Println(string(prettyJSON))
		}
	}
}

// 检查关键Cookie是否存在
func checkKeyCookies(cookies map[string]string) {
	keyCookies := []string{"ssid", "csid", "clid", "sub", "tdid", "asid"}
	fmt.Println("检查关键Cookie:")

	for _, key := range keyCookies {
		if _, exists := cookies[key]; exists {
			fmt.Printf("  ✓ %s: ***敏感信息已隐藏***\n", key)
		} else {
			fmt.Printf("  ✗ %s: 不存在\n", key)
		}
	}

	fmt.Printf("总计Cookie数量: %d\n", len(cookies))
}

func printUsage() {
	fmt.Println("Valorant商店数据获取工具")
	fmt.Println("-------------------------")
	fmt.Println("此工具使用Cookie登录Valorant并获取当前商店数据")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Printf("  %s -cookies \"cookie1=value1; cookie2=value2\" -region ap [-output data.json] [-debug]\n", os.Args[0])
	fmt.Println()
	fmt.Println("参数:")
	fmt.Println("  -cookies string   Valorant cookies字符串")
	fmt.Println("  -region string    区域代码 (ap, na, eu, kr) (默认 \"ap\")")
	fmt.Println("  -output string    输出文件路径 (可选，默认输出到标准输出)")
	fmt.Println("  -debug            启用调试模式 (可选)")
	fmt.Println("  -help             显示帮助信息")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Printf("  %s -cookies \"ssid=value; csid=value\" -region ap -output store_data.json\n", os.Args[0])
}
