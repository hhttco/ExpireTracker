# 资产到期智能监控系统 (Resource Expiration Tracker)

这是一个基于 **Go (Golang)** 和 **原生 HTML/CSS/JS** 开发的轻量级全端自适应 Web 网站。专门用于记录和管理各类具有时效性的商品或资产（如：域名过期时间、VPS/服务器到期时间、SSL证书到期日等）。系统自带高颜值暗黑系毛玻璃（Glassmorphism）视觉特效，并完美支持手机端自适应。

---

## 🌟 功能特性

- 🔒 **安全身份验证**：内置管理员登录拦截机制（认证中间件），保障资产数据安全。
- 📊 **智能临期排序**：资产列表自动按到期时间由近到远（升序）排列，快要过期的资产自动置顶。
- 🚨 **三级临期色彩预警**：列表左侧拥有磨砂半透明发光条，根据剩余天数自动切换视觉状态：
  - 🟢 **安全（>30天）**：绿色常亮边框
  - 🟡 **警告（≤30天）**：黄色常亮边框，天数加粗显示
  - 🔴 **极危（≤7天）**：红色发光边框，状态文字伴随呼吸灯闪烁特效
  - 🔘 **已逾期（<0天）**：灰色置底并自动添加文字横线划掉效果
- 🔄 **快捷续费与编辑**：点击“续费”无刷新弹出毛玻璃对话框，直接日历选择新日期即可完成更新。
- ➕ **一键录入与删除**：资产录入隐藏于悬浮弹窗中，日期选择器智能默认填充“当天日期”；每行数据支持防误触弹窗确认删除。
- ⚙️ **微型安全管理**：管理菜单精简合并，支持在控制台内直接修改管理员登录密码。
- 🌓 **昼夜双模式切换**：支持“暗色极光模式”与“明亮磨砂模式”无缝一键切换，并完美记忆用户的主题偏好。
- 📱 **多端全自适应（响应式）**：在 PC 端呈现大气平铺的五等宽表格；在手机移动端自动解体并重构为适合单手操作的纵向流式卡片。
- 💾 **本地文件持久化**：采用轻量级嵌入式 SQLite 数据库，数据持久化存储于本地文件，重启不丢失，无需安装复杂的数据库服务。

---

## 🛠️ 技术栈

- **后端 (Backend)**: Go 1.20+ (使用官方内置标准库 `net/http`, `html/template`, `database/sql`)
- **前端 (Frontend)**: 原生 HTML5, CSS3 (包含 Flexbox, Grid 布局与 Media Queries 响应式), 原生 JavaScript (ES6+)
- **数据库 (Database)**: SQLite3 (嵌入式单文件数据库)
- **依赖驱动 (Driver)**: `github.com/mattn/go-sqlite3` (或纯 Go 版本 `modernc.org/sqlite`)

---

## 📁 项目目录结构

```text
expire-tracker/
├── main.go               # Go 后端核心逻辑、路由及 SQLite 数据库控制器
├── tracker.db            # 运行时自动生成的 SQLite 本地数据库文件 (自动创建)
└── templates/            # 前端 HTML 模板文件夹
    ├── login.html        # 高端星空流光登录页面
    └── index.html        # 响应式毛玻璃资产监控主面板
```

---
## 🚀 快速部署指南
```bash
apt install -y nginx && cd /var/www
wget -N https://github.com/hhttco/ExpireTracker/releases/download/v1.0.0/et.zip -O ./et.zip
unzip ./et.zip -d /var/www/et
chown -R www-data:www-data /var/www/et && chmod -R 755 /var/www/et
rm ./et.zip
chmod +x /var/www/et/et

vim /etc/systemd/system/et.service
[Unit]
Description=Et Service App
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/var/www/et
ExecStart=/var/www/et/et
Restart=always
MemoryLimit=50M

[Install]
WantedBy=multi-user.target

启动
systemctl daemon-reload
systemctl enable et
systemctl start et
systemctl status et

配置域名 安装证书

vim /etc/nginx/conf.d/et.conf
```

```bash
# ==========================================
# 极致优化的 HTTP (80端口) 
# ==========================================
server {
    # 替换为您的域名或服务器公网 IP
    server_name your_domain.com; 

    # 禁用未绑定域名的直接 IP 访问（防止恶意解析和网络垃圾爬虫扫描）
    if ($host != $server_name) {
        return 444; # Nginx 特有状态码：立刻掐断连接，不返回任何字节
    }

    # 限制上传文件大小上限（作为系统安全规范）
    client_max_body_size 10M;

    # ------------------------------------------
    # [极致优化的 Gzip 压缩] 解决毛玻璃复杂 HTML 加载速度问题
    # ------------------------------------------
    gzip on;
    gzip_min_length 1024; 
    gzip_comp_level 5; 
    gzip_vary on; 
    gzip_proxied any; 
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml text/javascript application/x-javascript;

    # ------------------------------------------
    # [网站安全加固请求头] 防跨站、防劫持、隐藏环境
    # ------------------------------------------
    add_header X-Frame-Options "SAMEORIGIN" always; 
    add_header X-XSS-Protection "1; mode=block" always; 
    add_header X-Content-Type-Options "nosniff" always; 
    add_header Referrer-Policy "strict-origin-when-cross-origin" always; 

    # ------------------------------------------
    # [反向代理核心路由] 对接后台 Golang 服务端口
    # ------------------------------------------
    location / {
        proxy_pass http://127.0.0.1:8089;

        # 高性能 HTTP/1.1 长连接支持
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";

        # 透传最精准的客户端真实 IP 及环境参数
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # [代理缓冲区深度优化] 彻底消除毛玻璃大网页加载中断隐患
        proxy_buffering on;
        proxy_buffer_size 128k; 
        proxy_buffers 4 256k; 
        proxy_busy_buffers_size 256k;
        proxy_max_temp_file_size 0; # 禁用磁盘临时文件，全在内存中极速传输

        proxy_connect_timeout 60s;
        proxy_read_timeout 60s;
        proxy_send_timeout 60s;
    }

    # ------------------------------------------
    # [漏洞探针拦截] 智能拦截黑客对恶意漏洞的扫描
    # ------------------------------------------
    location ~* \.(php|pl|py|jsp|sh|cgi|env|git|svn|hg)$ {
        deny all;
        access_log off;
        log_not_found off;
    }

    # ------------------------------------------
    # 当 Go 后端程序挂掉时，直接就地返回中文友好提示，不需要依赖任何本地 HTML 文件
    # ------------------------------------------
    error_page 502 503 504 = @backend_down;

    location @backend_down {
        default_type text/html;
        return 502 "<html><head><title>服务暂不可用</title></head><body style='background:#0f172a;color:#f8fafc;font-family:sans-serif;padding:50px;text-align:center;'><h2>⚠️ 资产监控系统服务暂时无法连接</h2><p style='color:#94a3b8;'>后端 Go 程序可能正在维护或已停止运行，请稍后再试。</p></body></html>";
    }
}
```

---

## 自定义启动指南

### 1. 克隆或下载项目
确保本地已配置好 Go 语言开发环境，将项目放置在您的工作目录下。

### 2. 初始化模块并下载 SQLite 驱动
在项目根目录下打开终端，执行以下命令：
```bash
# 初始化 Go 模块
go mod init expire-tracker

# 下载广泛应用的 SQLite3 驱动
go get github.com/mattn/go-sqlite3
```
> 💡 **Windows 用户避坑提示**：若运行或编译时提示缺少 `gcc` 环境，可在 `main.go` 中将驱动包替换为不需要 C 语言编译器的纯 Go 版本：
> ```bash
> go get modernc.org/sqlite
> ```
> 并将代码中的 `_ "github.com/mattn/go-sqlite3"` 修改为 `_ "modernc.org/sqlite"`，数据库打开方式改为 `sql.Open("sqlite", ...)`。

### 3. 本地编译与运行
在终端执行：
```bash
go run main.go
```
当控制台打印出：`服务器已启动: http://localhost:8080` 时，代表 Web 服务已成功驻留后台监听。

### 4. 浏览器访问
打开浏览器，访问地址：
👉 **[http://localhost:8080](http://localhost:8080)**

- **默认初始化管理员账号**：`admin`
- **默认初始化管理员密码**：`admin`

---

## 📦 打包与服务器部署

由于 Go 语言拥有极强的交叉编译能力，且本项目未引入任何外部 CDN 静态文件，部署非常简单：

### 1. 递归打包源码
若需将代码搬运至其他机器，请注意 `zip` 命令需加 `-r` 参数以递归包含 HTML 模板文件：
```bash
zip -r et.zip main.go templates/
```

### 2. 编译为独立二进制程序
在部署到服务器前，您可以在本地直接编译出无源码的可执行文件：
```bash
# 编译出可执行文件 (Windows 下为 expire-tracker.exe，Linux 下为 expire-tracker)
通用: go build -o expire-tracker main.go

Linux: CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o et main.go
```
**部署部署提示**：在服务器上运行编译出的程序时，请确保运行路径下存有 `templates` 文件夹及其中的 `.html` 文件，否则程序在解析路由模板时会因找不到路径而报错退出。
