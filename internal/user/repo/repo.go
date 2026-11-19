package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/user"
	"github.com/paroki/domus/ent/useraccount"
	"github.com/paroki/domus/ent/userotp"
	"github.com/paroki/domus/ent/usersession"
)

// Repository provides Ent-backed persistence for User, UserAccount, UserOTP, and UserSession.
type Repository struct {
	client *ent.Client
}

// New creates a Repository backed by the given Ent client.
func New(client *ent.Client) *Repository {
	return &Repository{client: client}
}

// WithTx executes fn inside a database transaction.
func (r *Repository) WithTx(ctx context.Context, fn func(txRepo *Repository) error) error {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return err
	}
	txRepo := &Repository{client: tx.Client()}
	if err := fn(txRepo); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			return fmt.Errorf("%w (rollback failed: %v)", err, rerr)
		}
		return err
	}
	return tx.Commit()
}

// FindUserByEmail returns the active (non-soft-deleted) user matching email.
func (r *Repository) FindUserByEmail(ctx context.Context, email string) (*ent.User, error) {
	return r.client.User.Query().
		Where(user.EmailEQ(email), user.DeletedAtIsNil()).
		Only(ctx)
}

// FindUserByID returns the non-soft-deleted user matching id.
func (r *Repository) FindUserByID(ctx context.Context, id uuid.UUID) (*ent.User, error) {
	return r.client.User.Query().
		Where(user.IDEQ(id), user.DeletedAtIsNil()).
		Only(ctx)
}

// CreateUser inserts a new User record.
func (r *Repository) CreateUser(ctx context.Context, email, name string, phone *string) (*ent.User, error) {
	q := r.client.User.Create().
		SetEmail(email).
		SetName(name)
	if phone != nil && *phone != "" {
		q.SetPhone(*phone)
	}
	return q.Save(ctx)
}

// UpdateUserLastLogin updates last_login_at for the given user ID.
func (r *Repository) UpdateUserLastLogin(ctx context.Context, id uuid.UUID, at time.Time) (*ent.User, error) {
	return r.client.User.UpdateOneID(id).
		SetLastLoginAt(at).
		Save(ctx)
}

// FindAccountByProviderSubject returns the UserAccount matching (provider, subject) if present.
func (r *Repository) FindAccountByProviderSubject(ctx context.Context, provider, subject string) (*ent.UserAccount, error) {
	return r.client.UserAccount.Query().
		Where(useraccount.ProviderEQ(provider), useraccount.SubjectEQ(subject)).
		Only(ctx)
}

// UpsertAccount creates or updates a UserAccount keyed by (provider, subject).
func (r *Repository) UpsertAccount(
	ctx context.Context,
	userID uuid.UUID,
	provider, subject string,
	email, displayName, avatarURL *string,
	lastUsedAt time.Time,
) (*ent.UserAccount, error) {
	existing, err := r.FindAccountByProviderSubject(ctx, provider, subject)
	if err == nil {
		upd := existing.Update().
			SetUserID(userID).
			SetLastUsedAt(lastUsedAt)
		if email != nil {
			upd.SetEmail(*email)
		}
		if displayName != nil {
			upd.SetDisplayName(*displayName)
		}
		if avatarURL != nil {
			upd.SetAvatarURL(*avatarURL)
		}
		return upd.Save(ctx)
	}
	if !ent.IsNotFound(err) {
		return nil, err
	}

	cre := r.client.UserAccount.Create().
		SetUserID(userID).
		SetProvider(provider).
		SetSubject(subject).
		SetLastUsedAt(lastUsedAt)
	if email != nil {
		cre.SetEmail(*email)
	}
	if displayName != nil {
		cre.SetDisplayName(*displayName)
	}
	if avatarURL != nil {
		cre.SetAvatarURL(*avatarURL)
	}
	return cre.Save(ctx)
}

// CreateOTP inserts a new UserOTP row.
func (r *Repository) CreateOTP(ctx context.Context, channel, target, codeHash string, expiresAt time.Time) (*ent.UserOTP, error) {
	return r.client.UserOTP.Create().
		SetChannel(channel).
		SetTarget(target).
		SetCodeHash(codeHash).
		SetExpiresAt(expiresAt).
		Save(ctx)
}

// FindLatestOTP returns the most recently created UserOTP for (channel, target).
func (r *Repository) FindLatestOTP(ctx context.Context, channel, target string) (*ent.UserOTP, error) {
	return r.client.UserOTP.Query().
		Where(userotp.ChannelEQ(channel), userotp.TargetEQ(target)).
		Order(ent.Desc(userotp.FieldCreatedAt)).
		First(ctx)
}

// IncrementOTPAttempts increments the attempts counter on a UserOTP row.
func (r *Repository) IncrementOTPAttempts(ctx context.Context, id uuid.UUID) (*ent.UserOTP, error) {
	return r.client.UserOTP.UpdateOneID(id).
		AddAttempts(1).
		Save(ctx)
}

// DeleteOTP hard-deletes a UserOTP row by ID.
func (r *Repository) DeleteOTP(ctx context.Context, id uuid.UUID) error {
	return r.client.UserOTP.DeleteOneID(id).Exec(ctx)
}

// DeleteOTPsByTarget hard-deletes all UserOTP rows for (channel, target).
func (r *Repository) DeleteOTPsByTarget(ctx context.Context, channel, target string) error {
	_, err := r.client.UserOTP.Delete().
		Where(userotp.ChannelEQ(channel), userotp.TargetEQ(target)).
		Exec(ctx)
	return err
}

// CreateSession inserts a new UserSession row.
func (r *Repository) CreateSession(
	ctx context.Context,
	userID uuid.UUID,
	tokenHash, userAgent, ipAddress string,
	activeParishID *uuid.UUID,
	expiresAt time.Time,
) (*ent.UserSession, error) {
	cre := r.client.UserSession.Create().
		SetUserID(userID).
		SetTokenHash(tokenHash).
		SetUserAgent(userAgent).
		SetIPAddress(ipAddress).
		SetExpiresAt(expiresAt)
	if activeParishID != nil {
		cre.SetActiveParishID(*activeParishID)
	}
	return cre.Save(ctx)
}

// FindSessionByTokenHash returns the UserSession matching tokenHash.
func (r *Repository) FindSessionByTokenHash(ctx context.Context, tokenHash string) (*ent.UserSession, error) {
	return r.client.UserSession.Query().
		Where(usersession.TokenHashEQ(tokenHash)).
		Only(ctx)
}

// DeleteSessionByTokenHash hard-deletes any UserSession matching tokenHash.
func (r *Repository) DeleteSessionByTokenHash(ctx context.Context, tokenHash string) error {
	_, err := r.client.UserSession.Delete().
		Where(usersession.TokenHashEQ(tokenHash)).
		Exec(ctx)
	return err
}

// FindSessionWithUserByTokenHash returns the UserSession and its non-deleted User.
func (r *Repository) FindSessionWithUserByTokenHash(ctx context.Context, tokenHash string) (*ent.UserSession, *ent.User, error) {
	sess, err := r.client.UserSession.Query().
		Where(usersession.TokenHashEQ(tokenHash)).
		WithUser(func(q *ent.UserQuery) {
			q.Where(user.DeletedAtIsNil())
		}).
		Only(ctx)
	if err != nil {
		return nil, nil, err
	}
	if sess.Edges.User == nil {
		return nil, nil, &ent.NotFoundError{}
	}
	return sess, sess.Edges.User, nil
}

// RefreshSession updates last_active_at and sliding expires_at on a UserSession.
func (r *Repository) RefreshSession(ctx context.Context, id uuid.UUID, lastActiveAt, expiresAt time.Time) error {
	return r.client.UserSession.UpdateOneID(id).
		SetLastActiveAt(lastActiveAt).
		SetExpiresAt(expiresAt).
		Exec(ctx)
}

// DeleteSessionByID hard-deletes a UserSession by ID.
func (r *Repository) DeleteSessionByID(ctx context.Context, id uuid.UUID) error {
	return r.client.UserSession.DeleteOneID(id).Exec(ctx)
}

// DeleteSessionsByUserID hard-deletes all UserSession rows for userID.
func (r *Repository) DeleteSessionsByUserID(ctx context.Context, userID uuid.UUID) error {
	_, err := r.client.UserSession.Delete().
		Where(usersession.UserIDEQ(userID)).
		Exec(ctx)
	return err
}

// UpdateAccountLastUsed updates last_used_at on a UserAccount.
func (r *Repository) UpdateAccountLastUsed(ctx context.Context, id uuid.UUID, at time.Time) error {
	return r.client.UserAccount.UpdateOneID(id).
		SetLastUsedAt(at).
		Exec(ctx)
}

// UpdateAccountActiveParish sets active_parish_id on a UserAccount.
func (r *Repository) UpdateAccountActiveParish(ctx context.Context, id, parishID uuid.UUID) error {
	return r.client.UserAccount.UpdateOneID(id).
		SetActiveParishID(parishID).
		Exec(ctx)
}

// SetUserDefaultParishIfNil sets default_parish_id on User only if currently nil.
func (r *Repository) SetUserDefaultParishIfNil(ctx context.Context, userID, parishID uuid.UUID) error {
	_, err := r.client.User.Update().
		Where(user.IDEQ(userID), user.DefaultParishIDIsNil()).
		SetDefaultParishID(parishID).
		Save(ctx)
	return err
}
