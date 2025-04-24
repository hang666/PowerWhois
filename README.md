
## 目录结构

```
.
├── deploy  (部署相关)
│   ├── backend (后端程序)
│   ├── docker-compose.yml
│   ├── frontend    (前端程序)
│   ├── nginx       (Nginx配置文件)
│   └── redis       (Redis相关)
├── README.md   (说明文档)
└── src (源码目录)
    ├── backend     (后端源码)
    ├── frontend    (前端源码)
    └── Makefile
```

# 部署

## 服务器依赖

- docker

## 上传文件

将 deploy 文件夹中的内容原样上传到服务器中

## 配置

- 将 Nginx 配置文件`deploy/nginx/typonamer.conf`中 443 端口和 80 端口的`server_name`修改为真实的目标域名。

- 将 Nginx 配置文件`deploy/nginx/typonamer.conf`中 443 端口的`ssl_certificate`路径和`ssl_certificate_key`路径修改为真实的目标证书路径和密钥路径。

- 将 Nginx 配置文件`deploy/nginx/typonamer.conf`中 443 端口的`root`路径修改为前端文件夹`deploy/frontend`在主机上的真实绝对路径。

- 将 Nginx 配置文件`deploy/nginx/typonamer.conf`复制到主机 Nginx 的配置目录中，主机的 Nginx 的配置目录一般为`/etc/nginx/conf.d/`

- 修改`deploy/backend/config.yaml`文件中的`AuthUsername`和`AuthPassword`。其中`AuthUsername`是管理员的用户名，`AuthPassword`是管理员密码。

## 部署

- cd 到服务器中的 deploy 目录
- 执行命令：`docker compose up -d`
- 重启主机上的 Nginx：`sudo systemctl restart nginx`
- 访问网站，例如：`https://typonamer.example.com`

# 开发

## 后端

后端程序是用 Golang 语言开发，Web 框架使用的是 Gofiber。

### 依赖

- Golang 1.23+

## 前端

前端程序使用 Vue3 开发，UI 框架使用的是 Quasar 2。

### 依赖

- Nodejs v20+
- yarn 1.22+

### 初始化

```
cd src/frontend
yarn
quasar dev
```

## 编译

项目源码目录中已编写好`src/Makefile`文件，在 Linux 中安装好后端依赖、前端依赖以及`make`命令之后即可开始编译。编译后的输出会自动替换 deploy 目录下 backend 和 frontend 中的内容，需要手动将编译后的新文件上传到服务器中替换旧的服务器文件，然后重启容器后生效。

### 单独编译后端代码

```
cd src/
make backend
```

### 单独编译前端代码

```
cd src/
make frontend
```

### 同时编译前端和后端代码

```
cd src/
make all
```
