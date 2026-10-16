package docs

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Mode string

const (
	ModeHybrid  Mode = "hybrid"
	ModeVector  Mode = "vector"
	ModeKeyword Mode = "keyword"
)

const (
	candidates = 40
	rrfK       = 60
	upsertRows = 500
)

type Result struct {
	Title      string  `json:"title"`
	URL        string  `json:"url"`
	Text       string  `json:"text"`
	Similarity float64 `json:"similarity"`
}

type Postgres struct {
	Pool *pgxpool.Pool
}

func (p Postgres) Hashes(ctx context.Context) (map[string]string, error) {
	rows, err := p.Pool.Query(ctx, `select id, hash from docs.chunks`)
	if err != nil {
		return nil, fmt.Errorf("read chunk hashes: %w", err)
	}
	hashes := map[string]string{}
	var id, hash string
	_, err = pgx.ForEachRow(rows, []any{&id, &hash}, func() error {
		hashes[id] = hash
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("read chunk hashes: %w", err)
	}
	return hashes, nil
}

func (p Postgres) Upsert(ctx context.Context, chunks []Chunk, vectors [][]float32) error {
	for start := 0; start < len(chunks); start += upsertRows {
		end := min(start+upsertRows, len(chunks))
		batch := &pgx.Batch{}
		for i := start; i < end; i++ {
			c := chunks[i]
			batch.Queue(`insert into docs.chunks (id, url, title, text, hash, embedding) values ($1, $2, $3, $4, $5, $6::vector)
				on conflict (id) do update set url = excluded.url, title = excluded.title, text = excluded.text,
				hash = excluded.hash, embedding = excluded.embedding`,
				c.ID, c.URL, c.Title, c.Text, c.Hash, vectorText(vectors[i]))
		}
		if err := p.Pool.SendBatch(ctx, batch).Close(); err != nil {
			return fmt.Errorf("store chunks: %w", err)
		}
	}
	return nil
}

func (p Postgres) Prune(ctx context.Context, keep []string) (int, error) {
	tag, err := p.Pool.Exec(ctx, `delete from docs.chunks where id <> all($1)`, keep)
	if err != nil {
		return 0, fmt.Errorf("prune chunks: %w", err)
	}
	return int(tag.RowsAffected()), nil
}

func (p Postgres) Meta(ctx context.Context, key string) (string, error) {
	var value string
	err := p.Pool.QueryRow(ctx, `select value from docs.meta where key = $1`, key).Scan(&value)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("read docs meta %s: %w", key, err)
	}
	return value, nil
}

func (p Postgres) SetMeta(ctx context.Context, key, value string) error {
	if _, err := p.Pool.Exec(ctx,
		`insert into docs.meta (key, value) values ($1, $2) on conflict (key) do update set value = excluded.value`, key, value,
	); err != nil {
		return fmt.Errorf("write docs meta %s: %w", key, err)
	}
	return nil
}

// The keyword half ORs the query's words: a learner's question rarely has every word on the page that answers it.
func (p Postgres) Search(ctx context.Context, vector []float32, query string, k int, mode Mode, minSimilarity float64) ([]Result, error) {
	vectorLimit, keywordLimit := candidates, candidates
	switch mode {
	case ModeVector:
		keywordLimit = 0
	case ModeKeyword:
		vectorLimit = 0
	}
	rows, err := p.Pool.Query(ctx, `
		with vec as (
			select id, row_number() over (order by distance) as rank
			from (select id, embedding <=> $1::vector as distance from docs.chunks order by distance limit $3) nearest
		), q as (
			select nullif(replace(plainto_tsquery('english', $2)::text, '&', '|'), '')::tsquery as words
		), kw as (
			select id, row_number() over (order by score desc) as rank
			from (select id, ts_rank_cd(tsv, q.words) as score from docs.chunks, q
			      where tsv @@ q.words order by score desc limit $4) matching
		), fused as (
			select id, sum(1.0 / ($5 + rank)) as score from (select * from vec union all select * from kw) ranked group by id
		)
		select c.title, c.url, c.text, 1 - (c.embedding <=> $1::vector) as similarity
		from fused join docs.chunks c using (id)
		where 1 - (c.embedding <=> $1::vector) >= $7
		order by fused.score desc, c.id limit $6`,
		vectorText(vector), query, vectorLimit, keywordLimit, rrfK, k, minSimilarity)
	if err != nil {
		return nil, fmt.Errorf("search the docs: %w", err)
	}
	results, err := pgx.CollectRows(rows, pgx.RowToStructByPos[Result])
	if err != nil {
		return nil, fmt.Errorf("search the docs: %w", err)
	}
	return results, nil
}

func (p Postgres) Empty(ctx context.Context) (bool, error) {
	var empty bool
	if err := p.Pool.QueryRow(ctx, `select not exists (select 1 from docs.chunks)`).Scan(&empty); err != nil {
		return false, fmt.Errorf("count the docs chunks: %w", err)
	}
	return empty, nil
}

func vectorText(v []float32) string {
	var b strings.Builder
	b.WriteByte('[')
	for i, f := range v {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.FormatFloat(float64(f), 'g', -1, 32))
	}
	b.WriteByte(']')
	return b.String()
}

// Ping now, so a missing Postgres fails at start instead of on the first search.
func Connect(ctx context.Context, url string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("open the docs database: %w", err)
	}
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("reach the docs database (make db starts it): %w", err)
	}
	return pool, nil
}
