# Valorant商店数据获取工具 - 故障排查指南

如果您在使用Valorant商店数据获取工具时遇到问题，特别是Cookie登录失败，可以参考本指南进行排查。

## Cookie登录失败的常见原因

1. **Cookie已过期或无效**
   - Riot Games的Cookie通常有效期为24小时
   - 某些区域可能有更短的有效期

2. **Cookie格式不正确**
   - 确保Cookie字符串格式正确，通常为`key1=value1; key2=value2`
   - 不要包含多余的空格或引号

3. **缺少关键Cookie**
   - 登录需要特定的Cookie，特别是`ssid`和`csid`
   - 确保完整复制浏览器中的所有Cookie

4. **区域设置不正确**
   - 确保使用与您账户相匹配的区域

## 修复步骤

### 1. 获取新的Cookie

1. 打开浏览器（推荐Chrome或Firefox）
2. 访问 [https://auth.riotgames.com](https://auth.riotgames.com)
3. 完成登录流程
4. 打开开发者工具（F12或右键->检查）
5. 选择"网络"(Network)选项卡
6. 刷新页面
7. 找到对`auth.riotgames.com`的请求
8. 在请求标头中找到Cookie字段
9. 右键复制完整的Cookie值

### 2. 使用调试模式

使用`--debug`或`-d`选项运行工具以获取更详细的日志：

```bash
./get_store.sh -d -c="your_cookies" -r=ap -o=output.json
```

或者直接运行可执行文件：

```bash
VSTORE_LOG_LEVEL=debug ./vstore-get -cookies "your_cookies" -region ap -output output.json
```

### 3. 检查关键Cookie

验证您的Cookie包含以下关键值：

- `ssid` - 最重要，用作访问令牌
- `csid` - 客户端会话ID
- `clid` - 客户端ID
- `sub` - 主题/用户ID

您可以使用此命令查看您的Cookie中是否包含这些关键值：

```bash
echo "your_cookie_string" | grep -E 'ssid|csid|clid|sub'
```

### 4. 调整请求头

如果标准方法不起作用，您可以尝试在源代码中调整请求头：

修改`internal/repositories/valorant_api.go`中的`authenticateWithCookiesViaAuthorizeEndpoint`函数，添加更多请求头：

```go
// 添加更多可能需要的头信息
req.Header.Set("Accept-Language", "en-US,en;q=0.9") 
req.Header.Set("Origin", "https://auth.riotgames.com")
req.Header.Set("Referer", "https://auth.riotgames.com/login")
```

### 5. 特殊情况处理

#### 区域检测失败

如果您确定自己的区域但工具无法正确识别，尝试显式指定区域：

```bash
./get_store.sh -c="your_cookies" -r=eu
```

#### 多账户问题

如果您有多个Riot账户，确保使用正确账户的Cookie。在单独的浏览器配置文件中登录可能会有所帮助。

#### 权限问题

一些防病毒软件或网络设置可能会阻止API请求。尝试暂时禁用防病毒软件或使用不同的网络连接。

## 常见错误消息及解决方法

| 错误消息 | 可能原因 | 解决方法 |
|---------|---------|---------|
| `无法解析Cookie字符串` | Cookie格式不正确 | 确保正确复制完整Cookie字符串 |
| `获取用户信息失败，状态码: 400` | Cookie无效或已过期 | 获取新的Cookie |
| `Cookie无效或已过期` | 重定向到登录页面 | 重新登录并获取新Cookie |
| `无法从响应中提取令牌` | 授权流程失败 | 确保Cookie完整，尝试重新获取 |
| `获取授权令牌失败` | 访问令牌无效 | 获取新的Cookie |

## 仍然遇到问题？

如果您尝试了上述所有方法但仍然无法解决问题，请考虑以下操作：

1. 检查Riot Games服务器状态
2. 确认API端点未发生变化
3. 查看最新的更新日志是否有相关说明

## 联系支持

如您需要进一步帮助，请提供以下信息：

- 详细的错误消息
- 使用的命令行参数
- 操作系统和环境信息
- 重现问题的步骤 