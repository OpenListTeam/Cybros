# Plugins

`plugins` 是插件注册层。

- `interface.go` 定义插件接口。
- `import.go` 统一注册具体插件。
- 具体插件放在独立子目录中。

插件接收 gotd 原始 `tg.UpdatesClass`，由插件自己解析需要的 update 类型和上下文。
