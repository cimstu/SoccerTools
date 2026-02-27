# 部署到 CentOS CVM

服务为单二进制，静态资源已 embed，无需额外文件。监听端口 **3000**。

## 1. 本地构建 Linux 二进制

在项目根目录（Windows 或 Mac 均可交叉编译）：

**Windows (PowerShell)：**

```powershell
cd D:\Work\Personal\SoccerTools
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o soccertools ./cmd/server
```

**Windows (CMD)：**

```cmd
set GOOS=linux
set GOARCH=amd64
go build -o soccertools ./cmd/server
```

**Linux / Mac：**

```bash
GOOS=linux GOARCH=amd64 go build -o soccertools ./cmd/server
```

得到 `soccertools` 可执行文件。

## 2. 上传到 CVM

```bash
scp soccertools root@<CVM_IP>:/tmp/
# 或使用你的用户名和密钥
# scp -i your.pem soccertools centos@<CVM_IP>:/tmp/
```

## 3. 在 CVM 上安装并启动

SSH 登录后：

```bash
# 创建目录和用户（可选，若直接用 root 可跳过创建用户并改 service 里的 User=root）
sudo mkdir -p /opt/soccertools
sudo useradd -r -s /bin/false soccertools 2>/dev/null || true
sudo mv /tmp/soccertools /opt/soccertools/
sudo chown -R soccertools:soccertools /opt/soccertools
sudo chmod +x /opt/soccertools/soccertools
```

上传并启用 systemd 服务（在**本机**把 `deploy/soccertools.service` 拷到 CVM）：

```bash
# 本机执行
scp deploy/soccertools.service root@<CVM_IP>:/tmp/
```

在 **CVM** 上：

```bash
sudo mv /tmp/soccertools.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable soccertools
sudo systemctl start soccertools
sudo systemctl status soccertools
```

## 4. 防火墙放行 3000 端口（如需外网访问）

```bash
sudo firewall-cmd --permanent --add-port=3000/tcp
sudo firewall-cmd --reload
# 或使用 iptables 的云安全组在控制台放行 3000
```

## 5. 验证

- 本机：`curl http://<CVM_IP>:3000/health`
- 浏览器：`http://<CVM_IP>:3000` 打开 Web 页

## 常用命令

```bash
sudo systemctl stop soccertools    # 停止
sudo systemctl start soccertools   # 启动
sudo systemctl restart soccertools # 重启
sudo journalctl -u soccertools -f  # 看日志
```

## 可选：用 Nginx 反代到 80 端口

若希望通过 80 访问且不直接暴露 3000：

```nginx
# /etc/nginx/conf.d/soccertools.conf
server {
    listen 80;
    server_name your-domain-or-ip;
    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

然后 `sudo nginx -t && sudo systemctl reload nginx`。
