// Package main 提供一个错误处理包的使用示例
package main

import (
	"database/sql"
	"fmt"

	"github.com/XXXXD-cation/XXXXD-TIP/service/pkg/errors"
)

// 模拟一个数据库操作函数
func getUserFromDB(id string) (string, error) {
	// 模拟数据库错误
	if id == "error" {
		return "", sql.ErrNoRows
	}

	// 模拟成功情况
	if id == "123" {
		return "张三", nil
	}

	// 模拟用户未找到
	return "", sql.ErrNoRows
}

// 业务逻辑层函数
func getUser(id string) (string, error) {
	// 调用数据访问层
	user, err := getUserFromDB(id)
	if err != nil {
		// 当数据库返回记录未找到错误时
		if err == sql.ErrNoRows {
			return "", errors.NotFound("用户不存在").WithDetail("找不到ID为" + id + "的用户记录")
		}

		// 其他数据库错误
		return "", errors.DatabaseError(err).WithDetail("查询用户数据失败")
	}

	return user, nil
}

// API层函数
func handleGetUser(id string) {
	fmt.Printf("处理获取用户请求: ID=%s\n", id)

	// 参数验证
	if id == "" {
		err := errors.ValidationError("用户ID不能为空")
		fmt.Printf("错误: %v\n", err)
		return
	}

	// 调用业务逻辑层
	user, err := getUser(id)
	if err != nil {
		// 检查是否为特定类型的错误
		if errors.IsErrorCode(err, errors.CodeNotFound) {
			fmt.Printf("错误: %v\n", err)
			return
		}

		// 其他错误
		fmt.Printf("错误: %v\n", err)
		return
	}

	fmt.Printf("成功: 用户名=%s\n", user)
}

func main() {
	fmt.Println("=== 错误处理示例 ===")

	// 1. 有效请求
	fmt.Println("\n案例1: 有效的用户ID")
	handleGetUser("123")

	// 2. 空ID
	fmt.Println("\n案例2: 空用户ID")
	handleGetUser("")

	// 3. 不存在的用户
	fmt.Println("\n案例3: 不存在的用户ID")
	handleGetUser("999")

	// 4. 数据库错误
	fmt.Println("\n案例4: 触发数据库错误")
	handleGetUser("error")

	// 5. 测试错误嵌套与Unwrap
	fmt.Println("\n案例5: 错误嵌套与Unwrap")
	baseErr := fmt.Errorf("原始错误")
	wrappedErr := errors.Wrap(baseErr, errors.CodeInternalServer, 500, "包装的错误")
	fmt.Printf("完整错误: %v\n", wrappedErr)
	unwrappedErr := wrappedErr.Unwrap()
	fmt.Printf("解包后: %v\n", unwrappedErr)
}
