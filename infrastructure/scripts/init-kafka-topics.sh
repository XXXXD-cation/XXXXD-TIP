 #!/bin/bash

# 等待Kafka服务启动
echo "等待Kafka服务启动..."
until kafka-topics --bootstrap-server kafka:9092 --list > /dev/null 2>&1; do
    printf '.'
    sleep 2
done
echo "Kafka服务已启动"

# 创建主题
echo "开始创建Kafka主题..."

# 情报采集与处理相关主题
kafka-topics --bootstrap-server kafka:9092 --create --if-not-exists --topic raw-intelligence --partitions 3 --replication-factor 1
kafka-topics --bootstrap-server kafka:9092 --create --if-not-exists --topic processed-intelligence --partitions 3 --replication-factor 1
kafka-topics --bootstrap-server kafka:9092 --create --if-not-exists --topic enrichment-requests --partitions 3 --replication-factor 1
kafka-topics --bootstrap-server kafka:9092 --create --if-not-exists --topic enrichment-results --partitions 3 --replication-factor 1

# 通知和告警相关主题
kafka-topics --bootstrap-server kafka:9092 --create --if-not-exists --topic notifications --partitions 3 --replication-factor 1
kafka-topics --bootstrap-server kafka:9092 --create --if-not-exists --topic alerts --partitions 3 --replication-factor 1

# 审计日志相关主题
kafka-topics --bootstrap-server kafka:9092 --create --if-not-exists --topic audit-logs --partitions 3 --replication-factor 1

echo "Kafka主题创建完成"