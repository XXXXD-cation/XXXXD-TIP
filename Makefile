 .PHONY: dev-env-up dev-env-down dev-env-restart dev-env-logs dev-env-status help

# 默认命令
help:
	@echo "威胁情报平台(TIP)开发环境管理工具"
	@echo ""
	@echo "可用命令:"
	@echo "  make dev-env-up      - 启动开发环境的所有依赖服务"
	@echo "  make dev-env-down    - 停止开发环境的所有依赖服务"
	@echo "  make dev-env-restart - 重新启动开发环境的所有依赖服务"
	@echo "  make dev-env-logs    - 查看开发环境的服务日志"
	@echo "  make dev-env-status  - 检查开发环境服务状态"
	@echo "  make help            - 显示帮助信息"
	@echo ""

# 启动开发环境
dev-env-up:
	@echo "启动开发环境..."
	docker-compose up -d
	@echo "开发环境启动完成"

# 停止开发环境
dev-env-down:
	@echo "停止开发环境..."
	docker-compose down
	@echo "开发环境已停止"

# 重新启动开发环境
dev-env-restart:
	@echo "重新启动开发环境..."
	docker-compose restart
	@echo "开发环境已重新启动"

# 查看日志
dev-env-logs:
	docker-compose logs -f

# 检查服务状态
dev-env-status:
	docker-compose ps