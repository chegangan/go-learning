# **🚀 项目简介**
你好！欢迎来到我的 Golang 学习项目。这是我用于练习与实践golang原生语言特性的项目，以后还会慢慢扩充。目前有go-checker和learn-gohttp两个项目，每个项目都经过了多个版本的迭代，从简单到复杂，逐步增加功能与优化架构。每个版本的代码中都写了详细的注释，记录下新增的知识点，以便于反复学习巩固。


## **Go-Checker: 并发网站状态检查工具**
Go-Checker 是一个使用 Go 语言从零开始构建的命令行工具，用于并发地检查大量网站的 HTTP 状态和响应延迟。

这个项目的诞生源于一个核心的学习目标：**通过实战深入理解并掌握 Go 语言的并发模型**。它从一个简单的单线程检查器开始，通过六个阶段的迭代，逐步演进为一个健壮、高效且经过良好测试的并发应用。

通过这个项目，我着重学习了以下 Go 语言的核心能力：

* **并发编程**: 从 Goroutine \+ WaitGroup 到 Channel 通信，再到最终的 **Worker Pool（工作池）** 模式。  
* **标准库应用**: 深度使用 net/http, os, flag, sync, time, testing 等标准库，不依赖任何第三方框架。  
* **软件测试**: 编写**单元测试**（使用 net/http/httptest 模拟服务器）和**性能基准测试**，用数据驱动开发和优化。  
* **命令行工具开发**: 使用 flag 包构建用户友好的命令行界面。  
* **代码演进与重构**: 清晰地展示了如何将一个简单的脚本逐步重构为一个结构良好、易于维护的程序。

### **✨ 功能特性**

* **高并发检查**: 利用 Goroutine 并发发送 HTTP 请求，极大缩短了批量检查所需的时间。  
* **并发控制**: 实现了 Worker Pool 模式，可以自定义并发数量，防止资源耗尽。  
* **命令行驱动**: 通过命令行参数指定URL文件路径和并发数，灵活易用。  
* **超时处理**: 为每个 HTTP 请求设置独立的超时时间，避免因单个网站无响应而阻塞整个程序。  
* **格式化报告**: 使用 text/tabwriter 输出对齐工整的检查结果表格。  
* **统计摘要**: 在检查结束后，提供成功/失败数量、平均响应延迟等统计信息。  
* **经过测试**: 包含单元测试和性能基准测试，保证核心逻辑的正确性和高效性。

### **🌱 项目的演进之旅**

这个项目分为六个清晰的阶段，每个阶段都引入了新的概念并解决了前一阶段的不足。

#### **阶段一: 单线程基线版**

这个版本是一个最简单的顺序执行检查器

* **实现**: 使用 for 循环遍历 URL 列表，依次调用核心检查函数 checkURL。  
* **知识点**: flag 包（解析命令行参数）、os 和 bufio 包（读取文件）、net/http 包（发起 GET 请求）、time 包（计算延迟）。

#### **阶段二: 初识并发 (Goroutine \+ WaitGroup)**

为了解决单线程执行效率低下的问题，此版本首次引入了并发。

* **实现**: 为每个 URL 启动一个独立的 Goroutine 进行检查。使用 sync.WaitGroup 来确保 main 函数等待所有 Goroutine 执行完毕后再退出。  
* **知识点**: go 关键字、sync.WaitGroup 的使用 (Add, Done, Wait)、Goroutine 闭包陷阱。  
* **效果**: 检查总耗时显著缩短，大约等于耗时最长的单个请求的时间。

#### **阶段三: Go 的哲学 (Channel 通信)**

直接在 Goroutine 中打印结果不利于后续处理。此版本引入 Channel 来安全地在 Goroutine 之间传递结果，避免数据竞争。

* **实现**: 创建一个 chan CheckResult。每个 Goroutine 将检查结果发送到 Channel 中，主 Goroutine 从 Channel 中接收所有结果。  
* **知识点**: make 创建 Channel、发送 (\<-) 与接收 (\<-) 操作、close 关闭 Channel、使用 for range 优雅地遍历 Channel。  
* **优势**: 实现了“任务执行”与“结果收集”的解耦，代码更安全、更符合 Go 的设计哲学。

#### **阶段四: 精细化控制 (Worker Pool 模式)**

无限制地创建 Goroutine 会有耗尽系统资源的风险。此版本通过实现“工作池”模式来控制并发的粒度。

* **实现**: 创建一个任务 Channel (jobs) 和一个结果 Channel (results)。启动固定数量的 worker Goroutine，它们从 jobs Channel 中获取任务，并将结果发送到 results Channel。  
* **知识点**: Worker Pool 设计模式，通过 Channel 实现任务分发和负载均衡。  
* **优势**: 有效控制了并发数量，使程序在处理海量任务时依然保持稳定和高效。

#### **阶段五: 锦上添花 (格式化输出与统计)**

为了让工具更专业、用户体验更好，此版本对输出进行了美化。

* **实现**: 在收集完所有结果后，使用 text/tabwriter 包将结果打印为对齐的表格。计算并展示成功/失败数量和平均延迟等统计数据。

#### **阶段六: 灵魂升华 (测试与基准)**

此版本为项目添加了完整的测试体系。

* **实现**:  
  1. **单元测试**: 使用 net/http/httptest 创建模拟服务器来测试 checkURL 函数，使测试不依赖外部网络。  
  2. **竞争检测**: 运行 go test \-race 来证明基于 Channel 的并发模型是数据安全的。  
  3. **性能基准测试**: 编写 Benchmark 函数，用数据清晰地展示了并发版本相对于单线程版本的巨大性能提升。  
* **知识点**: testing 包、httptest 包、go test 和 go test \-bench 命令的使用。

### **🛠️ 如何使用**

1. **克隆仓库**  
   git clone \[你的仓库地址\]  
   cd go-checker

2. 准备 URL 列表  
   在项目根目录下创建一个名为 urls.txt 的文件，每行放置一个需要检查的网址。例如：  
   \[https://www.google.com\](https://www.google.com)  
   \[https://www.baidu.com\](https://www.baidu.com)  
   \[https://github.com\](https://github.com)  
   \[https://this-is-an-invalid-url.com\](https://this-is-an-invalid-url.com)

3. **运行程序**  
   * **默认运行 (使用 urls.txt 和 10 个并发)**:  
     go run main.go

   * 指定文件和并发数:  
     使用 \-file 参数指定文件路径，使用 \-c 参数控制并发的 worker 数量。  
     \# 使用 my\_urls.txt 文件和 5 个并发 worker  
     go run main.go \-file=my\_urls.txt \-c=5

4. **运行测试**  
   * **运行单元测试**:  
     go test

   * **运行性能基准测试**:  
     go test \-bench=.

感谢你的时间和关注！这个项目是我对 Go 语言并发编程的一次深度探索与记录






## **Go 原生 HTTP 服务器学习之旅 (Go Native HTTP Server Learning Journey)**

这个项目旨在记录我使用 Go 语言**原生 net/http 包**，从零开始逐步构建一个功能完善的 Web 服务器的全过程。项目从一个最简单的 "Hello World" 服务器开始，通过四个版本的迭代，最终演进为一个结构清晰、功能模块化的应用程序。

创建这个项目的目的是为了深入学习和展示以下 Go Web 开发的核心概念：

* 路由处理与 ServeMux 的使用  
* RESTful API 设计与实现 (GET, POST)  
* JSON 数据的处理（编码与解码）  
* 中间件（Middleware）的应用（日志、认证）  
* 数据库集成 (MySQL)  
* 用户认证系统（密码哈希与 JWT）  
* 项目结构演进与代码重构

### **🚀 版本演进**

本项目分为四个主要版本，每个版本都在前一个版本的基础上引入了新的功能和概念。

#### **V1: 基础入门 (version1.go)**

第一个版本是所有 Web 应用的起点。它专注于 net/http 包最基础的功能。

**💡 核心功能与知识点：**

* **启动服务器**: 使用 http.ListenAndServe 启动一个 HTTP 服务器。  
* **基础路由**: 使用 http.HandleFunc 为不同路径 (/, /hello, /greet) 注册处理器函数。  
* **处理请求参数**: 从 URL 查询字符串中获取参数 (r.URL.Query().Get("name"))。  
* **返回 JSON**: 定义 struct 结构体，并使用 encoding/json 将其编码为 JSON 格式返回。  
* **静态文件服务**: 使用 http.FileServer 和 http.StripPrefix 托管 static 目录下的静态资源（如 HTML, CSS, JS 文件）。

这个版本的目标是熟悉 Go 如何处理最基本的 HTTP 请求和响应。

#### **V2: 构建 RESTful API (version2.go)**

在 V1 的基础上，这个版本引入了更复杂的 API 交互，并开始关注代码的组织和并发安全。

**💡 核心功能与知识点：**

* **RESTful API**: 实现了一个 /users 接口，支持 GET（获取所有用户）和 POST（创建新用户）方法。  
* **请求体解析**: 使用 json.NewDecoder(r.Body).Decode() 解析 POST 请求体中的 JSON 数据。  
* **使用 http.ServeMux**: 创建了一个自定义的 ServeMux 实例来管理路由，而不是使用全局默认的路由器。这使得路由管理更加清晰和模块化。  
* **并发安全**: 使用 sync.RWMutex（读写锁）来保护内存中的用户数据（模拟数据库），确保在并发请求下的数据一致性。  
* **返回 HTTP 状态码**: 返回更精确的 HTTP 状态码，如 http.StatusCreated (201) 和 http.StatusConflict (409)。

这个版本标志着从一个简单的网站后端向一个真正的 API 服务器的转变。

#### **V3: 集成数据库与认证 (version3.go)**

V3 是一个巨大的飞跃，引入了持久化存储、用户认证和中间件，使应用更加接近生产环境的标准。

**💡 核心功能与知识点：**

* **数据库集成**:  
  * 使用 database/sql 包和 go-sql-driver/mysql 驱动连接到 MySQL 数据库。  
  * 实现了用户的持久化存储。  
* **用户认证**:  
  * **密码哈希**: 使用 golang.org/x/crypto/bcrypt 对用户密码进行哈希处理，安全地存储密码。  
  * **JWT (JSON Web Tokens)**: 实现 /auth/register 和 /auth/login 接口。登录成功后，使用 github.com/golang-jwt/jwt/v5 生成 JWT 返回给客户端。  
* **中间件 (Middleware)**:  
  * **链式调用**: 学习了中间件的核心思想——它是一个接收 http.Handler 并返回新 http.Handler 的函数。  
  * **日志中间件**: 记录每个收到的请求。  
  * **计时中间件**: 记录每个请求的处理耗时。  
  * **认证中间件**: jwtAuthMiddleware 用于保护需要认证才能访问的 API 路由，它会校验请求头中的 JWT。  
* **模块化路由**: 使用多个 ServeMux 来组织不同功能的路由（如 authMux 和 apiMux），使路由结构更加清晰。  
* **自定义 http.Server**: 配置了自定义的 http.Server，可以设置读取/写入超时等参数，增强了服务器的健壮性。

此时，所有功能都还在一个文件中，代码开始变得臃肿，为 V4 的重构埋下了伏笔。

#### **V4: 项目结构重构 (version4/)**

最后一个版本专注于代码的组织和工程化。我们将 V3 的单体文件拆分成一个清晰、可维护、可扩展的项目结构。

**💡 核心功能与知识点：**

* **分层架构**: 项目被拆分为逻辑清晰的几个部分：  
  * cmd/api/main.go: 程序入口，负责组装和启动服务器。  
  * internal/model: 定义核心的数据结构 (structs)。  
  * internal/store: 数据存储层，负责所有与数据库的交互。  
  * internal/handler: HTTP 处理层，负责解析请求、调用 store 并返回响应。  
  * internal/auth: 存放认证相关的逻辑（如 JWT 生成）。  
  * internal/middleware: (在此版本中合并到 handler，也可以独立) 存放中间件。  
* **依赖注入 (Dependency Injection)**: 在 main.go 中创建依赖实例（如数据库连接、store 实例），然后将它们注入到需要它们的模块中（如将 UserStore 注入到 AuthHandler）。这大大降低了代码的耦合度。  
* **高内聚，低耦合**: 每个包都有明确的职责，修改一个模块不会轻易影响到其他模块，使得代码更易于测试和维护。

这个版本是项目成熟的标志，展示了如何将一个原型应用重构为一个结构良好的 Go 项目。

### **📦 如何运行**

在运行项目之前，请确保你已经安装了 Go (1.18+) 和 MySQL。

1. **克隆仓库**  
   git clone \[你的仓库地址\]  
   cd \[你的项目目录\]

2. **运行不同版本**  
   * **运行 V1, V2, 或 V3:**  
     go run version1.go  
     \# 或者  
     go run version2.go  
     \# 或者  
     go run version3.go

   * **运行 V4:**  
     go run version4/cmd/api/[main.go](http://main.go)  
   * V3和V4版本需要先使用文件夹中的sql语句进行mysql建表

3. 访问服务器  
   服务器将在 http://localhost:8000 启动。你可以使用 curl 或 Postman 等工具来测试 API 接口。

