# Cybros

OpenList 官方群 Telegram UserBot 审核服务。

## 审核策略

本项目的审核策略以安全边界为首要目标。

面对违法违规、黑灰产推广或无法充分确认的可疑内容时，系统倾向于采取更保守的处置方式，优先降低风险进入群组的可能性；因此误判带来的成本不作为主要设计约束。

社区可以直接补充更严格的审核规则、插件策略或风险样本，用于进一步收紧可疑内容的处置边界。

当前仅处理 OpenList 官方群，暂不支持其他群组审核。

监听指定 supergroup 新消息，提取文本、链接和用户信息，交给插件处理。

使用 Go 和 [gotd](https://github.com/gotd/td)。Telegram API 密钥：<https://my.telegram.org/>
