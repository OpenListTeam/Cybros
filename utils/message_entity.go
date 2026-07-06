package utils

import (
	"unicode/utf16"
	"unicode/utf8"

	"github.com/gotd/td/tg"
)

type MessageEntityInfo struct {
	TypeID     uint32 // gotd TL type ID
	TypeName   string // gotd TL type name
	Offset     int    // UTF-16 offset
	Length     int    // UTF-16 length
	StartByte  int    // 原文起始 byte offset
	EndByte    int    // 原文结束 byte offset
	Text       string // 实体覆盖的原文
	URL        string // URL、email、phone 的可访问值
	UserID     int64  // mentionName 用户 ID
	Language   string // pre 代码语言
	DocumentID int64  // customEmoji document ID
	PackName   string // customEmoji pack short name
	Collapsed  bool   // blockquote 是否折叠
	Date       int    // formattedDate UNIX timestamp
	Relative   bool   // formattedDate relative
	ShortTime  bool   // formattedDate short time
	LongTime   bool   // formattedDate long time
	ShortDate  bool   // formattedDate short date
	LongDate   bool   // formattedDate long date
	DayOfWeek  bool   // formattedDate day of week
	OldText    string // diffReplace 旧文本
}

type messageEntityRange interface {
	GetOffset() int
	GetLength() int
}

func ExtractMessageEntities(text string, entities []tg.MessageEntityClass, customEmojiPackNames map[int64]string) []MessageEntityInfo {
	if text == "" || len(entities) == 0 {
		return nil
	}

	byteOffsets := utf16ByteOffsets(text)
	out := []MessageEntityInfo{}
	for _, entity := range entities {
		entityRange, ok := entity.(messageEntityRange)
		if !ok {
			continue
		}

		start, ok := byteOffsets[entityRange.GetOffset()]
		if !ok {
			continue
		}
		end, ok := byteOffsets[entityRange.GetOffset()+entityRange.GetLength()]
		if !ok || end <= start {
			continue
		}

		entityInfo, ok := messageEntityInfoFromTG(entity, entityRange, start, end, text[start:end], customEmojiPackNames)
		if !ok {
			continue
		}
		out = append(out, entityInfo)
	}

	return out
}

func CustomEmojiDocumentIDsFromMessageEntities(entities []tg.MessageEntityClass) []int64 {
	ids := []int64{}
	for _, entity := range entities {
		customEmoji, ok := entity.(*tg.MessageEntityCustomEmoji)
		if !ok {
			continue
		}
		ids = append(ids, customEmoji.DocumentID)
	}
	return uniqueInt64(ids)
}

func utf16ByteOffsets(text string) map[int]int {
	offsets := map[int]int{0: 0}
	units := 0
	for index, r := range text {
		offsets[units] = index
		runeSize := utf8.RuneLen(r)
		if runeSize < 0 {
			runeSize = 1
		}
		units += utf16.RuneLen(r)
		offsets[units] = index + runeSize
	}
	offsets[units] = len(text)
	return offsets
}

func messageEntityInfoFromTG(entity tg.MessageEntityClass, entityRange messageEntityRange, start int, end int, text string, customEmojiPackNames map[int64]string) (MessageEntityInfo, bool) {
	info := MessageEntityInfo{
		TypeID:    entity.TypeID(),
		TypeName:  entity.TypeName(),
		Offset:    entityRange.GetOffset(),
		Length:    entityRange.GetLength(),
		StartByte: start,
		EndByte:   end,
		Text:      text,
	}

	switch value := entity.(type) {
	case *tg.MessageEntityPre:
		info.Language = value.Language
	case *tg.MessageEntityTextURL:
		info.URL = value.URL
	case *tg.MessageEntityURL:
		info.URL = text
	case *tg.MessageEntityEmail:
		info.URL = "mailto:" + text
	case *tg.MessageEntityPhone:
		info.URL = "tel:" + text
	case *tg.MessageEntityMentionName:
		info.UserID = value.UserID
	case *tg.MessageEntityCustomEmoji:
		info.DocumentID = value.DocumentID
		info.PackName = customEmojiPackNames[value.DocumentID]
	case *tg.MessageEntityBlockquote:
		info.Collapsed = value.Collapsed
	case *tg.MessageEntityFormattedDate:
		info.Date = value.Date
		info.Relative = value.Relative
		info.ShortTime = value.ShortTime
		info.LongTime = value.LongTime
		info.ShortDate = value.ShortDate
		info.LongDate = value.LongDate
		info.DayOfWeek = value.DayOfWeek
	case *tg.MessageEntityDiffReplace:
		info.OldText = value.OldText
	}

	if info.TypeName == "" {
		return MessageEntityInfo{}, false
	}
	return info, true
}
