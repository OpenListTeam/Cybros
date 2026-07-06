package contentaudit

import "cybros/utils"

func contentAuditMessageEntities(entities []utils.MessageEntityInfo) []messageEntityInfo {
	out := []messageEntityInfo{}
	for _, entity := range entities {
		out = append(out, messageEntityInfo{
			Type:     entity.TypeName,
			Text:     entity.Text,
			URL:      entity.URL,
			UserID:   entity.UserID,
			PackName: entity.PackName,
		})
	}
	return out
}

func contentAuditRichTexts(richTexts []utils.RichTextInfo) []richTextInfo {
	out := []richTextInfo{}
	for _, richText := range richTexts {
		if richText.TypeName == "textImage" {
			continue
		}
		out = append(out, richTextInfo{
			Type:     richText.TypeName,
			Text:     richText.Text,
			URL:      richText.URL,
			UserID:   richText.UserID,
			PackName: richText.PackName,
		})
	}
	return out
}
