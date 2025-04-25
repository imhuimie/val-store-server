# Val-Store-Server

这是Val-Store-Server服务器，专注于提供Valorant皮肤商店信息查询服务。当前版本（v1）只实现了cookie登录Valorant的功能。

## 项目结构

```
val-store-server/
├── cmd/                # 应用程序入口点
│   └── server/         # 主服务器应用
├── internal/           # 私有应用程序和库代码
│   ├── api/            # API处理和路由
│   │   ├── handlers/   # HTTP处理器
│   │   ├── middleware/ # HTTP中间件
│   │   └── router.go   # 路由配置
│   ├── config/         # 配置管理
│   ├── models/         # 数据模型
│   ├── repositories/   # 数据存储和外部API交互
│   └── services/       # 业务逻辑
├── .env.example        # 环境变量示例
├── go.mod              # Go模块文件
└── go.sum              # Go模块校验和
```

## 环境变量配置

创建`.env`文件，并配置以下环境变量：

```
# 服务器配置
PORT=8080                   # 服务器端口
GIN_MODE=debug              # Gin框架模式（debug, release, test）
JWT_SECRET=your-secret-key  # JWT密钥（生产环境必须更改）
ALLOWED_ORIGINS=http://localhost:3000  # 允许的CORS源（多个值用逗号分隔）
```

## 使用方法

### 安装和编译

1. 克隆仓库并进入项目目录:
   ```
   git clone https://github.com/emper0r/val-store-server.git
   cd val-store-server
   ```

2. 安装依赖:
   ```
   go mod tidy
   ```

3. 编译:
   ```
   go build -o val-store-server ./cmd/server
   ```

4. 运行:
   ```
   ./val-store-server
   ```

## API接口

### 认证

#### Cookie登录

- **URL**: `/api/auth/login/cookies`
- **方法**: `POST`
- **描述**: 使用Riot游戏Cookies登录
- **请求体**:
  ```json
  {
    "cookies": "ssid=xxx; csid=xxx; ...",
    "region": "ap"  // 可选，指定游戏区域
  }
  ```
- **响应**:
  ```json
  {
    "status": 200,
    "message": "登录成功",
    "data": {
      "token": "eyJhbGciOiJIUzI1NiIs...",
      "user": {
        "username": "your_username",
        "user_id": "your_user_id"
      }
    }
  }
  ```

#### 健康检查

- **URL**: `/api/auth/ping`
- **方法**: `GET`
- **描述**: 检查服务是否正常运行
- **响应**:
  ```json
  {
    "status": 200,
    "message": "服务正常运行"
  }
  ```

## Cookie获取方法

要获取用于登录的Riot/Valorant Cookie，可以按照以下步骤操作：

1. 在浏览器中打开 https://auth.riotgames.com
2. 使用你的Riot账号登录
3. 登录成功后，按F12打开开发者工具
4. 切换到"应用"或"Application"标签页
5. 在左侧边栏中找到"Cookies" > "auth.riotgames.com"
6. 找到并复制以下几个关键Cookie：ssid, csid 等
7. 将这些Cookie按照格式整理成字符串："ssid=xxx; csid=xxx; ..."

## 错误码

| 状态码 | 描述                  |
|--------|---------------------|
| 200    | 请求成功              |
| 400    | 请求参数错误           |
| 401    | 未授权（认证失败）      |
| 404    | 请求的资源不存在       |
| 500    | 服务器内部错误         |

## 注意事项

- 这只是初始版本，仅实现了cookie登录功能
- 请妥善保管你的Riot Cookie，不要分享给他人
- 仅用于个人用途，不得用于商业目的 