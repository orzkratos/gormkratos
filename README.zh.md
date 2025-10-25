[![GitHub Workflow Status (branch)](https://img.shields.io/github/actions/workflow/status/orzkratos/gormkratos/release.yml?branch=main&label=BUILD)](https://github.com/orzkratos/gormkratos/actions/workflows/release.yml?query=branch%3Amain)
[![GoDoc](https://pkg.go.dev/badge/github.com/orzkratos/gormkratos)](https://pkg.go.dev/github.com/orzkratos/gormkratos)
[![Coverage Status](https://img.shields.io/coveralls/github/orzkratos/gormkratos/main.svg)](https://coveralls.io/github/orzkratos/gormkratos?branch=main)
[![Supported Go Versions](https://img.shields.io/badge/Go-1.25+-lightgrey.svg)](https://github.com/orzkratos/gormkratos)
[![GitHub Release](https://img.shields.io/github/release/orzkratos/gormkratos.svg)](https://github.com/orzkratos/gormkratos/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/orzkratos/gormkratos)](https://goreportcard.com/report/github.com/orzkratos/gormkratos)

# gormkratos

Kratos 框架的 GORM 事务封装,具备双错误返回模式。

---

<!-- TEMPLATE (ZH) BEGIN: LANGUAGE NAVIGATION -->
## 英文文档

[ENGLISH README](README.md)
<!-- TEMPLATE (ZH) END: LANGUAGE NAVIGATION -->

## 主要特性

🎯 **双错误模式**: 区分业务逻辑错误和数据库事务错误
⚡ **上下文支持**: 内置上下文超时和取消处理
🔄 **自动回滚**: 业务逻辑错误时的事务自动回滚
🌍 **Kratos 集成**: 与 Kratos 微服务框架的顺畅集成
📋 **简洁 API**: 干净简洁的事务封装函数

## 安装

```bash
go get github.com/orzkratos/gormkratos
```

## 使用方法

### 基础事务

此示例展示 gormkratos.Transaction 的最简单用法。

```go
package main

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/google/uuid"
	"github.com/orzkratos/gormkratos"
	"github.com/yyle88/must"
	"github.com/yyle88/rese"
	"github.com/yyle88/zaplog"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Admin struct {
	ID   uint   `gorm:"primarykey"`
	Name string `gorm:"not null"`
}

func main() {
	dsn := fmt.Sprintf("file:db-%s?mode=memory&cache=shared", uuid.New().String())
	db := rese.P1(gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}))
	defer rese.F0(rese.P1(db.DB()).Close)

	must.Done(db.AutoMigrate(&Admin{}))

	ctx := context.Background()

	erk := Transaction(ctx, db, func(db *gorm.DB) *errors.Error {
		admin := &Admin{Name: "Alice"}
		if err := db.Create(admin).Error; err != nil {
			return ErrorServerDbError("create failed: %v", err)
		}
		zaplog.LOG.Debug("Created admin", zap.Uint("id", admin.ID), zap.String("name", admin.Name))
		return nil
	})
	if erk != nil {
		zaplog.LOG.Error("Error", zap.Error(erk))
	}
}

func ErrorServerDbError(format string, args ...interface{}) *errors.Error {
	return errors.New(500, "DB_ERROR", fmt.Sprintf(format, args...))
}

func ErrorServerDbTransactionError(format string, args ...interface{}) *errors.Error {
	return errors.New(500, "TRANSACTION_ERROR", fmt.Sprintf(format, args...))
}

func Transaction(ctx context.Context, db *gorm.DB, run func(db *gorm.DB) *errors.Error) *errors.Error {
	erk, err := gormkratos.Transaction(ctx, db, run)
	if err != nil {
		if erk != nil {
			return erk
		}
		return ErrorServerDbTransactionError("transaction failed: %v", err)
	}
	return nil
}
```

⬆️ **源码:** [源码](internal/demos/demo1x/main.go)

### 事务回滚

此示例展示业务逻辑返回错误时的自动回滚。

```go
package main

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/google/uuid"
	"github.com/orzkratos/gormkratos"
	"github.com/yyle88/must"
	"github.com/yyle88/rese"
	"github.com/yyle88/zaplog"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Guest struct {
	ID   uint   `gorm:"primarykey"`
	Name string `gorm:"not null"`
}

func main() {
	dsn := fmt.Sprintf("file:db-%s?mode=memory&cache=shared", uuid.New().String())
	db := rese.P1(gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}))
	defer rese.F0(rese.P1(db.DB()).Close)

	must.Done(db.AutoMigrate(&Guest{}))

	ctx := context.Background()

	erk := Transaction(ctx, db, func(db *gorm.DB) *errors.Error {
		guest := &Guest{Name: "Bob"}
		if err := db.Create(guest).Error; err != nil {
			return ErrorServerDbError("create failed: %v", err)
		}
		zaplog.LOG.Debug("Created guest (then rollback)", zap.Uint("id", guest.ID), zap.String("name", guest.Name))
		return ErrorBadRequest("validation failed")
	})
	zaplog.LOG.Error("Error", zap.Error(erk))

	var count int64
	db.Model(&Guest{}).Count(&count)
	zaplog.LOG.Debug("Guest count post rollback", zap.Int64("count", count))
}

func ErrorServerDbError(format string, args ...interface{}) *errors.Error {
	return errors.New(500, "DB_ERROR", fmt.Sprintf(format, args...))
}

func ErrorBadRequest(format string, args ...interface{}) *errors.Error {
	return errors.New(400, "BAD_REQUEST", fmt.Sprintf(format, args...))
}

func ErrorServerDbTransactionError(format string, args ...interface{}) *errors.Error {
	return errors.New(500, "TRANSACTION_ERROR", fmt.Sprintf(format, args...))
}

func Transaction(ctx context.Context, db *gorm.DB, run func(db *gorm.DB) *errors.Error) *errors.Error {
	erk, err := gormkratos.Transaction(ctx, db, run)
	if err != nil {
		if erk != nil {
			return erk
		}
		return ErrorServerDbTransactionError("transaction failed: %v", err)
	}
	return nil
}
```

⬆️ **源码:** [源码](internal/demos/demo2x/main.go)

### 多个操作

此示例展示在一个原子事务中组合创建和更新操作。

```go
package main

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/google/uuid"
	"github.com/orzkratos/gormkratos"
	"github.com/yyle88/must"
	"github.com/yyle88/rese"
	"github.com/yyle88/zaplog"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Product struct {
	ID    uint   `gorm:"primarykey"`
	Name  string `gorm:"not null"`
	Price int
}

func main() {
	dsn := fmt.Sprintf("file:db-%s?mode=memory&cache=shared", uuid.New().String())
	db := rese.P1(gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}))
	defer rese.F0(rese.P1(db.DB()).Close)

	must.Done(db.AutoMigrate(&Product{}))

	ctx := context.Background()

	erk := Transaction(ctx, db, func(db *gorm.DB) *errors.Error {
		product := &Product{Name: "Laptop", Price: 5000}
		if err := db.Create(product).Error; err != nil {
			return ErrorServerDbError("create failed: %v", err)
		}
		zaplog.LOG.Debug("Created product", zap.Uint("id", product.ID), zap.String("name", product.Name), zap.Int("price", product.Price))

		product.Price = 4500
		if err := db.Updates(product).Error; err != nil {
			return ErrorServerDbError("update failed: %v", err)
		}
		zaplog.LOG.Debug("Updated product", zap.Uint("id", product.ID), zap.String("name", product.Name), zap.Int("price", product.Price))
		return nil
	})
	if erk != nil {
		zaplog.LOG.Error("Error", zap.Error(erk))
	}
}

func ErrorServerDbError(format string, args ...interface{}) *errors.Error {
	return errors.New(500, "DB_ERROR", fmt.Sprintf(format, args...))
}

func ErrorServerDbTransactionError(format string, args ...interface{}) *errors.Error {
	return errors.New(500, "TRANSACTION_ERROR", fmt.Sprintf(format, args...))
}

func Transaction(ctx context.Context, db *gorm.DB, run func(db *gorm.DB) *errors.Error) *errors.Error {
	erk, err := gormkratos.Transaction(ctx, db, run)
	if err != nil {
		if erk != nil {
			return erk
		}
		return ErrorServerDbTransactionError("transaction failed: %v", err)
	}
	return nil
}
```

⬆️ **源码:** [源码](internal/demos/demo3x/main.go)

## 错误处理

`gormkratos.Transaction` 函数返回两个错误以帮助区分不同类型：

1. **业务逻辑错误** (`erk *errors.Error`): 来自业务逻辑的 Kratos 框架错误
2. **数据库事务错误** (`err error`): 数据库事务错误

### 场景

**当 err != nil:**
- `erk != nil`: 业务逻辑错误导致回滚
- `erk == nil`: 数据库提交失败

**当 err == nil:**
- `erk` 也是 nil，两者都成功

## 示例

### 基础双错误返回

**直接使用 gormkratos.Transaction:**
```go
erk, err := gormkratos.Transaction(ctx, db, func(db *gorm.DB) *errors.Error {
    user := &User{Name: "test"}
    if err := db.Create(user).Error; err != nil {
        return errorspb.ErrorServerDbError("创建失败: %v", err)
    }
    return nil
})
```

**检查业务错误:**
```go
if erk != nil {
    // 处理 Kratos 业务错误
    log.Printf("业务逻辑失败: %v", erk)
}
```

**检查数据库错误:**
```go
if err != nil {
    // 处理数据库事务错误
    log.Printf("数据库事务失败: %v", err)
}
```

### 使用事务选项

**设置事务隔离级别:**
```go
import "database/sql"

erk, err := gormkratos.Transaction(ctx, db, func(db *gorm.DB) *errors.Error {
    // 自定义隔离级别的事务逻辑
    return nil
}, &sql.TxOptions{
    Isolation: sql.LevelReadCommitted,
    ReadOnly:  false,
})
```

### 单个事务中的多个操作

**组合创建和更新:**
```go
erk, err := gormkratos.Transaction(ctx, db, func(db *gorm.DB) *errors.Error {
    product := &Product{Name: "Laptop", Price: 5000}
    if err := db.Create(product).Error; err != nil {
        return ErrorServerDbError("创建失败: %v", err)
    }

    product.Price = 4500
    if err := db.Updates(product).Error; err != nil {
        return ErrorServerDbError("更新失败: %v", err)
    }
    return nil
})
```

### 上下文超时处理

**超时时自动回滚:**
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

erk, err := gormkratos.Transaction(ctx, db, func(db *gorm.DB) *errors.Error {
    // 长时间运行的操作
    time.Sleep(10 * time.Second) // 超过超时时间
    return nil
})
// err 将包含超时错误
```

<!-- TEMPLATE (ZH) BEGIN: STANDARD PROJECT FOOTER -->
<!-- VERSION 2025-09-26 07:39:27.188023 +0000 UTC -->

## 📄 许可证类型

MIT 许可证。详见 [LICENSE](LICENSE)。

---

## 🤝 项目贡献

非常欢迎贡献代码！报告 BUG、建议功能、贡献代码：

- 🐛 **发现问题？** 在 GitHub 上提交问题并附上重现步骤
- 💡 **功能建议？** 创建 issue 讨论您的想法
- 📖 **文档疑惑？** 报告问题，帮助我们改进文档
- 🚀 **需要功能？** 分享使用场景，帮助理解需求
- ⚡ **性能瓶颈？** 报告慢操作，帮助我们优化性能
- 🔧 **配置困扰？** 询问复杂设置的相关问题
- 📢 **关注进展？** 关注仓库以获取新版本和功能
- 🌟 **成功案例？** 分享这个包如何改善工作流程
- 💬 **反馈意见？** 欢迎提出建议和意见

---

## 🔧 代码贡献

新代码贡献，请遵循此流程：

1. **Fork**：在 GitHub 上 Fork 仓库（使用网页界面）
2. **克隆**：克隆 Fork 的项目（`git clone https://github.com/yourname/repo-name.git`）
3. **导航**：进入克隆的项目（`cd repo-name`）
4. **分支**：创建功能分支（`git checkout -b feature/xxx`）
5. **编码**：实现您的更改并编写全面的测试
6. **测试**：（Golang 项目）确保测试通过（`go test ./...`）并遵循 Go 代码风格约定
7. **文档**：为面向用户的更改更新文档，并使用有意义的提交消息
8. **暂存**：暂存更改（`git add .`）
9. **提交**：提交更改（`git commit -m "Add feature xxx"`）确保向后兼容的代码
10. **推送**：推送到分支（`git push origin feature/xxx`）
11. **PR**：在 GitHub 上打开 Merge Request（在 GitHub 网页上）并提供详细描述

请确保测试通过并包含相关的文档更新。

---

## 🌟 项目支持

非常欢迎通过提交 Merge Request 和报告问题来为此项目做出贡献。

**项目支持：**

- ⭐ **给予星标**如果项目对您有帮助
- 🤝 **分享项目**给团队成员和（golang）编程朋友
- 📝 **撰写博客**关于开发工具和工作流程 - 我们提供写作支持
- 🌟 **加入生态** - 致力于支持开源和（golang）开发场景

**祝你用这个包编程愉快！** 🎉🎉🎉

<!-- TEMPLATE (ZH) END: STANDARD PROJECT FOOTER -->

---

## GitHub 标星点赞

[![Stargazers](https://starchart.cc/orzkratos/gormkratos.svg?variant=adaptive)](https://starchart.cc/orzkratos/gormkratos)