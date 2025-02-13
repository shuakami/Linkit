# Linkit API 文档

本目录包含了 Linkit 短链接服务的 API 文档。

## 文档格式

API 文档使用 OpenAPI 3.0 (前身是 Swagger) 规范编写，存储在 `api.yaml` 文件中。

## 查看文档

有多种方式可以查看该文档：

### 1. 在线查看器

您可以使用以下在线工具查看文档：

- [Redoc](https://redocly.github.io/redoc/)
  1. 访问 https://redocly.github.io/redoc/
  2. 点击 "Upload a file"
  3. 将 `api.yaml` 的内容上传上去

- [Swagger Editor](https://editor.swagger.io/)
  1. 访问 https://editor.swagger.io/
  2. 将 `api.yaml` 的内容复制粘贴到编辑器中
  3. 右侧会自动显示可交互的文档界面

### 2. 本地查看

您也可以在本地搭建文档服务器：

#### 使用 Docker

```bash
# 使用 Swagger UI
docker run -p 80:8080 -e SWAGGER_JSON=/api.yaml -v $(pwd):/usr/share/nginx/html/api swaggerapi/swagger-ui

# 使用 Redoc
docker run -p 80:80 -e SPEC_URL=api.yaml -v $(pwd):/usr/share/nginx/html redocly/redoc
```

#### 使用 Node.js

```bash
# 安装 redoc-cli
npm install -g redoc-cli

# 启动文档服务器
redoc-cli serve api.yaml
```

## 最佳实践

1. 在开发前，您应该仔细阅读 API 概述部分，了解使用方法和处理机制。

2. 您可以使用文档中提供的示例作为参考，这些示例涵盖了最常见的使用场景。


## 更新记录

#### 2024-02-13

> 初始版本 

None