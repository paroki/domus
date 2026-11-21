package testutil_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/event"
	"github.com/paroki/domus/ent/eventrsvp"
	"github.com/paroki/domus/ent/meta"
	"github.com/paroki/domus/ent/organization"
	"github.com/paroki/domus/ent/orginvite"
	"github.com/paroki/domus/ent/orgmember"
	"github.com/paroki/domus/ent/parishmember"
	"github.com/paroki/domus/ent/period"
	"github.com/paroki/domus/ent/schema"
	"github.com/paroki/domus/ent/user"
	"github.com/paroki/domus/ent/zone"
	"github.com/paroki/domus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEntClient_MetaCRUDAndMixins(t *testing.T) {
	ctx := context.Background()
	client := testutil.NewEntClient(t)

	// UUIDv7 IDMixin & TimeMixin
	first, err := client.Meta.Create().SetKey("site.name").SetValue("Domus").Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, uuid.Version(7), first.ID.Version())
	assert.False(t, first.CreatedAt.IsZero())
	assert.False(t, first.UpdatedAt.IsZero())

	time.Sleep(2 * time.Millisecond)

	second, err := client.Meta.Create().SetKey("site.lang").SetValue("id").Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, uuid.Version(7), second.ID.Version())
	assert.Less(t, first.ID.String(), second.ID.String())

	// Unique constraint on key
	_, err = client.Meta.Create().SetKey("site.name").SetValue("Duplicate").Save(ctx)
	require.Error(t, err)
	assert.True(t, ent.IsConstraintError(err))

	// Read & Update
	got, err := client.Meta.Query().Where(meta.KeyEQ("site.name")).Only(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Domus", got.Value)

	time.Sleep(2 * time.Millisecond)
	updated, err := got.Update().SetValue("Paroki").Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Paroki", updated.Value)
	assert.True(t, updated.CreatedAt.Truncate(time.Microsecond).Equal(first.CreatedAt.Truncate(time.Microsecond)))
	assert.True(t, updated.UpdatedAt.After(first.UpdatedAt) || updated.UpdatedAt.Equal(first.UpdatedAt))

	// Delete
	require.NoError(t, client.Meta.DeleteOneID(updated.ID).Exec(ctx))
	exists, err := client.Meta.Query().Where(meta.ID(updated.ID)).Exist(ctx)
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

	diocese, err := client.Diocese.Create().
		SetCode("KAJ").
		SetName("Keuskupan Agung Jakarta").
		Save(ctx)
	require.NoError(t, err)
	assert.True(t, diocese.Active)
	assert.Nil(t, diocese.DeletedAt)

	p1, err := client.Parish.Create().
		SetDiocese(diocese).
		SetCode("P001").
		SetName("Paroki Katedral").
		Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, diocese.ID, p1.DioceseID)

	p2, err := client.Parish.Create().
		SetDiocese(diocese).
		SetCode("P002").
		SetName("Paroki Blok B").
		Save(ctx)
	require.NoError(t, err)

	z1, err := client.Zone.Create().
		SetParish(p1).
		SetCode("WIL-01").
		SetName("Wilayah 1").
		Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, zone.TypeZone, z1.Type)

	// Same (parish_id, code) on Zone fails; different parish succeeds.
	_, err = client.Zone.Create().
		SetParish(p1).
		SetCode("WIL-01").
		SetName("Wilayah Duplikat").
		Save(ctx)
	require.Error(t, err)
	assert.True(t, ent.IsConstraintError(err))

	z2, err := client.Zone.Create().
		SetParish(p2).
		SetType(zone.TypeStation).
		SetCode("WIL-01").
		SetName("Stasi 1").
		Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, zone.TypeStation, z2.Type)

	w1, err := client.Ward.Create().
		SetParish(p1).
		SetZone(z1).
		SetCode("LING-01").
		SetName("Lingkungan 1").
		Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, p1.ID, w1.ParishID)
	assert.Equal(t, z1.ID, w1.ZoneID)

	// Same (parish_id, code) on Ward fails.
	_, err = client.Ward.Create().
		SetParish(p1).
		SetZone(z1).
		SetCode("LING-01").
		SetName("Lingkungan Duplikat").
		Save(ctx)
	require.Error(t, err)
	assert.True(t, ent.IsConstraintError(err))

	// Traverse Diocese -> Parishes -> Zones -> Wards.
	parishes, err := diocese.QueryParishes().All(ctx)
	require.NoError(t, err)
	assert.Len(t, parishes, 2)

	wards, err := z1.QueryWards().All(ctx)
	require.NoError(t, err)
	assert.Len(t, wards, 1)
	assert.Equal(t, w1.ID, wards[0].ID)
}

func TestParishMemberSchema_UniqueParishUserAndNoSoftDelete(t *testing.T) {
	ctx := context.Background()
	client := testutil.NewEntClient(t)

	// ParishMember must NOT embed SoftDeleteMixin.
	assert.Len(t, schema.ParishMember{}.Mixin(), 2)

	diocese, err := client.Diocese.Create().SetCode("KAS").SetName("Keuskupan Agung Semarang").Save(ctx)
	require.NoError(t, err)
	parish, err := client.Parish.Create().SetDiocese(diocese).SetCode("P-KAS-01").SetName("Paroki Kotabaru").Save(ctx)
	require.NoError(t, err)

	userID := schema.NewID()
	member, err := client.ParishMember.Create().
		SetParish(parish).
		SetUserID(userID).
		SetRole(parishmember.RoleAdmin).
		Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, parishmember.RoleAdmin, member.Role)

	_, err = client.ParishMember.Create().
		SetParish(parish).
		SetUserID(userID).
		Save(ctx)
	require.Error(t, err)
	assert.True(t, ent.IsConstraintError(err))
}

func TestLookupAndPeriodSchemas_GlobalVsTenantAndPersistence(t *testing.T) {
	ctx := context.Background()
	client := testutil.NewEntClient(t)

	diocese, err := client.Diocese.Create().SetCode("KBD").SetName("Keuskupan Bandung").Save(ctx)
	require.NoError(t, err)
	parish, err := client.Parish.Create().SetDiocese(diocese).SetCode("P-KBD-01").SetName("Paroki Katedral Bandung").Save(ctx)
	require.NoError(t, err)

	// Global lookup (parish_id = nil)
	globalLookup, err := client.Lookup.Create().
		SetCategory("sacrament").
		SetCode("BAPTISM").
		SetLabel("Baptis").
		SetSortOrder(1).
		Save(ctx)
	require.NoError(t, err)
	assert.Nil(t, globalLookup.ParishID)

	// Tenant lookup (parish_id = parish.ID)
	tenantLookup, err := client.Lookup.Create().
		SetParish(parish).
		SetCategory("sacrament").
		SetCode("BAPTISM").
		SetLabel("Baptis Paroki").
		SetSortOrder(1).
		Save(ctx)
	require.NoError(t, err)
	require.NotNil(t, tenantLookup.ParishID)
	assert.Equal(t, parish.ID, *tenantLookup.ParishID)

	// Duplicate tenant lookup fails
	_, err = client.Lookup.Create().
		SetParish(parish).
		SetCategory("sacrament").
		SetCode("BAPTISM").
		SetLabel("Duplicate").
		Save(ctx)
	require.Error(t, err)
	assert.True(t, ent.IsConstraintError(err))

	// Period persistence
	p, err := client.Period.Create().
		SetParish(parish).
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

	parishID := schema.NewID()
	otherParishID := schema.NewID()
	userID := schema.NewID()

	// Organization parent/child self-edge.
	parent, err := client.Organization.Create().
		SetParishID(parishID).
		SetKind(organization.KindDpp).
		SetCode("DPP-01").
		SetName("Dewan Pastoral Paroki").
		Save(ctx)
	require.NoError(t, err)
	assert.True(t, parent.Active)

	child, err := client.Organization.Create().
		SetParishID(parishID).
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

	// (parish_id, code) uniqueness on Organization: same parish fails, different parish succeeds.
	_, err = client.Organization.Create().
		SetParishID(parishID).
		SetCode("DPP-01").
		SetName("Duplicate DPP").
		Save(ctx)
	require.Error(t, err)
	assert.True(t, ent.IsConstraintError(err))

	_, err = client.Organization.Create().
		SetParishID(otherParishID).
		SetCode("DPP-01").
		SetName("Other Parish DPP").
		Save(ctx)
	require.NoError(t, err)

	// OrgMember (org_id, user_id) uniqueness.
	member, err := client.OrgMember.Create().
		SetParishID(parishID).
		SetOrganization(parent).
		SetUserID(userID).
		SetRole(orgmember.RoleAdmin).
		Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, orgmember.RoleAdmin, member.Role)
	assert.Equal(t, orgmember.StatusActive, member.Status)

	_, err = client.OrgMember.Create().
		SetParishID(parishID).
		SetOrganization(parent).
		SetUserID(userID).
		Save(ctx)
	require.Error(t, err)
	assert.True(t, ent.IsConstraintError(err))

	// OrgInvite code uniqueness.
	invite, err := client.OrgInvite.Create().
		SetParishID(parishID).
		SetOrganization(parent).
		SetCode("INV-ORG-01").
		SetExpiresAt(time.Now().Add(24 * time.Hour)).
		SetCreatedBy(userID).
		Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, invite.UsedCount)
	assert.Equal(t, orginvite.RoleMember, invite.Role)

	_, err = client.OrgInvite.Create().
		SetParishID(parishID).
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

	parishID := schema.NewID()
	orgID := schema.NewID()
	creatorID := schema.NewID()
	userID := schema.NewID()
	startAt := time.Now().UTC().Add(24 * time.Hour)
	endAt := startAt.Add(2 * time.Hour)

	ev, err := client.Event.Create().
		SetParishID(parishID).
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
		SetParishID(parishID).
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
		SetParishID(parishID).
		SetEvent(ev).
		SetUserID(userID).
		SetStatus(eventrsvp.StatusMaybe).
		Save(ctx)
	require.Error(t, err)
	assert.True(t, ent.IsConstraintError(err))
}
