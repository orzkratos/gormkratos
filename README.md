# gormkratos

GORM transaction wrapper for Kratos framework with dual-error-return pattern.

---

<!-- TEMPLATE (EN) BEGIN: LANGUAGE NAVIGATION -->
## CHINESE README

[中文说明](README.zh.md)
<!-- TEMPLATE (EN) END: LANGUAGE NAVIGATION -->

## Key Features

🎯 **Dual-Error Pattern**: Distinguishes business logic errors from database transaction errors  
⚡ **Context Support**: Built-in context timeout and cancellation handling  
🔄 **Automatic Rollback**: Transaction rollback on business logic errors  
🌍 **Kratos Integration**: Seamless integration with Kratos microservice framework  
📋 **Simple API**: Clean, easy-to-use transaction wrapper functions

## Install

```bash
go install github.com/orzkratos/gormkratos@latest
```

## Usage

### Basic Transaction

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
            return errorspb.ErrorServerDbError("failed to create user: %v", err)
        }
        return nil
    })
    
    if err != nil {
        if erk != nil {
            // Business logic error
            return erk
        }
        // Database transaction error
        return fmt.Errorf("transaction failed: %w", err)
    }
    return nil
}
```

### Business Layer Wrapper

```go
// Wrap gormkratos.Transaction for business layer usage
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

// Usage
func BusinessOperation(ctx context.Context, db *gorm.DB) *errkratos.Erk {
    return Transaction(ctx, db, func(db *gorm.DB) *errkratos.Erk {
        // Your business logic here
        return nil
    })
}
```

### With Transaction Options

```go
import "database/sql"

erk, err := gormkratos.Transaction(ctx, db, func(db *gorm.DB) *errors.Error {
    // Your transaction logic
    return nil
}, &sql.TxOptions{
    Isolation: sql.LevelReadCommitted,
    ReadOnly:  false,
})
```

## Error Handling

The `gormkratos.Transaction` function returns two errors to help distinguish between different error types:

1. **Business Logic Error** (`erk *errors.Error`): Kratos framework errors from your business logic
2. **Database Transaction Error** (`err error`): Low-level database or transaction errors

### Error Scenarios

- **Success**: `erk = nil, err = nil`
- **Business Error**: `erk != nil, err != nil` (transaction rollback triggered)
- **Database Error**: `erk = nil, err != nil` (database-level issue)

## Examples

Check out the [demo](internal/demos/demo1x/) for comprehensive examples showing:

- Successful transactions
- Business logic error handling
- Transaction rollback behavior
- Context timeout handling
- Different error scenarios

Run the demo:

```bash
cd internal/demos/demo1x
go run main.go
```

## Testing

```bash
# Run all tests
go test -v ./...

# Run specific test
go test -v -run TestTransactionSuccess
```

<!-- TEMPLATE (EN) BEGIN: STANDARD PROJECT FOOTER -->
<!-- VERSION 2025-08-28 08:33:43.829511 +0000 UTC -->

## 📄 License

MIT License. See [LICENSE](LICENSE).

---

## 🤝 Contributing

Contributions are welcome! Report bugs, suggest features, and contribute code:

- 🐛 **Found a bug?** Open an issue on GitHub with reproduction steps
- 💡 **Have a feature idea?** Create an issue to discuss the suggestion
- 📖 **Documentation confusing?** Report it so we can improve
- 🚀 **Need new features?** Share your use cases to help us understand requirements
- ⚡ **Performance issue?** Help us optimize by reporting slow operations
- 🔧 **Configuration problem?** Ask questions about complex setups
- 📢 **Follow project progress?** Watch the repo for new releases and features
- 🌟 **Success stories?** Share how this package improved your workflow
- 💬 **General feedback?** All suggestions and comments are welcome

---

## 🔧 Development

New code contributions, follow this process:

1. **Fork**: Fork the repo on GitHub (using the webpage interface).
2. **Clone**: Clone the forked project (`git clone https://github.com/yourname/repo-name.git`).
3. **Navigate**: Navigate to the cloned project (`cd repo-name`)
4. **Branch**: Create a feature branch (`git checkout -b feature/xxx`).
5. **Code**: Implement your changes with comprehensive tests
6. **Testing**: (Golang project) Ensure tests pass (`go test ./...`) and follow Go code style conventions
7. **Documentation**: Update documentation for user-facing changes and use meaningful commit messages
8. **Stage**: Stage changes (`git add .`)
9. **Commit**: Commit changes (`git commit -m "Add feature xxx"`) ensuring backward compatible code
10. **Push**: Push to the branch (`git push origin feature/xxx`).
11. **PR**: Open a pull request on GitHub (on the GitHub webpage) with detailed description.

Please ensure tests pass and include relevant documentation updates.

---

## 🌟 Support

Welcome to contribute to this project by submitting pull requests and reporting issues.

**Project Support:**

- ⭐ **Give GitHub stars** if this project helps you
- 🤝 **Share with teammates** and (golang) programming friends
- 📝 **Write tech blogs** about development tools and workflows - we provide content writing support
- 🌟 **Join the ecosystem** - committed to supporting open source and the (golang) development scene

**Happy Coding with this package!** 🎉

<!-- TEMPLATE (EN) END: STANDARD PROJECT FOOTER -->

---

## GitHub Stars

[![Stargazers](https://starchart.cc/orzkratos/gormkratos.svg?variant=adaptive)](https://starchart.cc/orzkratos/gormkratos)
