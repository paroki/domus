package service

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/internal/event/domain"
	"github.com/paroki/domus/internal/event/repo"
	"github.com/paroki/domus/internal/platform/authx"
)

// CreateEventLagu adds a liturgical song entry to eventID (requires event manager).
func (s *Service) CreateEventLagu(
	ctx context.Context,
	p authx.Principal,
	eventID uuid.UUID,
	req domain.CreateEventLaguRequest,
) (domain.EventLaguResponse, error) {
	ev, err := s.repo.FindEventInParoki(ctx, p.ParokiID, eventID)
	if ent.IsNotFound(err) {
		return domain.EventLaguResponse{}, domain.ErrEventNotFound
	}
	if err != nil {
		return domain.EventLaguResponse{}, err
	}

	canManage, err := s.canManageEvent(ctx, p, ev)
	if err != nil {
		return domain.EventLaguResponse{}, err
	}
	if !canManage {
		return domain.EventLaguResponse{}, domain.ErrForbidden
	}

	bagian := strings.TrimSpace(req.Bagian)
	if bagian == "" {
		return domain.EventLaguResponse{}, domain.ValidationError("bagian", "REQUIRED", "Song part (bagian) is required.")
	}
	judul := strings.TrimSpace(req.Judul)
	if judul == "" {
		return domain.EventLaguResponse{}, domain.ValidationError("judul", "REQUIRED", "Song title (judul) is required.")
	}

	urutan := 0
	if req.Urutan != nil {
		urutan = *req.Urutan
	}

	created, err := s.repo.CreateLagu(ctx, repo.CreateEventLaguInput{
		ParokiID:  p.ParokiID,
		EventID:   eventID,
		Urutan:    urutan,
		Bagian:    bagian,
		Sumber:    trimOptString(req.Sumber),
		Nomor:     trimOptString(req.Nomor),
		Judul:     judul,
		NadaDasar: trimOptString(req.NadaDasar),
		Note:      trimOptString(req.Note),
	})
	if err != nil {
		return domain.EventLaguResponse{}, err
	}
	return toLaguResponse(created), nil
}

// UpdateEventLagu updates an existing song entry in eventID (requires event manager).
func (s *Service) UpdateEventLagu(
	ctx context.Context,
	p authx.Principal,
	eventID, laguID uuid.UUID,
	req domain.UpdateEventLaguRequest,
) (domain.EventLaguResponse, error) {
	ev, err := s.repo.FindEventInParoki(ctx, p.ParokiID, eventID)
	if ent.IsNotFound(err) {
		return domain.EventLaguResponse{}, domain.ErrEventNotFound
	}
	if err != nil {
		return domain.EventLaguResponse{}, err
	}

	canManage, err := s.canManageEvent(ctx, p, ev)
	if err != nil {
		return domain.EventLaguResponse{}, err
	}
	if !canManage {
		return domain.EventLaguResponse{}, domain.ErrForbidden
	}

	if _, err := s.repo.FindLaguInEvent(ctx, p.ParokiID, eventID, laguID); err != nil {
		if ent.IsNotFound(err) {
			return domain.EventLaguResponse{}, domain.ErrLaguNotFound
		}
		return domain.EventLaguResponse{}, err
	}

	in := repo.UpdateEventLaguInput{
		Urutan: req.Urutan,
	}
	if req.Bagian != nil {
		b := strings.TrimSpace(*req.Bagian)
		if b == "" {
			return domain.EventLaguResponse{}, domain.ValidationError("bagian", "REQUIRED", "Song part (bagian) cannot be empty.")
		}
		in.Bagian = &b
	}
	if req.Judul != nil {
		j := strings.TrimSpace(*req.Judul)
		if j == "" {
			return domain.EventLaguResponse{}, domain.ValidationError("judul", "REQUIRED", "Song title (judul) cannot be empty.")
		}
		in.Judul = &j
	}
	if req.Sumber != nil {
		applyOptClearString(req.Sumber, &in.Sumber, &in.ClearSumber)
	}
	if req.Nomor != nil {
		applyOptClearString(req.Nomor, &in.Nomor, &in.ClearNomor)
	}
	if req.NadaDasar != nil {
		applyOptClearString(req.NadaDasar, &in.NadaDasar, &in.ClearNadaDasar)
	}
	if req.Note != nil {
		applyOptClearString(req.Note, &in.Note, &in.ClearNote)
	}

	updated, err := s.repo.UpdateLagu(ctx, laguID, in)
	if err != nil {
		return domain.EventLaguResponse{}, err
	}
	return toLaguResponse(updated), nil
}

// DeleteEventLagu hard-deletes a song entry from eventID (requires event manager).
func (s *Service) DeleteEventLagu(ctx context.Context, p authx.Principal, eventID, laguID uuid.UUID) error {
	ev, err := s.repo.FindEventInParoki(ctx, p.ParokiID, eventID)
	if ent.IsNotFound(err) {
		return domain.ErrEventNotFound
	}
	if err != nil {
		return err
	}

	canManage, err := s.canManageEvent(ctx, p, ev)
	if err != nil {
		return err
	}
	if !canManage {
		return domain.ErrForbidden
	}

	if _, err := s.repo.FindLaguInEvent(ctx, p.ParokiID, eventID, laguID); err != nil {
		if ent.IsNotFound(err) {
			return domain.ErrLaguNotFound
		}
		return err
	}

	return s.repo.DeleteLagu(ctx, laguID)
}

func applyOptClearString(raw *string, set **string, clear *bool) {
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		*clear = true
	} else {
		*set = &trimmed
	}
}

func toLaguResponse(row *ent.EventLagu) domain.EventLaguResponse {
	return domain.EventLaguResponse{
		ID:        row.ID,
		ParokiID:  row.ParokiID,
		EventID:   row.EventID,
		Urutan:    row.Urutan,
		Bagian:    row.Bagian,
		Sumber:    row.Sumber,
		Nomor:     row.Nomor,
		Judul:     row.Judul,
		NadaDasar: row.NadaDasar,
		Note:      row.Note,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
