# 错误处理工具包

本包提供了威胁情报平台(TIP)的标准错误处理功能，包括错误类型定义、错误码规范和错误处理工具函数。

## 功能特性

- 统一的错误结构和错误码系统
- 支持错误嵌套和错误链
- 区分内部错误和用户可见错误
- 与HTTP状态码集成
- 支持详细错误信息和上下文
- 预定义常见错误类型的工厂函数

## 使用方法

### 基本用法

```go
import "github.com/XXXXD-cation/XXXXD-TIP/service/pkg/errors"

// 创建一个简单错误
err := errors.NotFound("用户不存在")

// 添加详细信息
err = err.WithDetail("找不到ID为123的用户记录")

// 包装标准库错误
if dbErr != nil {
    return errors.DatabaseError(dbErr)
}
```

### 错误检查

```go
// 检查错误类型
if errors.IsErrorCode(err, errors.CodeNotFound) {
    // 处理资源不存在情况
}
```

### 错误链和Unwrap

```go
// 包装一个已有错误
baseErr := someFunction()
wrappedErr := errors.Wrap(baseErr, errors.CodeInternalServer, 500, "操作失败")

// 解包获取原始错误
originalErr := wrappedErr.Unwrap()
```

## 预定义错误类型

- `BadRequest(message)` - 请求参数错误
- `Unauthorized(message)` - 认证失败
- `Forbidden(message)` - 权限不足
- `NotFound(message)` - 资源不存在
- `Conflict(message)` - 资源冲突
- `InternalServerError(message)` - 服务器内部错误
- `ServiceUnavailable(message)` - 服务不可用
- `RequestTimeout(message)` - 请求超时
- `ValidationError(message)` - 参数验证错误
- `DatabaseError(err)` - 数据库操作错误

## 错误码范围

- `CodeInvalid (1000-1999)` - 无效的输入参数错误
- `CodeUnauthorized (2000-2999)` - 认证错误 
- `CodeForbidden (3000-3999)` - 授权错误
- `CodeNotFound (4000-4999)` - 资源不存在错误
- `CodeAlreadyExists (5000-5999)` - 资源已存在错误
- `CodeResourceExhausted (6000-6999)` - 资源耗尽错误
- `CodeInternalServer (7000-7999)` - 内部服务器错误
- `CodeUnavailable (8000-8999)` - 服务不可用错误
- `CodeTimeout (9000-9999)` - 超时错误

## 示例代码

请参考 `example/main.go` 获取完整示例。 