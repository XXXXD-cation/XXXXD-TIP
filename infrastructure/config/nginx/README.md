# NGINX API网关配置

## 概述

本目录包含威胁情报平台(TIP)的NGINX API网关配置文件。NGINX在此项目中用作API网关，负责路由前端请求到后端微服务，并提供基本的反向代理、负载均衡和缓存功能。

## 配置文件结构

- `nginx.conf`: 主配置文件，包含全局NGINX设置
- `conf.d/default.conf`: 默认服务器配置，包含所有API路由和微服务代理设置

## 路由配置

当前配置了以下路由：

### API路由

- `/api/auth/` - 用户认证服务
- `/api/iocs/` - 情报查询服务 
- `/api/cases/` - 案例管理服务
- `/api/ingest/` - 情报导入服务
- `/api/admin/` - 平台管理服务
- `/api/reports/` - 报告服务
- `/api/integrations/` - 集成服务
- `/api/search/` - 搜索服务

### 前端路由

- `/app/` - 前端UI应用
- `/static/` - 静态资源文件
- `/docs/` - 文档站点

### 特殊路由

- `/health` - 健康检查端点
- `/` - API网关默认欢迎页

## 运行环境

NGINX容器在以下端口运行：

- HTTP: 8000 (本地访问地址: http://localhost:8000)
- HTTPS: 8443 (本地访问地址: https://localhost:8443) - 尚未配置SSL证书

## 使用说明

### 添加新微服务路由

要添加新的微服务路由，请在`conf.d/default.conf`文件中按照现有示例添加新的`location`块：

```nginx
location /api/your-service/ {
    proxy_pass http://your-service-host:port/;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header X-Request-ID $request_id;
    
    # 超时设置
    proxy_connect_timeout 5s;
    proxy_send_timeout 10s;
    proxy_read_timeout 10s;
}
```

### 修改配置后重启

修改配置文件后，需要重启NGINX容器以应用更改：

```bash
docker-compose restart nginx
``` 