# Valorant商店数据获取脚本

这个脚本允许你使用Cookie登录Valorant并获取当前商店数据，方便后续开发和调试。

## 功能特性

- 使用Cookie认证登录Valorant API
- 获取当前商店物品信息
- 支持指定不同区域 (ap, na, eu, kr)
- 可以将结果输出到文件或标准输出
- 提供丰富的调试信息

## 使用方法

### 编译脚本

```bash
cd val-store-server
go build -o vstore-get ./cmd/scripts/get_store_data.go
```

### 运行脚本

```bash
./vstore-get -cookies "your_cookies_here" -region ap -output store_data.json
```

### 参数说明

- `-cookies` (必须): Valorant cookies字符串，格式为："cookie1=value1; cookie2=value2"
- `-region` (可选): 区域代码 (ap, na, eu, kr)，默认为 "ap"
- `-output` (可选): 输出文件路径，如果不指定则输出到标准输出
- `-debug` (可选): 启用调试模式，提供更详细的日志和Cookie检查信息
- `-help`: 显示帮助信息

### 调试模式

使用`-debug`参数启用调试模式，这将提供更详细的信息：

```bash
./vstore-get -cookies "your_cookies_here" -region ap -debug
```

调试模式会输出：
- Cookie检查结果，显示关键Cookie是否存在
- 详细的认证流程日志
- API请求和响应信息

## 获取Cookie的方法

1. 打开浏览器，登录 [playvalorant.com](https://playvalorant.com/)
2. 使用浏览器的开发者工具 (F12)
3. 切换到"Network"(网络)选项卡
4. 刷新页面
5. 找到对 auth.riotgames.com 的请求
6. 在请求头中找到 Cookie 部分，复制完整的cookie字符串

## 输出示例

```json
{
  "timestamp": "2023-05-12T16:30:45+08:00",
  "region": "ap",
  "user_id": "your-user-id",
  "username": "YourName#TAG",
  "storefront": {
    "FeaturedBundle": {
      // 精选套装信息
    },
    "SkinsPanelLayout": {
      // 每日商店皮肤信息
    }
    // 其他商店数据
  },
  "request_url": "https://pd.ap.a.pvp.net/store/v3/storefront/your-user-id",
  "access_token": "your-access-token",
  "entitlement": "your-entitlement-token"
}
```

## 故障排查

如果您在使用脚本时遇到问题，请参阅 [TROUBLESHOOTING.md](./TROUBLESHOOTING.md) 获取详细的故障排查步骤。

## 注意事项

- Cookie有效期通常为24小时，过期后需要重新获取
- 区域设置应与你的账号所在区域匹配
- 对于敏感信息(令牌等)，建议不要在生产环境或共享环境中保存
- 确保Cookie中包含关键值，尤其是`ssid`和`csid` 