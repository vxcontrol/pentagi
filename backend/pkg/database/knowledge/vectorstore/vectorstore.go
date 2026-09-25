package vectorstore

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vxcontrol/langchaingo/embeddings"
	"github.com/vxcontrol/langchaingo/vectorstores/pgvector"
)

func Options(embedder embeddings.Embedder, pool *pgxpool.Pool, databaseURL string) []pgvector.Option {
	options := []pgvector.Option{
		pgvector.WithEmbedder(embedder),
		pgvector.WithCollectionName("langchain"),
	}
	if pool != nil {
		return append(options, pgvector.WithConn(pool))
	}
	return append(options, pgvector.WithConnectionURL(databaseURL))
}
