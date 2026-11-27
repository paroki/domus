package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/usernotificationlog"
	"github.com/paroki/domus/ent/userpushsubscription"
)

// UpsertPushSubscription creates or updates a Web Push subscription keyed by endpoint.
func (r *Repository) UpsertPushSubscription(
	ctx context.Context,
	userID uuid.UUID,
	sessionID *uuid.UUID,
	endpoint, p256dh, auth, userAgent string,
) (*ent.UserPushSubscription, error) {
	existing, err := r.client.UserPushSubscription.Query().
		Where(userpushsubscription.EndpointEQ(endpoint)).
		Only(ctx)
	if err == nil {
		upd := r.client.UserPushSubscription.UpdateOneID(existing.ID).
			SetUserID(userID).
			SetP256dh(p256dh).
			SetAuth(auth).
			SetUserAgent(userAgent)
		if sessionID != nil && *sessionID != uuid.Nil {
			upd.SetSessionID(*sessionID)
		} else {
			upd.ClearSessionID()
		}
		return upd.Save(ctx)
	}
	if !ent.IsNotFound(err) {
		return nil, err
	}

	cre := r.client.UserPushSubscription.Create().
		SetUserID(userID).
		SetEndpoint(endpoint).
		SetP256dh(p256dh).
		SetAuth(auth).
		SetUserAgent(userAgent)
	if sessionID != nil && *sessionID != uuid.Nil {
		cre.SetSessionID(*sessionID)
	}
	return cre.Save(ctx)
}

// DeletePushSubscriptionByEndpoint removes a subscription owned by userID matching endpoint.
func (r *Repository) DeletePushSubscriptionByEndpoint(ctx context.Context, userID uuid.UUID, endpoint string) (int, error) {
	return r.client.UserPushSubscription.Delete().
		Where(
			userpushsubscription.UserIDEQ(userID),
			userpushsubscription.EndpointEQ(endpoint),
		).
		Exec(ctx)
}

// DeletePushSubscriptionByID hard-deletes a subscription by primary key.
func (r *Repository) DeletePushSubscriptionByID(ctx context.Context, id uuid.UUID) error {
	return r.client.UserPushSubscription.DeleteOneID(id).Exec(ctx)
}

// DeletePushSubscriptionsBySessionID hard-deletes subscriptions bound to sessionID.
func (r *Repository) DeletePushSubscriptionsBySessionID(ctx context.Context, sessionID uuid.UUID) error {
	_, err := r.client.UserPushSubscription.Delete().
		Where(userpushsubscription.SessionIDEQ(sessionID)).
		Exec(ctx)
	return err
}

// DeletePushSubscriptionsByUserID hard-deletes all subscriptions owned by userID.
func (r *Repository) DeletePushSubscriptionsByUserID(ctx context.Context, userID uuid.UUID) error {
	_, err := r.client.UserPushSubscription.Delete().
		Where(userpushsubscription.UserIDEQ(userID)).
		Exec(ctx)
	return err
}

// ListPushSubscriptionsByUserID returns all Web Push subscriptions for userID.
func (r *Repository) ListPushSubscriptionsByUserID(ctx context.Context, userID uuid.UUID) ([]*ent.UserPushSubscription, error) {
	return r.client.UserPushSubscription.Query().
		Where(userpushsubscription.UserIDEQ(userID)).
		Order(ent.Desc(userpushsubscription.FieldCreatedAt)).
		All(ctx)
}

// CountPushSubscriptionsByUserID counts active Web Push subscriptions for userID.
func (r *Repository) CountPushSubscriptionsByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	return r.client.UserPushSubscription.Query().
		Where(userpushsubscription.UserIDEQ(userID)).
		Count(ctx)
}

// TouchPushSubscriptionLastUsed records the latest successful delivery timestamp.
func (r *Repository) TouchPushSubscriptionLastUsed(ctx context.Context, id uuid.UUID, at time.Time) error {
	return r.client.UserPushSubscription.UpdateOneID(id).
		SetLastUsedAt(at).
		Exec(ctx)
}

// TryRecordNotificationLog inserts a deduplication log row; returns (false, nil) if already recorded.
func (r *Repository) TryRecordNotificationLog(
	ctx context.Context,
	userID uuid.UUID,
	parokiID *uuid.UUID,
	kind string,
	refID uuid.UUID,
	window string,
	sentAt time.Time,
) (bool, error) {
	cre := r.client.UserNotificationLog.Create().
		SetUserID(userID).
		SetKind(kind).
		SetRefID(refID).
		SetWindow(window).
		SetSentAt(sentAt)
	if parokiID != nil && *parokiID != uuid.Nil {
		cre.SetParokiID(*parokiID)
	}
	if _, err := cre.Save(ctx); err != nil {
		if ent.IsConstraintError(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// DeleteNotificationLogsBefore prunes ephemeral notification logs older than before.
func (r *Repository) DeleteNotificationLogsBefore(ctx context.Context, before time.Time) (int, error) {
	return r.client.UserNotificationLog.Delete().
		Where(usernotificationlog.SentAtLT(before)).
		Exec(ctx)
}
