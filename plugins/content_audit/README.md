# Content Audit

内容审核插件。

当前只处理 `ChannelID` 为 `2573155438` 的 supergroup 新消息。

插件内部从 gotd updates 中提取审核字段：

- 消息文本和 caption。
- 消息 `MessageEntity` 会提取成基础字段。
- `MessageEntityTextURL` 会保留真实 URL。
- 自定义 emoji 会尽量提取所属 pack short name。
- 网页预览中的 `RichText` 会提取成基础字段。
- 图片等媒体内容暂不识别。
- 来源用户 ID、用户名、完整昵称、bot 状态、bio、premium emoji status pack short name。

非目标群消息会在构造内部 `messageInfo` 之前跳过。
