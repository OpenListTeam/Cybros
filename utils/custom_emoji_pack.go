package utils

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"cybros/cache"
	"cybros/consts"

	"github.com/gotd/td/tg"
)

const customEmojiPackCacheTTL = time.Hour

func LoadCustomEmojiPackTitlesAndShortNames(ctx context.Context, api *tg.Client, documentIDs []int64) (map[int64]string, map[int64]string, error) {
	uniqueDocumentIDs := uniqueInt64(documentIDs)
	if len(uniqueDocumentIDs) == 0 {
		return nil, nil, nil
	}
	if api == nil {
		return nil, nil, errors.New(consts.ErrorTelegramAPIUninitialized)
	}

	packTitles := map[int64]string{}
	packShortNames := map[int64]string{}
	missingDocumentIDs := []int64{}
	for _, documentID := range uniqueDocumentIDs {
		cacheKey := strconv.FormatInt(documentID, 10)
		packTitle, titleOK := cache.Get[string]("custom_emoji_pack_title", cacheKey, customEmojiPackCacheTTL)
		packShortName, shortNameOK := cache.Get[string]("custom_emoji_pack_name", cacheKey, customEmojiPackCacheTTL)
		if titleOK && shortNameOK {
			if packShortName != "" {
				packTitles[documentID] = packTitle
				packShortNames[documentID] = packShortName
			}
			continue
		}
		missingDocumentIDs = append(missingDocumentIDs, documentID)
	}
	if len(missingDocumentIDs) == 0 {
		return packTitles, packShortNames, nil
	}

	documents, err := api.MessagesGetCustomEmojiDocuments(ctx, missingDocumentIDs)
	if err != nil {
		return nil, nil, err
	}

	foundDocumentIDs := map[int64]struct{}{}
	for _, documentClass := range documents {
		document, ok := documentClass.(*tg.Document)
		if !ok {
			continue
		}

		for _, attributeClass := range document.Attributes {
			attribute, ok := attributeClass.(*tg.DocumentAttributeCustomEmoji)
			if !ok {
				continue
			}

			packTitle, packShortName, err := customEmojiPackTitleAndShortName(ctx, api, attribute.Stickerset)
			if err != nil {
				return nil, nil, err
			}
			if packShortName == "" {
				continue
			}
			foundDocumentIDs[document.ID] = struct{}{}
			packTitles[document.ID] = packTitle
			packShortNames[document.ID] = packShortName
			cacheKey := strconv.FormatInt(document.ID, 10)
			cache.Set("custom_emoji_pack_title", cacheKey, packTitle)
			cache.Set("custom_emoji_pack_name", cacheKey, packShortName)
		}
	}
	for _, documentID := range missingDocumentIDs {
		_, ok := foundDocumentIDs[documentID]
		if ok {
			continue
		}
		cacheKey := strconv.FormatInt(documentID, 10)
		cache.Set("custom_emoji_pack_title", cacheKey, "")
		cache.Set("custom_emoji_pack_name", cacheKey, "")
	}
	return packTitles, packShortNames, nil
}

func LoadCustomEmojiPackNames(ctx context.Context, api *tg.Client, documentIDs []int64) (map[int64]string, error) {
	uniqueDocumentIDs := uniqueInt64(documentIDs)
	if len(uniqueDocumentIDs) == 0 {
		return nil, nil
	}
	if api == nil {
		return nil, errors.New(consts.ErrorTelegramAPIUninitialized)
	}

	packNames := map[int64]string{}
	missingDocumentIDs := []int64{}
	for _, documentID := range uniqueDocumentIDs {
		packName, ok := cache.Get[string]("custom_emoji_pack_name", strconv.FormatInt(documentID, 10), customEmojiPackCacheTTL)
		if ok {
			if packName != "" {
				packNames[documentID] = packName
			}
			continue
		}
		missingDocumentIDs = append(missingDocumentIDs, documentID)
	}
	if len(missingDocumentIDs) == 0 {
		return packNames, nil
	}

	documents, err := api.MessagesGetCustomEmojiDocuments(ctx, missingDocumentIDs)
	if err != nil {
		return nil, err
	}

	foundDocumentIDs := map[int64]struct{}{}
	for _, documentClass := range documents {
		document, ok := documentClass.(*tg.Document)
		if !ok {
			continue
		}

		for _, attributeClass := range document.Attributes {
			attribute, ok := attributeClass.(*tg.DocumentAttributeCustomEmoji)
			if !ok {
				continue
			}

			packName, err := customEmojiPackShortName(ctx, api, attribute.Stickerset)
			if err != nil {
				return nil, err
			}
			if packName == "" {
				continue
			}
			foundDocumentIDs[document.ID] = struct{}{}
			packNames[document.ID] = packName
			cache.Set("custom_emoji_pack_name", strconv.FormatInt(document.ID, 10), packName)
		}
	}
	for _, documentID := range missingDocumentIDs {
		_, ok := foundDocumentIDs[documentID]
		if ok {
			continue
		}
		cache.Set("custom_emoji_pack_name", strconv.FormatInt(documentID, 10), "")
	}
	return packNames, nil
}

func LoadCustomEmojiPackTitleAndShortName(ctx context.Context, api *tg.Client, documentID int64) (string, string, error) {
	packTitles, packShortNames, err := LoadCustomEmojiPackTitlesAndShortNames(ctx, api, []int64{documentID})
	if err != nil {
		return "", "", err
	}
	return packTitles[documentID], packShortNames[documentID], nil
}

func LoadCustomEmojiPackName(ctx context.Context, api *tg.Client, documentID int64) (string, error) {
	packNames, err := LoadCustomEmojiPackNames(ctx, api, []int64{documentID})
	if err != nil {
		return "", err
	}
	return packNames[documentID], nil
}

func customEmojiPackShortName(ctx context.Context, api *tg.Client, pack tg.InputStickerSetClass) (string, error) {
	shortName, ok := pack.(*tg.InputStickerSetShortName)
	if ok {
		return shortName.ShortName, nil
	}
	if pack == nil {
		return "", nil
	}

	return cache.Load[string]("emoji_pack_short_name", inputStickerSetCacheKey(pack), customEmojiPackCacheTTL, func() (string, error) {
		packClass, err := api.MessagesGetStickerSet(ctx, &tg.MessagesGetStickerSetRequest{
			Stickerset: pack,
			Hash:       0,
		})
		if err != nil {
			return "", err
		}

		packValue, ok := packClass.(*tg.MessagesStickerSet)
		if !ok {
			return "", nil
		}
		return packValue.Set.ShortName, nil
	})
}

func customEmojiPackTitleAndShortName(ctx context.Context, api *tg.Client, pack tg.InputStickerSetClass) (string, string, error) {
	if pack == nil {
		return "", "", nil
	}

	cacheKey := inputStickerSetCacheKey(pack)
	packTitle, titleOK := cache.Get[string]("emoji_pack_title", cacheKey, customEmojiPackCacheTTL)
	packShortName, shortNameOK := cache.Get[string]("emoji_pack_short_name", cacheKey, customEmojiPackCacheTTL)
	if titleOK && shortNameOK {
		return packTitle, packShortName, nil
	}

	packClass, err := api.MessagesGetStickerSet(ctx, &tg.MessagesGetStickerSetRequest{
		Stickerset: pack,
		Hash:       0,
	})
	if err != nil {
		return "", "", err
	}

	packValue, ok := packClass.(*tg.MessagesStickerSet)
	if !ok {
		return "", "", nil
	}

	cache.Set("emoji_pack_title", cacheKey, packValue.Set.Title)
	cache.Set("emoji_pack_short_name", cacheKey, packValue.Set.ShortName)
	return packValue.Set.Title, packValue.Set.ShortName, nil
}

func inputStickerSetCacheKey(pack tg.InputStickerSetClass) string {
	packID, ok := pack.(*tg.InputStickerSetID)
	if ok {
		return fmt.Sprintf("id:%d:%d", packID.ID, packID.AccessHash)
	}

	shortName, ok := pack.(*tg.InputStickerSetShortName)
	if ok {
		return "short:" + shortName.ShortName
	}

	return fmt.Sprintf("%s:%s", pack.TypeName(), pack.String())
}

func uniqueInt64(values []int64) []int64 {
	seen := map[int64]struct{}{}
	out := []int64{}
	for _, value := range values {
		if value == 0 {
			continue
		}
		_, ok := seen[value]
		if ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
