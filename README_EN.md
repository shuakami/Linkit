<p align="center">
  <img src="images/image.png" alt="Linkit Logo">
</p>

<h1>Linkit</h1>

<p align="center">
<a href="README.md">简体中文</a> | English
</p>

<p align="center">
<img src="https://img.shields.io/badge/Go-1.21%2B-007ACC" alt="Go Version">
<img src="https://img.shields.io/badge/License-AGPL--3.0-blue" alt="License">
<img src="https://img.shields.io/badge/build-passing-44CC11" alt="Build Status">
<a href="https://redocly.github.io/redoc/?url=https://raw.githubusercontent.com/shuakami/linkit/master/docs/api.yaml"><img src="https://img.shields.io/badge/API-Documentation-2ea44f" alt="API Docs"></a>
</p>

A high-performance URL shortening service system developed in Go, implemented using Domain-Driven Design (DDD) and Clean Architecture principles.

It not only provides basic URL shortening functionality but also supports intelligent redirection and detailed access statistics to help businesses better manage and analyze link data.

## Product Features

<p align="center">
  <img src="images/other/short_zh.png" alt="Linkit Management" width="800">
</p>

<p align="center">
  <img src="images/other/fast_cn.png" alt="Linkit Performance" width="800">
</p>

## Core Features

- **URL Shortening**: Support for converting long URLs to short ones, with custom short code options.
- **Smart Redirection**: Intelligent redirection based on visitor's device type, geographic location, and other conditions.
- **Access Analytics**: Detailed access statistics including visit counts, sources, device types, etc.
- **Security Management**: Support for link expiration time settings and access count limits to ensure link security.

## Technical Architecture

Linkit uses a modern technology stack to ensure high performance and scalability:

- **Web Framework**: Gin - Lightweight and efficient, suitable for high-concurrency scenarios.
- **Database**: PostgreSQL - Powerful data storage and query capabilities.
- **Cache**: Redis - Multi-level caching strategy to improve system response speed.
- **Architecture Design**: DDD + Clean Architecture - Clear business logic, easy to maintain and extend.

## Quick Start

Follow these steps to quickly start the Linkit service:

1. Clone the project locally:
   ```bash
   git clone https://github.com/shuakami/linkit.git
   cd linkit
   ```

2. Environment preparation:
   - Install Go 1.21+
   - Install PostgreSQL 14+
   - Install Redis 7+

3. Configure the service:
   
   There's a configuration example file `configs/config.example.yaml` in the project root. Make a copy and rename it to `config.yaml`:

   Windows:
   ```cmd
   copy configs\config.example.yaml configs\config.yaml
   ```
   
   Linux/Mac:
   ```bash
   cp configs/config.example.yaml configs/config.yaml
   ```

   Then edit the `config.yaml` file, you mainly need to modify these configurations:
   - Database connection (host, port, user, password, dbname)
   - Redis connection (host, port, password)
   - Short link domain (domain)

4. Start the service:
   ```bash
   # Download dependencies
   go mod download
   
   # Initialize database
   go run scripts/migrate.go
   
   # Start service
   go run cmd/main.go
   ```

After the service starts, you can:
1. Visit https://redocly.github.io/redoc/?url=https://raw.githubusercontent.com/shuakami/linkit/master/docs/api.yaml to view the online API documentation
2. Or directly check the `docs/api.yaml` file in the project to understand the API details

> If you find this project helpful, please consider giving it a star! 