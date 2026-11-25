package repo_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/internal/agent/repo"
	"github.com/paroki/domus/testutil"
	"github.com/stretchr/testify/require"
)

func makeVec(idx int) []float32 {
	v := make([]float32, repo.EmbeddingDimensions)
	v[idx] = 1.0
	return v
}

func TestSessionAndMessageIsolationAndKeyset(t *testing.T) {
	ctx := context.Background()
	db, sqlDB, _ := testutil.NewEntClientWithDB(t)
	r := repo.New(db, sqlDB)

	parokiA := uuid.New()
	parokiB := uuid.New()
	user1 := uuid.New()
	user2 := uuid.New()

	s1, err := r.CreateSession(ctx, parokiA, user1, "Sesi 1", nil, nil)
	require.NoError(t, err)
	s2, err := r.CreateSession(ctx, parokiA, user1, "Sesi 2", nil, nil)
	require.NoError(t, err)
	s3, err := r.CreateSession(ctx, parokiA, user1, "Sesi 3", nil, nil)
	require.NoError(t, err)
	_, err = r.CreateSession(ctx, parokiA, user2, "User 2", nil, nil)
	require.NoError(t, err)
	_, err = r.CreateSession(ctx, parokiB, user1, "Parish B", nil, nil)
	require.NoError(t, err)

	// Isolation check
	_, err = r.GetSession(ctx, parokiA, user2, s1.ID)
	require.ErrorIs(t, err, domain.ErrSessionNotFound)
	_, err = r.GetSession(ctx, parokiB, user1, s1.ID)
	require.ErrorIs(t, err, domain.ErrSessionNotFound)

	// Keyset pagination
	page1, hasMore, err := r.SearchSessions(ctx, parokiA, user1, "", nil, nil, 2)
	require.NoError(t, err)
	require.True(t, hasMore)
	require.Len(t, page1, 2)
	require.Equal(t, s3.ID, page1[0].ID)
	require.Equal(t, s2.ID, page1[1].ID)

	page2, hasMore, err := r.SearchSessions(ctx, parokiA, user1, "", &page1[1].LastMessageAt, &page1[1].ID, 2)
	require.NoError(t, err)
	require.False(t, hasMore)
	require.Len(t, page2, 1)
	require.Equal(t, s1.ID, page2[0].ID)

	// AppendMessages & ListMessages
	msgs, err := r.AppendMessages(ctx, parokiA, user1, s1.ID, []domain.AppendMessageInput{
		{Role: "user", Content: "Halo Gabriel"},
		{Role: "assistant", AgentName: "gabriel", Content: "Halo! Ada yang bisa dibantu?"},
	})
	require.NoError(t, err)
	require.Len(t, msgs, 2)

	listed, err := r.ListMessages(ctx, parokiA, user1, s1.ID, 50)
	require.NoError(t, err)
	require.Len(t, listed, 2)
	require.Equal(t, "Halo Gabriel", listed[0].Content)

	// SoftDeleteSession hides session
	require.NoError(t, r.SoftDeleteSession(ctx, parokiA, user1, s1.ID))
	_, err = r.GetSession(ctx, parokiA, user1, s1.ID)
	require.ErrorIs(t, err, domain.ErrSessionNotFound)
}

func TestMemorySearch_BM25_RRF_And_Isolation(t *testing.T) {
	ctx := context.Background()
	db, sqlDB, _ := testutil.NewEntClientWithDB(t)
	require.NoError(t, repo.EnsureMemorySearchSchema(ctx, sqlDB))

	r := repo.New(db, sqlDB)
	parokiA := uuid.New()
	parokiB := uuid.New()
	user1 := uuid.New()
	user2 := uuid.New()

	m1, err := r.CreateMemory(ctx, domain.CreateMemoryInput{
		ParokiID: parokiA,
		UserID:   user1,
		Category: domain.MemoryCategoryPreference,
		Content:  "Jadwal misa pagi setiap hari pukul 06:00 WIB di gereja utama",
	}, makeVec(0))
	require.NoError(t, err)

	m2, err := r.CreateMemory(ctx, domain.CreateMemoryInput{
		ParokiID: parokiA,
		UserID:   user1,
		Category: domain.MemoryCategoryFact,
		Content:  "Koordinator koor lingkungan St. Lukas adalah Ibu Maria",
	}, makeVec(1))
	require.NoError(t, err)

	m3, err := r.CreateMemory(ctx, domain.CreateMemoryInput{
		ParokiID: parokiA,
		UserID:   user1,
		Category: domain.MemoryCategoryContext,
		Content:  "Rapat Dewan Paroki bulan depan membahas anggaran",
	}, makeVec(2))
	require.NoError(t, err)

	// Other user / parish memories must never leak
	_, err = r.CreateMemory(ctx, domain.CreateMemoryInput{
		ParokiID: parokiA,
		UserID:   user2,
		Category: domain.MemoryCategoryFact,
		Content:  "Jadwal misa user lain",
	}, makeVec(0))
	require.NoError(t, err)
	_, err = r.CreateMemory(ctx, domain.CreateMemoryInput{
		ParokiID: parokiB,
		UserID:   user1,
		Category: domain.MemoryCategoryFact,
		Content:  "Jadwal misa paroki lain",
	}, makeVec(0))
	require.NoError(t, err)

	// 1. q == "" keyset pagination
	page1, hasMore, err := r.SearchMemories(ctx, parokiA, user1, "", "", nil, nil, nil, 2)
	require.NoError(t, err)
	require.True(t, hasMore)
	require.Len(t, page1, 2)
	require.Equal(t, m3.ID, page1[0].ID)
	require.Equal(t, m2.ID, page1[1].ID)

	page2, hasMore, err := r.SearchMemories(ctx, parokiA, user1, "", "", nil, &page1[1].CreatedAt, &page1[1].ID, 2)
	require.NoError(t, err)
	require.False(t, hasMore)
	require.Len(t, page2, 1)
	require.Equal(t, m1.ID, page2[0].ID)

	// 2. q != "" BM25-only search with Tantivy-special characters (: ( ) AND OR)
	bm25Res, hasMore, err := r.SearchMemories(ctx, parokiA, user1, "misa: (pagi) AND gereja OR koor", "", nil, nil, nil, 10)
	require.NoError(t, err)
	require.False(t, hasMore)
	require.NotEmpty(t, bm25Res)
	for _, m := range bm25Res {
		require.Equal(t, parokiA, m.ParokiID)
		require.Equal(t, user1, m.UserID)
	}

	// 3. q != "" RRF (k=60) hybrid search combining BM25 + halfvec(768) cosine similarity
	rrfRes, hasMore, err := r.SearchMemories(ctx, parokiA, user1, "jadwal misa", "", makeVec(0), nil, nil, 10)
	require.NoError(t, err)
	require.False(t, hasMore)
	require.NotEmpty(t, rrfRes)
	require.Equal(t, m1.ID, rrfRes[0].ID)
	require.Greater(t, rrfRes[0].Score, 0.0)
}
