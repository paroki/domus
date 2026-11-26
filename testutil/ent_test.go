package testutil_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/event"
	"github.com/paroki/domus/ent/eventrsvp"
	"github.com/paroki/domus/ent/organization"
	"github.com/paroki/domus/ent/orginvite"
	"github.com/paroki/domus/ent/orgmember"
	"github.com/paroki/domus/ent/parokimember"
	"github.com/paroki/domus/ent/period"
	"github.com/paroki/domus/ent/schema"
	"github.com/paroki/domus/ent/setting"
	"github.com/paroki/domus/ent/user"
	"github.com/paroki/domus/ent/wilayah"
	"github.com/paroki/domus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEntClient_SettingCRUDAndMixins(t *testing.T) {
	ctx := context.Background()
	client := testutil.NewEntClient(t)

	// UUIDv7 IDMixin & TimeMixin
	first, err := client.Setting.Create().SetKey("site.name").SetValue("Domus").Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, uuid.Version(7), first.ID.Version())
	assert.False(t, first.CreatedAt.IsZero())
	assert.False(t, first.UpdatedAt.IsZero())

	time.Sleep(2 * time.Millisecond)

	second, err := client.Setting.Create().SetKey("site.lang").SetValue("id").Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, uuid.Version(7), second.ID.Version())
	assert.Less(t, first.ID.String(), second.ID.String())

	// Unique constraint on key
	_, err = client.Setting.Create().SetKey("site.name").SetValue("Duplicate").Save(ctx)
	require.Error(t, err)
	assert.True(t, ent.IsConstraintError(err))

	// Read & Update
	got, err := client.Setting.Query().Where(setting.KeyEQ("site.name")).Only(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Domus", got.Value)

	time.Sleep(2 * time.Millisecond)
	updated, err := got.Update().SetValue("Paroki").Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Paroki", updated.Value)
	assert.True(t, updated.CreatedAt.Truncate(time.Microsecond).Equal(first.CreatedAt.Truncate(time.Microsecond)))
	assert.True(t, updated.UpdatedAt.After(first.UpdatedAt) || updated.UpdatedAt.Equal(first.UpdatedAt))

	// Delete
	require.NoError(t, client.Setting.DeleteOneID(updated.ID).Exec(ctx))
	exists, err := client.Setting.Query().Where(setting.ID(updated.ID)).Exist(ctx)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestMixins_SoftDeleteAndNewID(t *testing.T) {
	id := schema.NewID()
	assert.Equal(t, uuid.Version(7), id.Version())
	assert.Len(t, schema.SoftDeleteMixin{}.Fields(), 1)
}

func TestUserSchema_SoftDeleteAndUniqueEmail(t *testing.T) {
	ctx := context.Background()
	client := testutil.NewEntClient(t)

	u, err := client.User.Create().
		SetEmail("alice@paroki.org").
		SetName("Alice").
		Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, uuid.Version(7), u.ID.Version())
	assert.Equal(t, user.StatusActive, u.Status)
	assert.Nil(t, u.DeletedAt)

	_, err = client.User.Create().
		SetEmail("alice@paroki.org").
		SetName("Alice Clone").
		Save(ctx)
	require.Error(t, err)
	assert.True(t, ent.IsConstraintError(err))

	deletedAt := time.Now().UTC().Truncate(time.Second)
	u, err = u.Update().SetDeletedAt(deletedAt).Save(ctx)
	require.NoError(t, err)
	require.NotNil(t, u.DeletedAt)
	assert.True(t, u.DeletedAt.Equal(deletedAt))
}

func TestUserAccountSchema_UniqueProviderSubject(t *testing.T) {
	ctx := context.Background()
	client := testutil.NewEntClient(t)

	u, err := client.User.Create().
		SetEmail("bob@paroki.org").
		SetName("Bob").
		Save(ctx)
	require.NoError(t, err)

	providers := []string{"google", "github", "email", "telegram", "whatsapp"}
	for _, provider := range providers {
		acc, err := client.UserAccount.Create().
			SetUserID(u.ID).
			SetProvider(provider).
			SetSubject("shared-subject-123").
			Save(ctx)
		require.NoError(t, err, "provider %s should allow same subject", provider)
		assert.Equal(t, u.ID, acc.UserID)
	}

	_, err = client.UserAccount.Create().
		SetUserID(u.ID).
		SetProvider("google").
		SetSubject("shared-subject-123").
		Save(ctx)
	require.Error(t, err)
	assert.True(t, ent.IsConstraintError(err))
}

func TestUserSessionSchema_MultiDeviceAndUniqueTokenHash(t *testing.T) {
	ctx := context.Background()
	client := testutil.NewEntClient(t)

	u, err := client.User.Create().
		SetEmail("carol@paroki.org").
		SetName("Carol").
		Save(ctx)
	require.NoError(t, err)

	expiresAt := time.Now().UTC().Add(24 * time.Hour)
	s1, err := client.UserSession.Create().
		SetUserID(u.ID).
		SetTokenHash("hash-device-1").
		SetUserAgent("Firefox/130").
		SetIPAddress("127.0.0.1").
		SetExpiresAt(expiresAt).
		Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, u.ID, s1.UserID)

	s2, err := client.UserSession.Create().
		SetUserID(u.ID).
		SetTokenHash("hash-device-2").
		SetUserAgent("Mobile/1.0").
		SetIPAddress("10.0.0.2").
		SetExpiresAt(expiresAt).
		Save(ctx)
	require.NoError(t, err)
	assert.NotEqual(t, s1.ID, s2.ID)

	sessions, err := u.QuerySessions().All(ctx)
	require.NoError(t, err)
	assert.Len(t, sessions, 2)

	_, err = client.UserSession.Create().
		SetUserID(u.ID).
		SetTokenHash("hash-device-1").
		SetExpiresAt(expiresAt).
		Save(ctx)
	require.Error(t, err)
	assert.True(t, ent.IsConstraintError(err))
}

func TestUserTokenSchema_UniqueTokenHashAndOptionalExpiry(t *testing.T) {
	ctx := context.Background()
	client := testutil.NewEntClient(t)

	u, err := client.User.Create().
		SetEmail("pat@paroki.org").
		SetName("PAT User").
		Save(ctx)
	require.NoError(t, err)

	// Token without expiry (expires_at = nil)
	t1, err := client.UserToken.Create().
		SetUserID(u.ID).
		SetName("CLI Exporter").
		SetTokenPrefix("dms_abcdefgh").
		SetTokenHash("sha256-pat-hash-1").
		Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, u.ID, t1.UserID)
	assert.Nil(t, t1.ExpiresAt)
	assert.Nil(t, t1.LastUsedAt)

	// Token with expiry
	expiresAt := time.Now().UTC().Add(30 * 24 * time.Hour)
	t2, err := client.UserToken.Create().
		SetUserID(u.ID).
		SetName("MCP Client").
		SetTokenPrefix("dms_12345678").
		SetTokenHash("sha256-pat-hash-2").
		SetExpiresAt(expiresAt).
		Save(ctx)
	require.NoError(t, err)
	assert.NotNil(t, t2.ExpiresAt)

	tokens, err := u.QueryTokens().All(ctx)
	require.NoError(t, err)
	assert.Len(t, tokens, 2)

	// Duplicate token_hash must fail with constraint error
	_, err = client.UserToken.Create().
		SetUserID(u.ID).
		SetName("Duplicate").
		SetTokenPrefix("dms_abcdefgh").
		SetTokenHash("sha256-pat-hash-1").
		Save(ctx)
	require.Error(t, err)
	assert.True(t, ent.IsConstraintError(err))
}

func TestUserOTPSchema_Persistence(t *testing.T) {
	ctx := context.Background()
	client := testutil.NewEntClient(t)

	expiresAt := time.Now().UTC().Add(10 * time.Minute)
	otp, err := client.UserOTP.Create().
		SetTarget("dave@paroki.org").
		SetCodeHash("sha256-otp-hash").
		SetExpiresAt(expiresAt).
		Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, uuid.Version(7), otp.ID.Version())
	assert.Equal(t, "email", otp.Channel)
	assert.Equal(t, 0, otp.Attempts)
	assert.Equal(t, "dave@paroki.org", otp.Target)
	assert.Equal(t, "sha256-otp-hash", otp.CodeHash)
}

func TestRefSchemas_TerritoryHierarchyAndUniqueness(t *testing.T) {
	ctx := context.Background()
	client := testutil.NewEntClient(t)

	keuskupan, err := client.Keuskupan.Create().
		SetKode("KAJ").
		SetNama("Keuskupan Agung Jakarta").
		Save(ctx)
	require.NoError(t, err)
	assert.True(t, keuskupan.Active)
	assert.Nil(t, keuskupan.DeletedAt)

	p1, err := client.Paroki.Create().
		SetKeuskupan(keuskupan).
		SetKode("P001").
		SetNama("Paroki Katedral").
		Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, keuskupan.ID, p1.KeuskupanID)

	p2, err := client.Paroki.Create().
		SetKeuskupan(keuskupan).
		SetKode("P002").
		SetNama("Paroki Blok B").
		Save(ctx)
	require.NoError(t, err)

	w1, err := client.Wilayah.Create().
		SetParoki(p1).
		SetKode("WIL-01").
		SetNama("Wilayah 1").
		Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, wilayah.JenisWilayah, w1.Jenis)

	// Same (paroki_id, kode) on Wilayah fails; different paroki succeeds.
	_, err = client.Wilayah.Create().
		SetParoki(p1).
		SetKode("WIL-01").
		SetNama("Wilayah Duplikat").
		Save(ctx)
	require.Error(t, err)
	assert.True(t, ent.IsConstraintError(err))

	w2, err := client.Wilayah.Create().
		SetParoki(p2).
		SetJenis(wilayah.JenisStasi).
		SetKode("WIL-01").
		SetNama("Stasi 1").
		Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, wilayah.JenisStasi, w2.Jenis)

	l1, err := client.Lingkungan.Create().
		SetParoki(p1).
		SetWilayah(w1).
		SetKode("LING-01").
		SetNama("Lingkungan 1").
		Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, p1.ID, l1.ParokiID)
	assert.Equal(t, w1.ID, l1.WilayahID)

	// Same (paroki_id, kode) on Lingkungan fails.
	_, err = client.Lingkungan.Create().
		SetParoki(p1).
		SetWilayah(w1).
		SetKode("LING-01").
		SetNama("Lingkungan Duplikat").
		Save(ctx)
	require.Error(t, err)
	assert.True(t, ent.IsConstraintError(err))

	// Traverse Keuskupan -> Paroki -> Wilayah -> Lingkungan.
	parokis, err := keuskupan.QueryParoki().All(ctx)
	require.NoError(t, err)
	assert.Len(t, parokis, 2)

	lingkungans, err := w1.QueryLingkungan().All(ctx)
	require.NoError(t, err)
	assert.Len(t, lingkungans, 1)
	assert.Equal(t, l1.ID, lingkungans[0].ID)
}

func TestParokiMemberSchema_UniqueParokiUserAndNoSoftDelete(t *testing.T) {
	ctx := context.Background()
	client := testutil.NewEntClient(t)

	// ParokiMember must NOT embed SoftDeleteMixin.
	assert.Len(t, schema.ParokiMember{}.Mixin(), 2)

	keuskupan, err := client.Keuskupan.Create().SetKode("KAS").SetNama("Keuskupan Agung Semarang").Save(ctx)
	require.NoError(t, err)
	paroki, err := client.Paroki.Create().SetKeuskupan(keuskupan).SetKode("P-KAS-01").SetNama("Paroki Kotabaru").Save(ctx)
	require.NoError(t, err)

	userID := schema.NewID()
	member, err := client.ParokiMember.Create().
		SetParoki(paroki).
		SetUserID(userID).
		SetRole(parokimember.RoleAdmin).
		Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, parokimember.RoleAdmin, member.Role)

	_, err = client.ParokiMember.Create().
		SetParoki(paroki).
		SetUserID(userID).
		Save(ctx)
	require.Error(t, err)
	assert.True(t, ent.IsConstraintError(err))
}

func TestLookupAndPeriodSchemas_GlobalVsTenantAndPersistence(t *testing.T) {
	ctx := context.Background()
	client := testutil.NewEntClient(t)

	keuskupan, err := client.Keuskupan.Create().SetKode("KBD").SetNama("Keuskupan Bandung").Save(ctx)
	require.NoError(t, err)
	paroki, err := client.Paroki.Create().SetKeuskupan(keuskupan).SetKode("P-KBD-01").SetNama("Paroki Katedral Bandung").Save(ctx)
	require.NoError(t, err)

	// Global lookup (paroki_id = nil)
	globalLookup, err := client.Lookup.Create().
		SetCategory("sacrament").
		SetCode("BAPTISM").
		SetLabel("Baptis").
		SetSortOrder(1).
		Save(ctx)
	require.NoError(t, err)
	assert.Nil(t, globalLookup.ParokiID)

	// Tenant lookup (paroki_id = paroki.ID)
	tenantLookup, err := client.Lookup.Create().
		SetParoki(paroki).
		SetCategory("sacrament").
		SetCode("BAPTISM").
		SetLabel("Baptis Paroki").
		SetSortOrder(1).
		Save(ctx)
	require.NoError(t, err)
	require.NotNil(t, tenantLookup.ParokiID)
	assert.Equal(t, paroki.ID, *tenantLookup.ParokiID)

	// Duplicate tenant lookup fails
	_, err = client.Lookup.Create().
		SetParoki(paroki).
		SetCategory("sacrament").
		SetCode("BAPTISM").
		SetLabel("Duplicate").
		Save(ctx)
	require.Error(t, err)
	assert.True(t, ent.IsConstraintError(err))

	// Period persistence
	p, err := client.Period.Create().
		SetParoki(paroki).
		SetType(period.TypeLiturgical).
		SetCode("LIT-2026").
		SetName("Tahun Liturgi 2026").
		SetStartDate("2025-11-30").
		SetEndDate("2026-11-28").
		SetIsCurrent(true).
		Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, period.TypeLiturgical, p.Type)
	assert.Equal(t, period.StatusOpen, p.Status)
	assert.True(t, p.IsCurrent)
}

func TestOrgSchemas_HierarchyUniquenessAndMixins(t *testing.T) {
	ctx := context.Background()
	client := testutil.NewEntClient(t)

	// Mixin checks: Organization has SoftDeleteMixin (3); OrgMember and OrgInvite do not (2).
	assert.Len(t, schema.Organization{}.Mixin(), 3)
	assert.Len(t, schema.OrgMember{}.Mixin(), 2)
	assert.Len(t, schema.OrgInvite{}.Mixin(), 2)

	parokiID := schema.NewID()
	otherParokiID := schema.NewID()
	userID := schema.NewID()

	// Organization parent/child self-edge.
	parent, err := client.Organization.Create().
		SetParokiID(parokiID).
		SetKind(organization.KindDpp).
		SetCode("DPP-01").
		SetName("Dewan Pastoral Paroki").
		Save(ctx)
	require.NoError(t, err)
	assert.True(t, parent.Active)

	child, err := client.Organization.Create().
		SetParokiID(parokiID).
		SetParent(parent).
		SetKind(organization.KindSection).
		SetCode("SEK-LIT").
		SetName("Seksi Liturgi").
		Save(ctx)
	require.NoError(t, err)
	require.NotNil(t, child.ParentID)
	assert.Equal(t, parent.ID, *child.ParentID)

	children, err := parent.QueryChildren().All(ctx)
	require.NoError(t, err)
	require.Len(t, children, 1)
	assert.Equal(t, child.ID, children[0].ID)

	// (paroki_id, code) uniqueness on Organization: same paroki fails, different paroki succeeds.
	_, err = client.Organization.Create().
		SetParokiID(parokiID).
		SetCode("DPP-01").
		SetName("Duplicate DPP").
		Save(ctx)
	require.Error(t, err)
	assert.True(t, ent.IsConstraintError(err))

	_, err = client.Organization.Create().
		SetParokiID(otherParokiID).
		SetCode("DPP-01").
		SetName("Other Paroki DPP").
		Save(ctx)
	require.NoError(t, err)

	// OrgMember (org_id, user_id) uniqueness.
	member, err := client.OrgMember.Create().
		SetParokiID(parokiID).
		SetOrganization(parent).
		SetUserID(userID).
		SetRole(orgmember.RoleAdmin).
		Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, orgmember.RoleAdmin, member.Role)
	assert.Equal(t, orgmember.StatusActive, member.Status)

	_, err = client.OrgMember.Create().
		SetParokiID(parokiID).
		SetOrganization(parent).
		SetUserID(userID).
		Save(ctx)
	require.Error(t, err)
	assert.True(t, ent.IsConstraintError(err))

	// OrgInvite code uniqueness.
	invite, err := client.OrgInvite.Create().
		SetParokiID(parokiID).
		SetOrganization(parent).
		SetCode("INV-ORG-01").
		SetExpiresAt(time.Now().Add(24 * time.Hour)).
		SetCreatedBy(userID).
		Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, invite.UsedCount)
	assert.Equal(t, orginvite.RoleMember, invite.Role)

	_, err = client.OrgInvite.Create().
		SetParokiID(parokiID).
		SetOrganization(child).
		SetCode("INV-ORG-01").
		SetExpiresAt(time.Now().Add(24 * time.Hour)).
		SetCreatedBy(userID).
		Save(ctx)
	require.Error(t, err)
	assert.True(t, ent.IsConstraintError(err))
}

func TestNewEntClient_EventAndRSVPSchemas(t *testing.T) {
	ctx := context.Background()
	client := testutil.NewEntClient(t)

	parokiID := schema.NewID()
	orgID := schema.NewID()
	creatorID := schema.NewID()
	userID := schema.NewID()
	startAt := time.Now().UTC().Add(24 * time.Hour)
	endAt := startAt.Add(2 * time.Hour)

	ev, err := client.Event.Create().
		SetParokiID(parokiID).
		SetOrgID(orgID).
		SetCreatedBy(creatorID).
		SetTitle("Rapat Pleno DPP").
		SetStartAt(startAt).
		SetEndAt(endAt).
		Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, 7, int(ev.ID.Version()))
	assert.False(t, ev.CreatedAt.IsZero())
	assert.False(t, ev.UpdatedAt.IsZero())
	assert.Nil(t, ev.DeletedAt)
	assert.Equal(t, event.VisibilityPublic, ev.Visibility)
	assert.True(t, ev.RsvpEnabled)
	assert.Equal(t, event.StatusScheduled, ev.Status)

	rsvp, err := client.EventRSVP.Create().
		SetParokiID(parokiID).
		SetEvent(ev).
		SetUserID(userID).
		SetStatus(eventrsvp.StatusGoing).
		Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, 7, int(rsvp.ID.Version()))
	assert.False(t, rsvp.CreatedAt.IsZero())
	assert.False(t, rsvp.UpdatedAt.IsZero())
	assert.Equal(t, ev.ID, rsvp.EventID)

	// Query Event -> EventRSVP edge.
	rsvps, err := ev.QueryRsvps().All(ctx)
	require.NoError(t, err)
	require.Len(t, rsvps, 1)
	assert.Equal(t, rsvp.ID, rsvps[0].ID)

	// (event_id, user_id) uniqueness.
	_, err = client.EventRSVP.Create().
		SetParokiID(parokiID).
		SetEvent(ev).
		SetUserID(userID).
		SetStatus(eventrsvp.StatusMaybe).
		Save(ctx)
	require.Error(t, err)
	assert.True(t, ent.IsConstraintError(err))

	// Parish-level event with org_id = nil and jenis.
	parishEv, err := client.Event.Create().
		SetParokiID(parokiID).
		SetCreatedBy(creatorID).
		SetTitle("Misa Syukur Paroki").
		SetJenis("misa").
		SetStartAt(startAt).
		SetEndAt(endAt).
		Save(ctx)
	require.NoError(t, err)
	assert.Nil(t, parishEv.OrgID)
	require.NotNil(t, parishEv.Jenis)
	assert.Equal(t, "misa", *parishEv.Jenis)
}
