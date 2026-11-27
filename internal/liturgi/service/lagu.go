package service

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/internal/liturgi/domain"
	"github.com/paroki/domus/internal/liturgi/repo"
	"github.com/paroki/domus/internal/platform/authx"
)

func (s *Service) canManageLagu(ctx context.Context, p authx.Principal, eventID uuid.UUID) (bool, error) {
	isAdmin, err := s.isLiturgiOrParokiAdmin(p)
	if err != nil || isAdmin {
		return isAdmin, err
	}
	if ok, err := s.canManageRole(ctx, p, "kor"); err != nil || ok {
		return ok, err
	}
	if ok, err := s.canManageRole(ctx, p, "organis"); err != nil || ok {
		return ok, err
	}
	rows, err := s.repo.ListActivePetugasForEvent(ctx, eventID)
	if err != nil {
		return false, err
	}
	for _, row := range rows {
		if row.Role != "organis" && row.Role != "kor" {
			continue
		}
		if row.UserID != nil && *row.UserID == p.UserID {
			return true, nil
		}
		if row.OrgID != nil && s.orgs != nil {
			adm, err := s.orgs.IsAdmin(ctx, *row.OrgID, p.UserID)
			if err != nil {
				return false, err
			}
			if adm {
				return true, nil
			}
		}
	}
	return false, nil
}

func (s *Service) validateSongPart(ctx context.Context, parokiID uuid.UUID, bagian string) (string, error) {
	b := strings.TrimSpace(bagian)
	if !domain.IsValidSongPartSlug(b) {
		return "", domain.ValidationError("bagian", "INVALID_SONG_PART", "Invalid liturgical song part slug.")
	}
	if domain.IsDefaultSongPart(b) {
		return b, nil
	}
	if s.refs != nil {
		ok, err := s.refs.HasActiveLookup(ctx, parokiID, "liturgi_song_part", b)
		if err != nil {
			return "", err
		}
		if ok {
			return b, nil
		}
	}
	return "", domain.ValidationError("bagian", "INVALID_SONG_PART", "Unknown liturgical song part slug.")
}

func (s *Service) resolveCantus(ctx context.Context, cantusID *uuid.UUID) (*CantusInfo, error) {
	if cantusID == nil || *cantusID == uuid.Nil {
		return nil, nil
	}
	if s.cantus == nil {
		return nil, domain.ValidationError("cantusId", "NOT_FOUND", "Cantus song not found.")
	}
	c, err := s.cantus.GetCantus(ctx, *cantusID)
	if err != nil || c == nil {
		return nil, domain.ValidationError("cantusId", "NOT_FOUND", "Cantus song not found.")
	}
	return c, nil
}

// CreateLagu adds a liturgical song entry to a parish Mass schedule.
func (s *Service) CreateLagu(
	ctx context.Context,
	p authx.Principal,
	eventID uuid.UUID,
	req domain.CreateLiturgiLaguRequest,
) (domain.LiturgiLaguResponse, error) {
	if _, err := s.ensureMassScheduleExists(ctx, p, eventID); err != nil {
		return domain.LiturgiLaguResponse{}, err
	}
	allowed, err := s.canManageLagu(ctx, p, eventID)
	if err != nil {
		return domain.LiturgiLaguResponse{}, err
	}
	if !allowed {
		return domain.LiturgiLaguResponse{}, domain.ErrForbidden
	}

	bagian, err := s.validateSongPart(ctx, p.ParokiID, req.Bagian)
	if err != nil {
		return domain.LiturgiLaguResponse{}, err
	}

	c, err := s.resolveCantus(ctx, req.CantusID)
	if err != nil {
		return domain.LiturgiLaguResponse{}, err
	}

	judul := strings.TrimSpace(req.Judul)
	sumber := trimOptionalPtr(req.Sumber)
	nomor := trimOptionalPtr(req.Nomor)
	nadaDasar := trimOptionalPtr(req.NadaDasar)
	var cantusID *uuid.UUID
	if c != nil {
		id := c.ID
		cantusID = &id
		if judul == "" {
			judul = strings.TrimSpace(c.Judul)
		}
		if sumber == nil {
			sumber = trimOptionalPtr(c.Sumber)
		}
		if nomor == nil {
			nomor = trimOptionalPtr(c.Nomor)
		}
		if nadaDasar == nil {
			nadaDasar = trimOptionalPtr(c.NadaDasar)
		}
	}
	if judul == "" {
		return domain.LiturgiLaguResponse{}, domain.ValidationError("judul", "REQUIRED", "Song title is required when cantusId is not provided.")
	}

	urutan := 0
	if req.Urutan != nil {
		urutan = *req.Urutan
	}

	created, err := s.repo.CreateLagu(ctx, repo.CreateLaguInput{
		ParokiID:  p.ParokiID,
		EventID:   eventID,
		Urutan:    urutan,
		Bagian:    bagian,
		CantusID:  cantusID,
		Sumber:    sumber,
		Nomor:     nomor,
		Judul:     judul,
		NadaDasar: nadaDasar,
		Note:      trimOptionalPtr(req.Note),
	})
	if err != nil {
		return domain.LiturgiLaguResponse{}, err
	}
	return toLaguResponse(created), nil
}

// UpdateLagu updates a liturgical song entry on a parish Mass schedule.
func (s *Service) UpdateLagu(
	ctx context.Context,
	p authx.Principal,
	eventID, laguID uuid.UUID,
	req domain.UpdateLiturgiLaguRequest,
) (domain.LiturgiLaguResponse, error) {
	if _, err := s.ensureMassScheduleExists(ctx, p, eventID); err != nil {
		return domain.LiturgiLaguResponse{}, err
	}
	existing, err := s.repo.FindLaguByID(ctx, p.ParokiID, eventID, laguID)
	if err != nil {
		if ent.IsNotFound(err) {
			return domain.LiturgiLaguResponse{}, domain.ErrLaguNotFound
		}
		return domain.LiturgiLaguResponse{}, err
	}
	allowed, err := s.canManageLagu(ctx, p, eventID)
	if err != nil {
		return domain.LiturgiLaguResponse{}, err
	}
	if !allowed {
		return domain.LiturgiLaguResponse{}, domain.ErrForbidden
	}

	var upd repo.UpdateLaguInput
	if req.Urutan != nil {
		upd.Urutan = req.Urutan
	}
	if req.Bagian != nil {
		bagian, err := s.validateSongPart(ctx, p.ParokiID, *req.Bagian)
		if err != nil {
			return domain.LiturgiLaguResponse{}, err
		}
		upd.Bagian = &bagian
	}

	if req.ClearCantusID != nil && *req.ClearCantusID {
		upd.ClearCantusID = true
	} else if req.CantusID != nil && *req.CantusID != uuid.Nil {
		c, err := s.resolveCantus(ctx, req.CantusID)
		if err != nil {
			return domain.LiturgiLaguResponse{}, err
		}
		id := c.ID
		upd.CantusID = &id
		if req.Judul == nil || strings.TrimSpace(*req.Judul) == "" {
			j := strings.TrimSpace(c.Judul)
			upd.Judul = &j
		}
		if req.Sumber == nil && c.Sumber != nil {
			upd.Sumber = trimOptionalPtr(c.Sumber)
		}
		if req.Nomor == nil && c.Nomor != nil {
			upd.Nomor = trimOptionalPtr(c.Nomor)
		}
		if req.NadaDasar == nil && c.NadaDasar != nil {
			upd.NadaDasar = trimOptionalPtr(c.NadaDasar)
		}
	}

	if req.Judul != nil && strings.TrimSpace(*req.Judul) != "" {
		j := strings.TrimSpace(*req.Judul)
		upd.Judul = &j
	} else if req.Judul != nil && upd.Judul == nil {
		return domain.LiturgiLaguResponse{}, domain.ValidationError("judul", "REQUIRED", "Song title cannot be empty.")
	}

	if req.Sumber != nil {
		if strings.TrimSpace(*req.Sumber) == "" {
			upd.ClearSumber = true
		} else {
			v := strings.TrimSpace(*req.Sumber)
			upd.Sumber = &v
		}
	}
	if req.Nomor != nil {
		if strings.TrimSpace(*req.Nomor) == "" {
			upd.ClearNomor = true
		} else {
			v := strings.TrimSpace(*req.Nomor)
			upd.Nomor = &v
		}
	}
	if req.NadaDasar != nil {
		if strings.TrimSpace(*req.NadaDasar) == "" {
			upd.ClearNada = true
		} else {
			v := strings.TrimSpace(*req.NadaDasar)
			upd.NadaDasar = &v
		}
	}
	if req.Note != nil {
		if strings.TrimSpace(*req.Note) == "" {
			upd.ClearNote = true
		} else {
			v := strings.TrimSpace(*req.Note)
			upd.Note = &v
		}
	}

	updated, err := s.repo.UpdateLagu(ctx, existing.ID, upd)
	if err != nil {
		return domain.LiturgiLaguResponse{}, err
	}
	return toLaguResponse(updated), nil
}

// DeleteLagu hard-deletes a liturgical song entry from a parish Mass schedule.
func (s *Service) DeleteLagu(ctx context.Context, p authx.Principal, eventID, laguID uuid.UUID) error {
	if _, err := s.ensureMassScheduleExists(ctx, p, eventID); err != nil {
		return err
	}
	existing, err := s.repo.FindLaguByID(ctx, p.ParokiID, eventID, laguID)
	if err != nil {
		if ent.IsNotFound(err) {
			return domain.ErrLaguNotFound
		}
		return err
	}
	allowed, err := s.canManageLagu(ctx, p, eventID)
	if err != nil {
		return err
	}
	if !allowed {
		return domain.ErrForbidden
	}
	return s.repo.DeleteLagu(ctx, existing.ID)
}

// IsCantusInUse reports whether cantusID is referenced by any liturgi_lagu row of a non-deleted Event.
func (s *Service) IsCantusInUse(ctx context.Context, cantusID uuid.UUID) (bool, error) {
	return s.repo.IsCantusInUse(ctx, cantusID)
}
