# ========================
# Stage 1: Builder
# ========================
FROM node:20-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 package 文件
COPY package*.json ./

# 安装依赖
RUN npm install --prefer-offline --no-audit

# 复制源代码
COPY . .

# 构建生产版本（增加Node.js内存限制到4GB）
ENV NODE_OPTIONS="--max-old-space-size=4096"
RUN npm run build

# ========================
# Stage 2: Runtime (Nginx)
# ========================
FROM nginx:alpine

# 复制自定义nginx配置
COPY <<'EOF' /etc/nginx/conf.d/default.conf
# HTTP：健康检查 + 自动跳转 HTTPS
server {
    listen 80;
    server_name _;

    location = /healthz {
        return 200 'ok';
    }

    location / {
        return 301 https://$host$request_uri;
    }
}

# HTTPS 主服务
server {
    listen 443 ssl;
    server_name _;
    root /usr/share/nginx/html;
    index index.html;

    # SSL 证书（挂载自宿主机 ./certs 目录）
    ssl_certificate     /etc/nginx/certs/fullchain.pem;
    ssl_certificate_key /etc/nginx/certs/privkey.pem;
    ssl_protocols       TLSv1.2 TLSv1.3;
    ssl_ciphers         HIGH:!aNULL:!MD5;
    ssl_session_cache   shared:SSL:10m;
    ssl_session_timeout 10m;

    # Gzip压缩
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css text/xml text/javascript application/javascript application/xml+rss application/json;

    # 安全头
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # SPA路由支持
    location / {
        try_files $uri $uri/ /index.html;
    }

    # API代理（如果前端和后端在同一域名）
    location /api/ {
        proxy_pass http://backend:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # 静态资源缓存
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
EOF

# 从builder复制构建产物
COPY --from=builder /app/dist /usr/share/nginx/html

# 暴露端口
EXPOSE 80
EXPOSE 443

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost/healthz || exit 1

# 启动nginx
CMD ["nginx", "-g", "daemon off;"]
