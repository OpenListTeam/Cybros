package contentaudit

type messageInfo struct {
	ID                               int                 // 消息 ID
	Text                             string              // 普通文本消息内容；有媒体时留空
	Caption                          string              // 媒体消息附带文字；无媒体时留空
	Entities                         []messageEntityInfo // 消息文本实体
	RichTexts                        []richTextInfo      // 网页预览 RichText 文本
	SourceUserID                     int64               // 发送者用户 ID
	SourceFullNickName               string              // 发送者完整昵称
	SourceUserBio                    string              // 发送者 bio
	SourcePremiumEmojiStatusPackName string              // 发送者 premium emoji status 所在 pack short name
	SourceUserIsBot                  bool                // 发送者是否为 bot
	SourceGroupUsername              string              // 来源群用户名
	SourceUserUsername               string              // 发送者用户名
}

type messageEntityInfo struct {
	Type     string // 实体类型
	Text     string // 实体覆盖的原文
	URL      string // 真实 URL；仅 URL 类实体有值
	UserID   int64  // 提及用户 ID；仅 mentionName 有值
	PackName string // 自定义 emoji pack short name；仅 customEmoji 有值
}

type richTextInfo struct {
	Type     string // RichText 类型
	Text     string // 节点提取出的纯文本
	URL      string // 真实 URL；仅 URL 类节点有值
	UserID   int64  // 提及用户 ID；仅 mentionName 有值
	PackName string // 自定义 emoji pack short name；仅 customEmoji 有值
}
