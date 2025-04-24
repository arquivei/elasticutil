package v7

import (
	"context"
	"time"

	"github.com/arquivei/foundationkit/contextmap"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func enrichLogWithIndexes(ctx context.Context, indexes []string) {
	log.Ctx(ctx).UpdateContext(func(zc zerolog.Context) zerolog.Context {
		return zc.Strs("elastic_indexes", indexes)
	})

	contextmap.Ctx(ctx).Set("elastic_indexes", indexes)
}

func enrichLogWithQuery(ctx context.Context, query string) {
	log.Ctx(ctx).UpdateContext(func(zc zerolog.Context) zerolog.Context {
		return zc.Str("elastic_query", truncate(query, 3000))
	})
}

func enrichLogWithTook(ctx context.Context, took int) {
	log.Ctx(ctx).UpdateContext(func(zc zerolog.Context) zerolog.Context {
		return zc.Dur("elastic_took_internal", time.Duration(took)*time.Millisecond)
	})
}

func enrichLogWithShards(ctx context.Context, shards int) {
	log.Ctx(ctx).UpdateContext(func(zc zerolog.Context) zerolog.Context {
		return zc.Int("elastic_shards", shards)
	})

	contextmap.Ctx(ctx).Set("elastic_shards", shards)
}

func truncate(str string, size int) string {
	if size <= 0 {
		return ""
	}

	runes := []rune(str)
	if len(runes) <= size {
		return str
	}

	return string(runes[:size])
}
