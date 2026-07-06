package utils

import (
	"strings"

	"github.com/gotd/td/tg"
)

type RichTextInfo struct {
	TypeID     uint32 // gotd TL type ID
	TypeName   string // gotd TL type name
	Text       string // 节点提取出的纯文本
	URL        string // textURL 的真实 URL 或 email/phone 可访问值
	Email      string // textEmail email
	Phone      string // textPhone phone
	UserID     int64  // textMentionName 用户 ID
	DocumentID int64  // textCustomEmoji/textImage document ID
	Alt        string // textCustomEmoji alt
	PackName   string // textCustomEmoji pack short name
	WebpageID  int64  // textURL webpage ID
	Name       string // textAnchor name
	Source     string // textMath source
	Date       int    // textDate UNIX timestamp
	Relative   bool   // textDate relative
	ShortTime  bool   // textDate short time
	LongTime   bool   // textDate long time
	ShortDate  bool   // textDate short date
	LongDate   bool   // textDate long date
	DayOfWeek  bool   // textDate day of week
	Width      int    // textImage width
	Height     int    // textImage height
}

type pageBlockTextGetter interface {
	GetText() tg.RichTextClass
}

type pageBlockTitleGetter interface {
	GetTitle() tg.RichTextClass
}

type pageBlockAuthorGetter interface {
	GetAuthor() tg.RichTextClass
}

type pageBlockRichCaptionGetter interface {
	GetCaption() tg.RichTextClass
}

type pageBlockPageCaptionGetter interface {
	GetCaption() tg.PageCaption
}

func ExtractMessageRichTexts(media tg.MessageMediaClass, customEmojiPackNames map[int64]string) []RichTextInfo {
	webPageMedia, ok := media.(*tg.MessageMediaWebPage)
	if !ok {
		return nil
	}

	webPage, ok := webPageMedia.Webpage.(*tg.WebPage)
	if !ok {
		return nil
	}

	page, ok := webPage.GetCachedPage()
	if !ok {
		return nil
	}

	out := []RichTextInfo{}
	for _, block := range page.Blocks {
		appendPageBlockRichTexts(&out, block, customEmojiPackNames)
	}
	return out
}

func ExtractRichText(text tg.RichTextClass, customEmojiPackNames map[int64]string) []RichTextInfo {
	out := []RichTextInfo{}
	appendRichText(&out, text, customEmojiPackNames, false)
	return out
}

func CustomEmojiDocumentIDsFromMessageMedia(media tg.MessageMediaClass) []int64 {
	webPageMedia, ok := media.(*tg.MessageMediaWebPage)
	if !ok {
		return nil
	}

	webPage, ok := webPageMedia.Webpage.(*tg.WebPage)
	if !ok {
		return nil
	}

	page, ok := webPage.GetCachedPage()
	if !ok {
		return nil
	}

	ids := []int64{}
	for _, block := range page.Blocks {
		appendPageBlockCustomEmojiDocumentIDs(&ids, block)
	}
	return uniqueInt64(ids)
}

func appendPageBlockRichTexts(out *[]RichTextInfo, block tg.PageBlockClass, customEmojiPackNames map[int64]string) {
	textGetter, ok := block.(pageBlockTextGetter)
	if ok {
		appendRichText(out, textGetter.GetText(), customEmojiPackNames, false)
	}

	titleGetter, ok := block.(pageBlockTitleGetter)
	if ok {
		appendRichText(out, titleGetter.GetTitle(), customEmojiPackNames, false)
	}

	authorGetter, ok := block.(pageBlockAuthorGetter)
	if ok {
		appendRichText(out, authorGetter.GetAuthor(), customEmojiPackNames, false)
	}

	richCaptionGetter, ok := block.(pageBlockRichCaptionGetter)
	if ok {
		appendRichText(out, richCaptionGetter.GetCaption(), customEmojiPackNames, false)
	}

	pageCaptionGetter, ok := block.(pageBlockPageCaptionGetter)
	if ok {
		appendPageCaptionRichTexts(out, pageCaptionGetter.GetCaption(), customEmojiPackNames)
	}

	switch value := block.(type) {
	case *tg.PageBlockCover:
		appendPageBlockRichTexts(out, value.Cover, customEmojiPackNames)
	case *tg.PageBlockDetails:
		for _, child := range value.Blocks {
			appendPageBlockRichTexts(out, child, customEmojiPackNames)
		}
	case *tg.PageBlockBlockquoteBlocks:
		for _, child := range value.Blocks {
			appendPageBlockRichTexts(out, child, customEmojiPackNames)
		}
	case *tg.PageBlockList:
		appendPageListItemRichTexts(out, value.Items, customEmojiPackNames)
	case *tg.PageBlockOrderedList:
		appendPageListOrderedItemRichTexts(out, value.Items, customEmojiPackNames)
	case *tg.PageBlockTable:
		appendPageTableRichTexts(out, value.Rows, customEmojiPackNames)
	}
}

func appendPageCaptionRichTexts(out *[]RichTextInfo, caption tg.PageCaption, customEmojiPackNames map[int64]string) {
	appendRichText(out, caption.Text, customEmojiPackNames, false)
	appendRichText(out, caption.Credit, customEmojiPackNames, false)
}

func appendPageListItemRichTexts(out *[]RichTextInfo, items []tg.PageListItemClass, customEmojiPackNames map[int64]string) {
	for _, item := range items {
		switch value := item.(type) {
		case *tg.PageListItemText:
			appendRichText(out, value.Text, customEmojiPackNames, false)
		case *tg.PageListItemBlocks:
			for _, block := range value.Blocks {
				appendPageBlockRichTexts(out, block, customEmojiPackNames)
			}
		}
	}
}

func appendPageListOrderedItemRichTexts(out *[]RichTextInfo, items []tg.PageListOrderedItemClass, customEmojiPackNames map[int64]string) {
	for _, item := range items {
		switch value := item.(type) {
		case *tg.PageListOrderedItemText:
			appendRichText(out, value.Text, customEmojiPackNames, false)
		case *tg.PageListOrderedItemBlocks:
			for _, block := range value.Blocks {
				appendPageBlockRichTexts(out, block, customEmojiPackNames)
			}
		}
	}
}

func appendPageTableRichTexts(out *[]RichTextInfo, rows []tg.PageTableRow, customEmojiPackNames map[int64]string) {
	for _, row := range rows {
		for _, cell := range row.Cells {
			appendRichText(out, cell.Text, customEmojiPackNames, false)
		}
	}
}

func appendPageBlockCustomEmojiDocumentIDs(ids *[]int64, block tg.PageBlockClass) {
	textGetter, ok := block.(pageBlockTextGetter)
	if ok {
		appendRichTextCustomEmojiDocumentIDs(ids, textGetter.GetText())
	}

	titleGetter, ok := block.(pageBlockTitleGetter)
	if ok {
		appendRichTextCustomEmojiDocumentIDs(ids, titleGetter.GetTitle())
	}

	authorGetter, ok := block.(pageBlockAuthorGetter)
	if ok {
		appendRichTextCustomEmojiDocumentIDs(ids, authorGetter.GetAuthor())
	}

	richCaptionGetter, ok := block.(pageBlockRichCaptionGetter)
	if ok {
		appendRichTextCustomEmojiDocumentIDs(ids, richCaptionGetter.GetCaption())
	}

	pageCaptionGetter, ok := block.(pageBlockPageCaptionGetter)
	if ok {
		caption := pageCaptionGetter.GetCaption()
		appendRichTextCustomEmojiDocumentIDs(ids, caption.Text)
		appendRichTextCustomEmojiDocumentIDs(ids, caption.Credit)
	}

	switch value := block.(type) {
	case *tg.PageBlockCover:
		appendPageBlockCustomEmojiDocumentIDs(ids, value.Cover)
	case *tg.PageBlockDetails:
		for _, child := range value.Blocks {
			appendPageBlockCustomEmojiDocumentIDs(ids, child)
		}
	case *tg.PageBlockBlockquoteBlocks:
		for _, child := range value.Blocks {
			appendPageBlockCustomEmojiDocumentIDs(ids, child)
		}
	case *tg.PageBlockList:
		appendPageListItemCustomEmojiDocumentIDs(ids, value.Items)
	case *tg.PageBlockOrderedList:
		appendPageListOrderedItemCustomEmojiDocumentIDs(ids, value.Items)
	case *tg.PageBlockTable:
		appendPageTableCustomEmojiDocumentIDs(ids, value.Rows)
	}
}

func appendPageListItemCustomEmojiDocumentIDs(ids *[]int64, items []tg.PageListItemClass) {
	for _, item := range items {
		switch value := item.(type) {
		case *tg.PageListItemText:
			appendRichTextCustomEmojiDocumentIDs(ids, value.Text)
		case *tg.PageListItemBlocks:
			for _, block := range value.Blocks {
				appendPageBlockCustomEmojiDocumentIDs(ids, block)
			}
		}
	}
}

func appendPageListOrderedItemCustomEmojiDocumentIDs(ids *[]int64, items []tg.PageListOrderedItemClass) {
	for _, item := range items {
		switch value := item.(type) {
		case *tg.PageListOrderedItemText:
			appendRichTextCustomEmojiDocumentIDs(ids, value.Text)
		case *tg.PageListOrderedItemBlocks:
			for _, block := range value.Blocks {
				appendPageBlockCustomEmojiDocumentIDs(ids, block)
			}
		}
	}
}

func appendPageTableCustomEmojiDocumentIDs(ids *[]int64, rows []tg.PageTableRow) {
	for _, row := range rows {
		for _, cell := range row.Cells {
			appendRichTextCustomEmojiDocumentIDs(ids, cell.Text)
		}
	}
}

func appendRichTextCustomEmojiDocumentIDs(ids *[]int64, text tg.RichTextClass) {
	if text == nil {
		return
	}

	switch value := text.(type) {
	case *tg.TextConcat:
		for _, child := range value.Texts {
			appendRichTextCustomEmojiDocumentIDs(ids, child)
		}
	case *tg.TextCustomEmoji:
		*ids = append(*ids, value.DocumentID)
	case *tg.TextURL:
		appendRichTextCustomEmojiDocumentIDs(ids, value.Text)
	case *tg.TextEmail:
		appendRichTextCustomEmojiDocumentIDs(ids, value.Text)
	case *tg.TextPhone:
		appendRichTextCustomEmojiDocumentIDs(ids, value.Text)
	case *tg.TextMentionName:
		appendRichTextCustomEmojiDocumentIDs(ids, value.Text)
	case *tg.TextDate:
		appendRichTextCustomEmojiDocumentIDs(ids, value.Text)
	default:
		child, ok := text.(interface {
			GetText() tg.RichTextClass
		})
		if ok {
			appendRichTextCustomEmojiDocumentIDs(ids, child.GetText())
		}
	}
}

func appendRichText(out *[]RichTextInfo, text tg.RichTextClass, customEmojiPackNames map[int64]string, wrapped bool) {
	if text == nil {
		return
	}

	switch value := text.(type) {
	case *tg.TextEmpty:
		return
	case *tg.TextPlain:
		if !wrapped && value.Text != "" {
			*out = append(*out, RichTextInfo{TypeID: value.TypeID(), TypeName: value.TypeName(), Text: value.Text})
		}
	case *tg.TextConcat:
		for _, child := range value.Texts {
			appendRichText(out, child, customEmojiPackNames, wrapped)
		}
	case *tg.TextURL:
		appendRichTextNode(out, RichTextInfo{
			TypeID:    value.TypeID(),
			TypeName:  value.TypeName(),
			Text:      richTextPlain(value.Text),
			URL:       value.URL,
			WebpageID: value.WebpageID,
		})
		appendRichText(out, value.Text, customEmojiPackNames, true)
	case *tg.TextEmail:
		appendRichTextNode(out, RichTextInfo{
			TypeID:   value.TypeID(),
			TypeName: value.TypeName(),
			Text:     richTextPlain(value.Text),
			URL:      "mailto:" + value.Email,
			Email:    value.Email,
		})
		appendRichText(out, value.Text, customEmojiPackNames, true)
	case *tg.TextPhone:
		appendRichTextNode(out, RichTextInfo{
			TypeID:   value.TypeID(),
			TypeName: value.TypeName(),
			Text:     richTextPlain(value.Text),
			URL:      "tel:" + value.Phone,
			Phone:    value.Phone,
		})
		appendRichText(out, value.Text, customEmojiPackNames, true)
	case *tg.TextCustomEmoji:
		appendRichTextNode(out, RichTextInfo{
			TypeID:     value.TypeID(),
			TypeName:   value.TypeName(),
			Text:       value.Alt,
			DocumentID: value.DocumentID,
			Alt:        value.Alt,
			PackName:   customEmojiPackNames[value.DocumentID],
		})
	case *tg.TextImage:
		appendRichTextNode(out, RichTextInfo{
			TypeID:     value.TypeID(),
			TypeName:   value.TypeName(),
			DocumentID: value.DocumentID,
			Width:      value.W,
			Height:     value.H,
		})
	case *tg.TextMath:
		appendRichTextNode(out, RichTextInfo{
			TypeID:   value.TypeID(),
			TypeName: value.TypeName(),
			Text:     value.Source,
			Source:   value.Source,
		})
	case *tg.TextMentionName:
		appendRichTextNode(out, RichTextInfo{
			TypeID:   value.TypeID(),
			TypeName: value.TypeName(),
			Text:     richTextPlain(value.Text),
			UserID:   value.UserID,
		})
		appendRichText(out, value.Text, customEmojiPackNames, true)
	case *tg.TextAnchor:
		appendRichTextNode(out, RichTextInfo{
			TypeID:   value.TypeID(),
			TypeName: value.TypeName(),
			Text:     richTextPlain(value.Text),
			Name:     value.Name,
		})
		appendRichText(out, value.Text, customEmojiPackNames, true)
	case *tg.TextDate:
		appendRichTextNode(out, RichTextInfo{
			TypeID:    value.TypeID(),
			TypeName:  value.TypeName(),
			Text:      richTextPlain(value.Text),
			Date:      value.Date,
			Relative:  value.Relative,
			ShortTime: value.ShortTime,
			LongTime:  value.LongTime,
			ShortDate: value.ShortDate,
			LongDate:  value.LongDate,
			DayOfWeek: value.DayOfWeek,
		})
		appendRichText(out, value.Text, customEmojiPackNames, true)
	default:
		appendWrappedRichText(out, text, customEmojiPackNames)
	}
}

func appendWrappedRichText(out *[]RichTextInfo, text tg.RichTextClass, customEmojiPackNames map[int64]string) {
	child, ok := text.(interface {
		GetText() tg.RichTextClass
	})
	if !ok {
		return
	}

	plainText := richTextPlain(child.GetText())
	if plainText != "" {
		*out = append(*out, RichTextInfo{
			TypeID:   text.TypeID(),
			TypeName: text.TypeName(),
			Text:     plainText,
		})
	}
	appendRichText(out, child.GetText(), customEmojiPackNames, true)
}

func appendRichTextNode(out *[]RichTextInfo, info RichTextInfo) {
	if info.Text == "" && info.URL == "" && info.Email == "" && info.Phone == "" && info.UserID == 0 && info.DocumentID == 0 && info.PackName == "" && info.WebpageID == 0 && info.Name == "" && info.Source == "" && info.Date == 0 && info.Width == 0 && info.Height == 0 {
		return
	}
	*out = append(*out, info)
}

func richTextPlain(text tg.RichTextClass) string {
	if text == nil {
		return ""
	}

	switch value := text.(type) {
	case *tg.TextEmpty:
		return ""
	case *tg.TextPlain:
		return value.Text
	case *tg.TextConcat:
		var builder strings.Builder
		for _, child := range value.Texts {
			builder.WriteString(richTextPlain(child))
		}
		return builder.String()
	case *tg.TextURL:
		return richTextPlain(value.Text)
	case *tg.TextEmail:
		return richTextPlain(value.Text)
	case *tg.TextPhone:
		return richTextPlain(value.Text)
	case *tg.TextCustomEmoji:
		return value.Alt
	case *tg.TextImage:
		return ""
	case *tg.TextMath:
		return value.Source
	case *tg.TextMentionName:
		return richTextPlain(value.Text)
	case *tg.TextDate:
		return richTextPlain(value.Text)
	default:
		child, ok := text.(interface {
			GetText() tg.RichTextClass
		})
		if ok {
			return richTextPlain(child.GetText())
		}
		return ""
	}
}
