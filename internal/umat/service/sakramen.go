package service

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/umatsakramen"
	"github.com/paroki/domus/internal/umat/domain"
)

func sakramenLifecycleRank(jenis string) int {
	switch jenis {
	case string(umatsakramen.JenisBaptis):
		return 1
	case string(umatsakramen.JenisKomuniPertama):
		return 2
	case string(umatsakramen.JenisKrisma):
		return 3
	case string(umatsakramen.JenisPerkawinan):
		return 4
	case string(umatsakramen.JenisImamat):
		return 5
	case string(umatsakramen.JenisPerminyakan):
		return 6
	case string(umatsakramen.JenisKematian):
		return 7
	default:
		return 99
	}
}

// UpsertSakramen idempotently upserts an UmatSakramen keyed by (parokiID, umatID, req.Jenis).
func (s *Service) UpsertSakramen(
	ctx context.Context,
	parokiID, umatID uuid.UUID,
	req domain.UpsertUmatSakramenRequest,
) (domain.UmatSakramenResponse, error) {
	if _, err := s.repo.FindUmatByID(ctx, parokiID, umatID); ent.IsNotFound(err) {
		return domain.UmatSakramenResponse{}, domain.ErrUmatNotFound
	} else if err != nil {
		return domain.UmatSakramenResponse{}, err
	}

	tanggal, err := parseOptDate("tanggal", req.Tanggal)
	if err != nil {
		return domain.UmatSakramenResponse{}, err
	}

	keuskupanID, parokiSakramenID, clearKeuskupan, clearParokiSakramen, err := s.resolveSakramenRefs(
		ctx, req.KeuskupanID, req.ParokiSakramenID, nil, nil,
	)
	if err != nil {
		return domain.UmatSakramenResponse{}, err
	}

	jenis := umatsakramen.Jenis(req.Jenis)
	existing, err := s.repo.FindSakramenByJenis(ctx, parokiID, umatID, jenis)
	if err == nil {
		updated, updErr := s.repo.UpdateSakramen(ctx, existing.ID, func(u *ent.UmatSakramenUpdateOne) *ent.UmatSakramenUpdateOne {
			applySakramenUpdates(u, tanggal, keuskupanID, parokiSakramenID, clearKeuskupan, clearParokiSakramen,
				req.KeuskupanNama, req.ParokiNama, req.Buku, req.Nomor, req.Catatan)
			return u
		})
		if updErr != nil {
			return domain.UmatSakramenResponse{}, updErr
		}
		return toSakramenResponse(updated), nil
	}
	if !ent.IsNotFound(err) {
		return domain.UmatSakramenResponse{}, err
	}

	created, err := s.repo.CreateSakramen(ctx, func(c *ent.UmatSakramenCreate) *ent.UmatSakramenCreate {
		c.SetParokiID(parokiID).
			SetUmatID(umatID).
			SetJenis(jenis).
			SetNillableTanggal(nilIfEmpty(tanggal)).
			SetNillableKeuskupanID(keuskupanID).
			SetNillableParokiSakramenID(parokiSakramenID).
			SetNillableKeuskupanNama(trimOptString(req.KeuskupanNama)).
			SetNillableParokiNama(trimOptString(req.ParokiNama)).
			SetNillableBuku(trimOptString(req.Buku)).
			SetNillableNomor(trimOptString(req.Nomor)).
			SetNillableCatatan(trimOptString(req.Catatan))
		return c
	})
	if err != nil {
		if ent.IsConstraintError(err) {
			return domain.UmatSakramenResponse{}, domain.ErrSakramenConflict
		}
		return domain.UmatSakramenResponse{}, err
	}
	return toSakramenResponse(created), nil
}

// UpdateSakramen partially updates an existing UmatSakramen row.
func (s *Service) UpdateSakramen(
	ctx context.Context,
	parokiID, umatID, sakramenID uuid.UUID,
	req domain.UpdateUmatSakramenRequest,
) (domain.UmatSakramenResponse, error) {
	existing, err := s.repo.FindSakramenByID(ctx, parokiID, umatID, sakramenID)
	if ent.IsNotFound(err) {
		return domain.UmatSakramenResponse{}, domain.ErrSakramenNotFound
	}
	if err != nil {
		return domain.UmatSakramenResponse{}, err
	}

	if req.Jenis != nil && *req.Jenis != string(existing.Jenis) {
		conflict, err := s.repo.HasSakramenByJenis(ctx, parokiID, umatID, umatsakramen.Jenis(*req.Jenis), &sakramenID)
		if err != nil {
			return domain.UmatSakramenResponse{}, err
		}
		if conflict {
			return domain.UmatSakramenResponse{}, domain.ErrSakramenConflict
		}
	}

	tanggal, err := parseOptDate("tanggal", req.Tanggal)
	if err != nil {
		return domain.UmatSakramenResponse{}, err
	}

	keuskupanID, parokiSakramenID, clearKeuskupan, clearParokiSakramen, err := s.resolveSakramenRefs(
		ctx, req.KeuskupanID, req.ParokiSakramenID, existing.KeuskupanID, existing.ParokiSakramenID,
	)
	if err != nil {
		return domain.UmatSakramenResponse{}, err
	}

	updated, err := s.repo.UpdateSakramen(ctx, existing.ID, func(u *ent.UmatSakramenUpdateOne) *ent.UmatSakramenUpdateOne {
		if req.Jenis != nil {
			u.SetJenis(umatsakramen.Jenis(*req.Jenis))
		}
		applySakramenUpdates(u, tanggal, keuskupanID, parokiSakramenID, clearKeuskupan, clearParokiSakramen,
			req.KeuskupanNama, req.ParokiNama, req.Buku, req.Nomor, req.Catatan)
		return u
	})
	if err != nil {
		if ent.IsConstraintError(err) {
			return domain.UmatSakramenResponse{}, domain.ErrSakramenConflict
		}
		return domain.UmatSakramenResponse{}, err
	}
	return toSakramenResponse(updated), nil
}

// DeleteSakramen physically hard-deletes an UmatSakramen row belonging to (parokiID, umatID).
func (s *Service) DeleteSakramen(ctx context.Context, parokiID, umatID, sakramenID uuid.UUID) error {
	existing, err := s.repo.FindSakramenByID(ctx, parokiID, umatID, sakramenID)
	if ent.IsNotFound(err) {
		return domain.ErrSakramenNotFound
	}
	if err != nil {
		return err
	}
	return s.repo.DeleteSakramen(ctx, existing.ID)
}

func (s *Service) resolveSakramenRefs(
	ctx context.Context,
	reqKeuskupanID, reqParokiSakramenID *uuid.UUID,
	curKeuskupanID, curParokiSakramenID *uuid.UUID,
) (keuskupanID, parokiSakramenID *uuid.UUID, clearKeuskupan, clearParokiSakramen bool, err error) {
	effKeuskupan := curKeuskupanID
	if reqKeuskupanID != nil {
		if *reqKeuskupanID == uuid.Nil {
			effKeuskupan = nil
			clearKeuskupan = true
		} else {
			if s.refs != nil {
				if _, err := s.refs.GetKeuskupan(ctx, *reqKeuskupanID); err != nil {
					return nil, nil, false, false, domain.ValidationError("keuskupanId", "INVALID_REF", "Keuskupan not found.")
				}
			}
			effKeuskupan = reqKeuskupanID
			keuskupanID = reqKeuskupanID
		}
	}

	effParoki := curParokiSakramenID
	if reqParokiSakramenID != nil {
		if *reqParokiSakramenID == uuid.Nil {
			clearParokiSakramen = true
		} else {
			if s.refs != nil {
				p, err := s.refs.GetParoki(ctx, *reqParokiSakramenID)
				if err != nil {
					return nil, nil, false, false, domain.ValidationError("parokiSakramenId", "INVALID_REF", "Paroki sakramen not found.")
				}
				if effKeuskupan == nil {
					kID := p.KeuskupanID
					keuskupanID = &kID
					clearKeuskupan = false
				} else if p.KeuskupanID != *effKeuskupan {
					return nil, nil, false, false, domain.ValidationError("keuskupanId", "KEUSKUPAN_MISMATCH", "Keuskupan does not match paroki sakramen.")
				}
			}
			parokiSakramenID = reqParokiSakramenID
		}
	} else if reqKeuskupanID != nil && *reqKeuskupanID != uuid.Nil && effParoki != nil && s.refs != nil {
		p, err := s.refs.GetParoki(ctx, *effParoki)
		if err == nil && p.KeuskupanID != *reqKeuskupanID {
			return nil, nil, false, false, domain.ValidationError("keuskupanId", "KEUSKUPAN_MISMATCH", "Keuskupan does not match paroki sakramen.")
		}
	}

	return keuskupanID, parokiSakramenID, clearKeuskupan, clearParokiSakramen, nil
}

func applySakramenUpdates(
	u *ent.UmatSakramenUpdateOne,
	tanggal *string,
	keuskupanID, parokiSakramenID *uuid.UUID,
	clearKeuskupan, clearParokiSakramen bool,
	keuskupanNama, parokiNama, buku, nomor, catatan *string,
) {
	if clearKeuskupan {
		u.ClearKeuskupanID()
	} else if keuskupanID != nil {
		u.SetKeuskupanID(*keuskupanID)
	}
	if clearParokiSakramen {
		u.ClearParokiSakramenID()
	} else if parokiSakramenID != nil {
		u.SetParokiSakramenID(*parokiSakramenID)
	}
	setOrClearSakramenStr(tanggal, u.SetTanggal, u.ClearTanggal)
	setOrClearSakramenStr(keuskupanNama, u.SetKeuskupanNama, u.ClearKeuskupanNama)
	setOrClearSakramenStr(parokiNama, u.SetParokiNama, u.ClearParokiNama)
	setOrClearSakramenStr(buku, u.SetBuku, u.ClearBuku)
	setOrClearSakramenStr(nomor, u.SetNomor, u.ClearNomor)
	setOrClearSakramenStr(catatan, u.SetCatatan, u.ClearCatatan)
}

func setOrClearSakramenStr(val *string, setFn func(string) *ent.UmatSakramenUpdateOne, clearFn func() *ent.UmatSakramenUpdateOne) {
	if val == nil {
		return
	}
	trimmed := strings.TrimSpace(*val)
	if trimmed == "" {
		clearFn()
		return
	}
	setFn(trimmed)
}

func toSakramenResponse(sk *ent.UmatSakramen) domain.UmatSakramenResponse {
	return domain.UmatSakramenResponse{
		ID:               sk.ID,
		ParokiID:         sk.ParokiID,
		UmatID:           sk.UmatID,
		Jenis:            string(sk.Jenis),
		Tanggal:          sk.Tanggal,
		KeuskupanID:      sk.KeuskupanID,
		ParokiSakramenID: sk.ParokiSakramenID,
		KeuskupanNama:    sk.KeuskupanNama,
		ParokiNama:       sk.ParokiNama,
		Buku:             sk.Buku,
		Nomor:            sk.Nomor,
		Catatan:          sk.Catatan,
		CreatedAt:        sk.CreatedAt,
		UpdatedAt:        sk.UpdatedAt,
	}
}
