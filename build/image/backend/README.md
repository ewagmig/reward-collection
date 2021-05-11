# Usage

## 编译镜像

- 进入到该目录下面
- 执行以下命令编译镜像

```sh

docker build -t baasBackend:1.0.0
```

## 执行镜像

```sh

docker run -d -v /path/to/baas/conf:/etc/baas/server -v /path/to/baas/runtime:/var/baas/server -p 8001:8001 baasBackend:1.1.0
```

## 容器中重要目录

- `/etc/baas/server`: 默认包含`baas-backend.yaml`和`api/index.html`, 可以映射主机目录来覆盖这这个目录来传递新的`配置`和`API文档`
- `/var/baas/server`: 保存运行时产生的`重要`的文件， **该目录必须备份**

## 重要的环境变量

- `FABRIC_BAAS_SERVER_MODE`: 如果希望用不同的名称管理不同环境的配置文件（默认是baas-backend.yaml), 比如测试和生成环境配置文件分别为`stg.yaml`,`prd.ymal`, 可以给该环境变量赋值分别为: `stg`, `prd`
- `其他`: 配置文件中的配置项都可以用环境变量覆盖, 例如配置文件中的以下配置:

```YAML

logging:
  level: debug
```

可以通过环境变量`BAAS_LOGGING_LEVEL`进行覆盖