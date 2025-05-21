 #!/bin/bash

# 等待MinIO服务启动
echo "等待MinIO服务启动..."
until $(curl --output /dev/null --silent --head --fail http://minio:9000); do
    printf '.'
    sleep 1
done
echo "MinIO服务已启动"

# 使用MinIO客户端创建存储桶
echo "开始创建MinIO存储桶..."
mc alias set minio http://minio:9000 minioadmin minioadmin

# 创建存储桶
mc mb minio/case-attachments
mc mb minio/reports
mc mb minio/ioc-exports

# 设置访问策略
# mc policy set download minio/case-attachments
# mc policy set download minio/reports
# mc policy set download minio/ioc-exports

echo "MinIO存储桶创建完成"