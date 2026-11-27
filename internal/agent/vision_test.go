package agent_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/paroki/domus/ent/agentprovider"
	"github.com/paroki/domus/internal/agent"
	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/internal/platform/setting"
	"github.com/paroki/domus/internal/user"
	"github.com/paroki/domus/testutil"
	"github.com/stretchr/testify/require"
)

type fakeVisionChatModel struct {
	lastMessages []*schema.Message
	respContent  string
	err          error
}

func (f *fakeVisionChatModel) Generate(_ context.Context, input []*schema.Message, _ ...model.Option) (*schema.Message, error) {
	f.lastMessages = input
	if f.err != nil {
		return nil, f.err
	}
	return &schema.Message{
		Role:    schema.Assistant,
		Content: f.respContent,
	}, nil
}

func (f *fakeVisionChatModel) Stream(_ context.Context, _ []*schema.Message, _ ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	return nil, errors.New("not implemented")
}

func (f *fakeVisionChatModel) WithTools(_ []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	return f, nil
}

func TestExtractIdentityDocument_FakeChatModel(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewEntClient(t)
	store := setting.NewStore(db)

	prov, err := db.AgentProvider.Create().
		SetName("Test Vision Provider").
		SetProvider(agentprovider.ProviderGemini).
		SetAPIKey("test-api-key").
		Save(ctx)
	require.NoError(t, err)

	feat := agent.New(nil, slog.New(slog.DiscardHandler), agent.WithDB(db), agent.WithSettings(store))
	feat.RegisterSettings(store)
	require.NoError(t, store.Load(ctx))

	_, err = store.Update(ctx, map[string]string{
		agent.SettingModelVision: prov.ID.String() + ":gemini-2.5-flash",
	})
	require.NoError(t, err)

	// 1. Happy path with markdown-fenced JSON response and multimodal image parts
	fake := &fakeVisionChatModel{
		respContent: "```json\n{\"ktp\":{\"nik\":\"3174010101900001\",\"nama\":\"Budi Santoso\",\"jenisKelamin\":\"pria\",\"tanggalLahir\":\"1990-01-01\"},\"kk\":{\"nomorKk\":\"3174010101900099\",\"namaKeluarga\":\"Keluarga Santoso\",\"anggota\":[{\"nama\":\"Budi Santoso\",\"nik\":\"3174010101900001\",\"jenisKelamin\":\"pria\",\"hubunganKeluarga\":\"kepala-keluarga\"}]}}\n```",
	}
	feat.SetChatModelFactory(func(_ context.Context, _, _, _, _ string) (model.ToolCallingChatModel, error) {
		return fake, nil
	})

	ext, err := feat.ExtractIdentityDocument(ctx, user.IdentityExtractInput{
		KTP: &user.IdentityImageInput{MimeType: "image/jpeg", Data: []byte{0xFF, 0xD8, 0xFF, 0xE0}},
		KK:  &user.IdentityImageInput{MimeType: "image/png", Data: []byte{0x89, 'P', 'N', 'G'}},
	})
	require.NoError(t, err)
	require.NotNil(t, ext.KTP)
	require.Equal(t, "3174010101900001", ext.KTP.NIK)
	require.Equal(t, "Budi Santoso", ext.KTP.Nama)
	require.NotNil(t, ext.KK)
	require.Equal(t, "3174010101900099", ext.KK.NomorKK)
	require.Len(t, ext.KK.Anggota, 1)

	require.Len(t, fake.lastMessages, 2)
	userMsg := fake.lastMessages[1]
	require.NotEmpty(t, userMsg.UserInputMultiContent)
	hasImage := false
	for _, part := range userMsg.UserInputMultiContent {
		if part.Type == schema.ChatMessagePartTypeImageURL && part.Image != nil && part.Image.Base64Data != nil {
			hasImage = true
		}
	}
	require.True(t, hasImage, "expected multimodal image part")

	// 2. Unparseable output -> 422 VERIFY_EXTRACTION_FAILED
	fake.respContent = "not valid json at all"
	_, err = feat.ExtractIdentityDocument(ctx, user.IdentityExtractInput{
		KTP: &user.IdentityImageInput{MimeType: "image/jpeg", Data: []byte{0xFF, 0xD8, 0xFF, 0xE0}},
	})
	require.ErrorIs(t, err, domain.ErrVerifyExtractionFailed)

	// 3. Provider rejects images -> 400 AGENT_PROVIDER_REQUEST_INVALID
	fake.err = errors.New("400 bad request: model does not support image input")
	_, err = feat.ExtractIdentityDocument(ctx, user.IdentityExtractInput{
		KTP: &user.IdentityImageInput{MimeType: "image/jpeg", Data: []byte{0xFF, 0xD8, 0xFF, 0xE0}},
	})
	require.ErrorIs(t, err, domain.ErrProviderRequestInvalid)
}
