# gormkratos

Kratos 框架的 GORM 事务封装，具备双错误返回模式。

---

<!-- TEMPLATE (ZH) BEGIN: LANGUAGE NAVIGATION -->
## 英文文档

[ENGLISH README](README.md)
<!-- TEMPLATE (ZH) END: LANGUAGE NAVIGATION -->

## 核心特性

🎯 **双错误模式**: 区分业务逻辑错误和数据库事务错误  
⚡ **上下文支持**: 内置上下文超时和取消处理  
🔄 **自动回滚**: 业务逻辑错误时的事务自动回滚  
🌍 **Kratos 集成**: 与 Kratos 微服务框架的无缝集成  
📋 **简洁 API**: 干净易用的事务封装函数

## 安装

```bash
go install github.com/orzkratos/gormkratos@latest
```

## 使用方法

### 基础事务

```go
package main

import (
    "context"
    "github.com/orzkratos/gormkratos"
    "gorm.io/gorm"
)

func CreateUser(ctx context.Context, db *gorm.DB, name string) error {
    erk, err := gormkratos.Transaction(ctx, db, func(db *gorm.DB) *errors.Error {
        user := &User{Name: name}
        if err := db.Create(user).Error; err != nil {
            return errorspb.ErrorServerDbError("创建用户失败: %v", err)
        }
        return nil
    })
    
    if err != nil {
        if erk != nil {
            // 业务逻辑错误
            return erk
        }
        // 数据库事务错误
        return fmt.Errorf("事务失败: %w", err)
    }
    return nil
}
```

### 业务层封装

```go
// 为业务层使用封装 gormkratos.Transaction
func Transaction(ctx context.Context, db *gorm.DB, run func(db *gorm.DB) *errkratos.Erk) *errkratos.Erk {
    erk, err := gormkratos.Transaction(ctx, db, run)
    if err != nil {
        if erk != nil {
            return erk
        }
        return errorspb.ErrorServerDbTransactionError("error=%v", err)
    }
    return nil
}

// 使用示例
func BusinessOperation(ctx context.Context, db *gorm.DB) *errkratos.Erk {
    return Transaction(ctx, db, func(db *gorm.DB) *errkratos.Erk {
        // 您的业务逻辑
        return nil
    })
}
```

### 使用事务选项

```go
import "database/sql"

erk, err := gormkratos.Transaction(ctx, db, func(db *gorm.DB) *errors.Error {
    // 您的事务逻辑
    return nil
}, &sql.TxOptions{
    Isolation: sql.LevelReadCommitted,
    ReadOnly:  false,
})
```

## 错误处理

`gormkratos.Transaction` 函数返回两个错误以帮助区分不同的错误类型：

1. **业务逻辑错误** (`erk *errors.Error`): 来自业务逻辑的 Kratos 框架错误
2. **数据库事务错误** (`err error`): 底层数据库或事务错误

### 错误场景

- **成功**: `erk = nil, err = nil`
- **业务错误**: `erk != nil, err != nil` (事务回滚触发)
- **数据库错误**: `erk = nil, err != nil` (数据库级别问题)

## 示例

查看 [演示代码](internal/demos/demo1x/) 获取全面的示例，包括：

- 成功事务
- 业务逻辑错误处理
- 事务回滚行为
- 上下文超时处理
- 不同错误场景

运行演示：

```bash
cd internal/demos/demo1x
go run main.go
```

## 测试

```bash
# 运行所有测试
go test -v ./...

# 运行特定测试
go test -v -run TestTransactionSuccess
```

<!-- TEMPLATE (ZH) BEGIN: STANDARD PROJECT FOOTER -->
<!-- VERSION 2025-08-28 08:33:43.829511 +0000 UTC -->

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
- 💬 **意见反馈？** 欢迎所有建议和宝贵意见

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
11. **PR**：在 GitHub 上打开 Pull Request（在 GitHub 网页上）并提供详细描述

请确保测试通过并包含相关的文档更新。

---

## 🌟 项目支持

非常欢迎通过提交 Pull Request 和报告问题来为此项目做出贡献。

**项目支持：**

- ⭐ **给予星标**如果项目对您有帮助
- 🤝 **分享项目**给团队成员和（golang）编程朋友
- 📝 **撰写博客**关于开发工具和工作流程 - 我们提供写作支持
- 🌟 **加入生态** - 致力于支持开源和（golang）开发场景

**使用这个包快乐编程！** 🎉

<!-- TEMPLATE (ZH) END: STANDARD PROJECT FOOTER -->

---

## GitHub 标星点赞

[![Stargazers](https://starchart.cc/orzkratos/gormkratos.svg?variant=adaptive)](https://starchart.cc/orzkratos/gormkratos)