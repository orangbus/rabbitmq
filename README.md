# An Rabbitmq For A Goravel Extend Package
快速使用 RabbitMQ

## 安装
1、install
```bash
go get -u github.com/orangbus/rabbitmq
```
2、Register service provider
```go
// config/app.go
import "github.com/orangbus/rabbitmq"

"providers": []foundation.ServiceProvider{
    &rabbitmq.ServiceProvider{},
},

3、 Publish Configuration
```bash
go run . artisan vendor:publish --package=github.com/orangbus/rabbitmq
```

4、add .env
```env
RABBITMQ_HOST=127.0.0.1
RABBITMQ_PORT=5672
RABBITMQ_USERNAME=admin
RABBITMQ_PASSWPRD
RABBITMQ_VHOST=bthost
```
使用方式查看 `test/app_test.go`

在 RabbitMQ 的 Topic 交换机中，`a.*.#` 和 `a.#` 都使用了通配符来匹配路由键（Routing Key），但它们匹配的行为略有不同：

- `a.*.#`:
  这个绑定会匹配以 `a.` 开头，后面跟着一个单词，然后是任意数量的单词的路由键。
  - `*` 匹配一个单词，`#` 匹配零个或多个单词。
  例如，它会匹配 `a.b`、`a.b.c`、`a.b.c.d` 等，但不会匹配 `a`（因为后面没有单词）或 `a.b.c.d.e`（因为 `*` 只匹配一个单词，而 `#` 从 `b` 开始匹配）。

- `a.#`:
  这个绑定会匹配以 `a.` 开头，后面跟着任意数量的单词的路由键。
  - `#` 匹配零个或多个单词，这意味着它可以匹配 `a`（后面没有单词）、`a.b`、`a.b.c` 等。
  例如，它会匹配 `a`、`a.b`、`a.b.c`、`a.b.c.d` 等。

- **区别:**
  - `a.*.#` 要求在 `a.` 之后至少有一个单词，然后可以跟任意数量的单词。
  - `a.#` 则允许在 `a.` 之后没有单词，或者有一个单词，或者有多个单词。

- **例子:**
  - 如果路由键是 `a.b.c`:
    - `a.*.#` 会匹配，因为 `b` 是 `a.` 之后的一个单词，后面跟着 `.c`。
    - `a.#` 也会匹配，因为 `b.c` 是 `a.` 之后的任意数量的单词。
  - 如果路由键是 `a`:
    - `a.*.#` 不会匹配，因为 `*` 需要至少一个单词。
    - `a.#` 会匹配，因为 `#` 可以匹配零个或多个单词。
  - 如果路由键是 `a.b.c.d.e`:
    - `a.*.#` 会匹配，因为 `b` 是 `a.` 之后的一个单词，`.c.d.e` 被 `#` 匹配。
    - `a.#` 也会匹配，因为 `b.c.d.e` 是 `a.` 之后的任意数量的单词。

- **选择使用 `a.*.#` 还是 `a.#` 取决于你希望路由键的模式如何匹配。**
  - 如果你需要确保 `a.` 之后至少有一个单词，使用 `a.*.#`。
  - 如果你希望 `a.` 之后可以没有单词，使用 `a.#`。
