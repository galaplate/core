# Galaplate Core

The core boilerplate package for the [Galaplate](https://github.com/galaplate/galaplate).

This is a Go library/framework package (`github.com/galaplate/core`) that provides the foundation for building web applications. It brings Laravel-like developer experience with Go idiomatic code patterns.

## Stack

- **HTTP Framework**: [Fiber v2](https://gofiber.io/) - Express-inspired web framework for Go
- **Database ORM**: [GORM](https://gorm.io/) - The fantastic ORM library
- **Databases**: MySQL, PostgreSQL, SQLite
- **Logging**: [Logrus](https://github.com/sirupsen/logrus) - Structured logging with file rotation
- **Validation**: [go-playground/validator](https://github.com/go-playground/validator) - Struct validation
- **Scheduling**: [robfig/cron](https://github.com/robfig/cron) - Cron job scheduling

## What is this?

Galaplate Core is the central framework package that applications import and use to access:
- Database migrations and schema building
- Queue and job processing
- Task scheduling (cron)
- Policy-based authorization
- Request validation
- Code generation tools
- Logging and error handling

For the full Galaplate project, visit [https://github.com/galaplate/galaplate](https://github.com/galaplate/galaplate).
