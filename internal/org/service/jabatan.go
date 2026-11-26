package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/internal/org/domain"
	"github.com/paroki/domus/internal/org/repo"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/umat"
)

// SearchOrgJabatan returns paginated non-deleted OrgJabatan rows for orgID.
func (s *Service) SearchOrgJabatan(
	ctx context.Context,
	p authx.Principal,
	orgID uuid.UUID,
	req domain.SearchOrgJabatanRequest,
	cursorUrutan *int,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]domain.OrgJabatanResponse, bool, error) {
	if _, err := s.requireOrgPeriodeRead(ctx, p, orgID); err != nil {
		return nil, false, err
	}

	var periodeID, parentID, userID, umatID *uuid.UUID
	var jabatan string
	var active *bool
	if req.Filter != nil {
		periodeID = req.Filter.PeriodeID
		parentID = req.Filter.ParentID
		userID = req.Filter.UserID
		umatID = req.Filter.UmatID
		jabatan = strings.TrimSpace(req.Filter.Jabatan)
		active = req.Filter.Active
	}

	rows, err := s.repo.SearchOrgJabatan(
		ctx,
		orgID,
		strings.TrimSpace(req.Q),
		periodeID,
		parentID,
		userID,
		umatID,
		jabatan,
		active,
		cursorUrutan,
		cursorCreatedAt,
		cursorID,
		limit+1,
	)
	if err != nil {
		return nil, false, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	items := make([]domain.OrgJabatanResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, toOrgJabatanResponse(row))
	}
	return items, hasMore, nil
}

// GetOrgJabatan returns a single non-deleted OrgJabatan in orgID.
func (s *Service) GetOrgJabatan(ctx context.Context, p authx.Principal, orgID, jabatanID uuid.UUID) (domain.OrgJabatanResponse, error) {
	if _, err := s.requireOrgPeriodeRead(ctx, p, orgID); err != nil {
		return domain.OrgJabatanResponse{}, err
	}
	row, err := s.repo.FindOrgJabatanInOrg(ctx, orgID, jabatanID)
	if ent.IsNotFound(err) {
		return domain.OrgJabatanResponse{}, domain.ErrJabatanNotFound
	}
	if err != nil {
		return domain.OrgJabatanResponse{}, err
	}
	return toOrgJabatanResponse(row), nil
}

// CurrentOrgPeriodeID returns the ID of the active (is_current == true) OrgPeriode for orgID.
func (s *Service) CurrentOrgPeriodeID(ctx context.Context, p authx.Principal, orgID uuid.UUID) (uuid.UUID, error) {
	if _, err := s.requireOrgPeriodeRead(ctx, p, orgID); err != nil {
		return uuid.Nil, err
	}
	row, err := s.repo.FindCurrentOrgPeriode(ctx, orgID)
	if ent.IsNotFound(err) {
		return uuid.Nil, domain.ErrPeriodeNotFound
	}
	if err != nil {
		return uuid.Nil, err
	}
	return row.ID, nil
}

// CreateOrgJabatan creates a new position node in an Organization period with hierarchy validation and user/umat/nama auto-sync.
func (s *Service) CreateOrgJabatan(
	ctx context.Context,
	p authx.Principal,
	orgID uuid.UUID,
	req domain.CreateOrgJabatanRequest,
) (domain.OrgJabatanResponse, error) {
	orgRow, err := s.requireOrgPeriodeWrite(ctx, p, orgID)
	if err != nil {
		return domain.OrgJabatanResponse{}, err
	}

	if _, err := s.repo.FindOrgPeriodeInOrg(ctx, orgID, req.PeriodeID); ent.IsNotFound(err) {
		return domain.OrgJabatanResponse{}, domain.ErrPeriodeNotFound
	} else if err != nil {
		return domain.OrgJabatanResponse{}, err
	}

	jabatanTitle := strings.TrimSpace(req.Jabatan)
	if jabatanTitle == "" {
		return domain.OrgJabatanResponse{}, domain.ValidationError("jabatan", "REQUIRED", "Jabatan cannot be empty.")
	}

	if err := s.validateJabatanParent(ctx, orgID, req.PeriodeID, uuid.Nil, req.ParentID); err != nil {
		return domain.OrgJabatanResponse{}, err
	}

	startDate, endDate, err := validateOptionalJabatanDates(req.StartDate, req.EndDate)
	if err != nil {
		return domain.OrgJabatanResponse{}, err
	}

	userID, umatID, nama, err := s.resolveJabatanPerson(ctx, orgRow.ParokiID, req.UserID, req.UmatID, req.Nama)
	if err != nil {
		return domain.OrgJabatanResponse{}, err
	}

	urutan := 0
	if req.Urutan != nil {
		urutan = *req.Urutan
	}
	active := true
	if req.Active != nil {
		active = *req.Active
	}

	created, err := s.repo.CreateOrgJabatan(ctx, repo.OrgJabatanInput{
		ParokiID:  orgRow.ParokiID,
		OrgID:     orgID,
		PeriodeID: req.PeriodeID,
		ParentID:  req.ParentID,
		Jabatan:   jabatanTitle,
		UserID:    userID,
		UmatID:    umatID,
		Nama:      nama,
		StartDate: startDate,
		EndDate:   endDate,
		Urutan:    urutan,
		Active:    active,
	})
	if err != nil {
		return domain.OrgJabatanResponse{}, err
	}
	return toOrgJabatanResponse(created), nil
}

// UpdateOrgJabatan updates mutable fields of an OrgJabatan with hierarchy validation and user/umat/nama auto-sync.
func (s *Service) UpdateOrgJabatan(
	ctx context.Context,
	p authx.Principal,
	orgID, jabatanID uuid.UUID,
	req domain.UpdateOrgJabatanRequest,
) (domain.OrgJabatanResponse, error) {
	orgRow, err := s.requireOrgPeriodeWrite(ctx, p, orgID)
	if err != nil {
		return domain.OrgJabatanResponse{}, err
	}

	existing, err := s.repo.FindOrgJabatanInOrg(ctx, orgID, jabatanID)
	if ent.IsNotFound(err) {
		return domain.OrgJabatanResponse{}, domain.ErrJabatanNotFound
	}
	if err != nil {
		return domain.OrgJabatanResponse{}, err
	}

	var jabatanPtr *string
	if req.Jabatan != nil {
		trimmed := strings.TrimSpace(*req.Jabatan)
		if trimmed == "" {
			return domain.OrgJabatanResponse{}, domain.ValidationError("jabatan", "REQUIRED", "Jabatan cannot be empty.")
		}
		jabatanPtr = &trimmed
	}

	if !req.ClearParentID && req.ParentID != nil && *req.ParentID != uuid.Nil {
		if err := s.validateJabatanParent(ctx, orgID, existing.PeriodeID, jabatanID, req.ParentID); err != nil {
			return domain.OrgJabatanResponse{}, err
		}
	}

	effStart := existing.StartDate
	if req.ClearStartDate || (req.StartDate != nil && strings.TrimSpace(*req.StartDate) == "") {
		effStart = nil
	} else if req.StartDate != nil {
		effStart = req.StartDate
	}
	effEnd := existing.EndDate
	if req.ClearEndDate || (req.EndDate != nil && strings.TrimSpace(*req.EndDate) == "") {
		effEnd = nil
	} else if req.EndDate != nil {
		effEnd = req.EndDate
	}
	normStart, normEnd, err := validateOptionalJabatanDates(effStart, effEnd)
	if err != nil {
		return domain.OrgJabatanResponse{}, err
	}
	var startPtr, endPtr *string
	if !req.ClearStartDate && req.StartDate != nil {
		if normStart == nil {
			empty := ""
			startPtr = &empty
		} else {
			startPtr = normStart
		}
	}
	if !req.ClearEndDate && req.EndDate != nil {
		if normEnd == nil {
			empty := ""
			endPtr = &empty
		} else {
			endPtr = normEnd
		}
	}

	effUserID := existing.UserID
	if req.ClearUserID || (req.UserID != nil && *req.UserID == uuid.Nil) {
		effUserID = nil
	} else if req.UserID != nil {
		effUserID = req.UserID
	}

	effUmatID := existing.UmatID
	if req.ClearUmatID || (req.UmatID != nil && *req.UmatID == uuid.Nil) {
		effUmatID = nil
	} else if req.UmatID != nil {
		effUmatID = req.UmatID
	}

	effNama := existing.Nama
	if req.ClearNama || (req.Nama != nil && strings.TrimSpace(*req.Nama) == "") {
		effNama = nil
	} else if req.Nama != nil {
		effNama = req.Nama
	} else if (req.UserID != nil || req.UmatID != nil) && !req.ClearNama {
		// Re-sync nama if person changed and caller did not explicitly override nama.
		if req.UserID != nil && (existing.UserID == nil || *req.UserID != *existing.UserID) {
			effNama = nil
		}
		if req.UmatID != nil && (existing.UmatID == nil || *req.UmatID != *existing.UmatID) {
			effNama = nil
		}
	}

	resUserID, resUmatID, resNama, err := s.resolveJabatanPerson(ctx, orgRow.ParokiID, effUserID, effUmatID, effNama)
	if err != nil {
		return domain.OrgJabatanResponse{}, err
	}

	clearUserID := resUserID == nil
	clearUmatID := resUmatID == nil
	clearNama := resNama == nil

	updated, err := s.repo.UpdateOrgJabatan(
		ctx,
		jabatanID,
		req.ParentID,
		req.ClearParentID,
		jabatanPtr,
		resUserID,
		clearUserID,
		resUmatID,
		clearUmatID,
		resNama,
		clearNama,
		startPtr,
		req.ClearStartDate,
		endPtr,
		req.ClearEndDate,
		req.Urutan,
		req.Active,
	)
	if err != nil {
		return domain.OrgJabatanResponse{}, err
	}
	return toOrgJabatanResponse(updated), nil
}

// DeleteOrgJabatan soft-deletes an OrgJabatan if it has no non-deleted child positions.
func (s *Service) DeleteOrgJabatan(ctx context.Context, p authx.Principal, orgID, jabatanID uuid.UUID) error {
	if _, err := s.requireOrgPeriodeWrite(ctx, p, orgID); err != nil {
		return err
	}
	if _, err := s.repo.FindOrgJabatanInOrg(ctx, orgID, jabatanID); ent.IsNotFound(err) {
		return domain.ErrJabatanNotFound
	} else if err != nil {
		return err
	}
	hasChildren, err := s.repo.HasOrgJabatanChildren(ctx, orgID, jabatanID)
	if err != nil {
		return err
	}
	if hasChildren {
		return domain.ErrInUse
	}
	return s.repo.SoftDeleteOrgJabatan(ctx, jabatanID, time.Now().UTC())
}

func (s *Service) validateJabatanParent(ctx context.Context, orgID, periodeID, selfID uuid.UUID, parentID *uuid.UUID) error {
	if parentID == nil || *parentID == uuid.Nil {
		return nil
	}
	if selfID != uuid.Nil && *parentID == selfID {
		return domain.ValidationError("parentId", "INVALID_PARENT", "Position cannot be its own parent.")
	}
	parent, err := s.repo.FindOrgJabatanInOrg(ctx, orgID, *parentID)
	if ent.IsNotFound(err) || (err == nil && parent.PeriodeID != periodeID) {
		return domain.ValidationError("parentId", "INVALID_PARENT", "Parent position must exist in the same organization and period.")
	}
	if err != nil {
		return err
	}
	if selfID != uuid.Nil {
		visited := map[uuid.UUID]struct{}{parent.ID: {}}
		cur := parent.ParentID
		for cur != nil && *cur != uuid.Nil {
			if *cur == selfID {
				return domain.ValidationError("parentId", "INVALID_PARENT", "Position hierarchy cannot contain a cycle.")
			}
			if _, seen := visited[*cur]; seen {
				break
			}
			visited[*cur] = struct{}{}
			anc, ancErr := s.repo.FindOrgJabatanInOrg(ctx, orgID, *cur)
			if ent.IsNotFound(ancErr) {
				break
			}
			if ancErr != nil {
				return ancErr
			}
			cur = anc.ParentID
		}
	}
	return nil
}

func (s *Service) resolveJabatanPerson(
	ctx context.Context,
	parokiID uuid.UUID,
	userID, umatID *uuid.UUID,
	rawNama *string,
) (*uuid.UUID, *uuid.UUID, *string, error) {
	if userID != nil && *userID == uuid.Nil {
		userID = nil
	}
	if umatID != nil && *umatID == uuid.Nil {
		umatID = nil
	}
	hadUserID := userID != nil
	hadUmatID := umatID != nil

	var userName string
	if userID != nil && s.users != nil {
		name, err := s.users.GetUserName(ctx, *userID)
		if err != nil {
			return nil, nil, nil, domain.ValidationError("userId", "INVALID_REF", "User not found.")
		}
		userName = name
	}

	var umatBasic *umat.UmatBasic
	if umatID != nil {
		if s.umats == nil {
			return nil, nil, nil, domain.ValidationError("umatId", "INVALID_REF", "Umat lookup is not configured.")
		}
		b, err := s.umats.GetBasic(ctx, parokiID, *umatID)
		if err != nil {
			return nil, nil, nil, domain.ValidationError("umatId", "INVALID_REF", "Umat not found in active paroki.")
		}
		umatBasic = b
	}

	if userID != nil && umatID != nil {
		if umatBasic.UserID != nil && *umatBasic.UserID != *userID {
			return nil, nil, nil, domain.ValidationError("umatId", "PERSON_MISMATCH", "Selected userId and umatId belong to different people.")
		}
		if s.umats != nil {
			linkedUmat, err := s.umats.GetByUserID(ctx, parokiID, *userID)
			if err != nil {
				return nil, nil, nil, err
			}
			if linkedUmat != nil && linkedUmat.ID != *umatID {
				return nil, nil, nil, domain.ValidationError("umatId", "PERSON_MISMATCH", "Selected userId is already linked to a different umat.")
			}
		}
	} else if userID != nil && umatID == nil && s.umats != nil {
		linkedUmat, err := s.umats.GetByUserID(ctx, parokiID, *userID)
		if err != nil {
			return nil, nil, nil, err
		}
		if linkedUmat != nil {
			id := linkedUmat.ID
			umatID = &id
			umatBasic = linkedUmat
		}
	} else if umatID != nil && userID == nil && umatBasic != nil && umatBasic.UserID != nil {
		uid := *umatBasic.UserID
		userID = &uid
		if s.users != nil {
			if name, err := s.users.GetUserName(ctx, uid); err == nil {
				userName = name
			}
		}
	}

	var nama *string
	if rawNama != nil {
		trimmed := strings.TrimSpace(*rawNama)
		if trimmed != "" {
			nama = &trimmed
		}
	}
	if nama == nil {
		switch {
		case hadUmatID && !hadUserID && umatBasic != nil && strings.TrimSpace(umatBasic.Nama) != "":
			n := strings.TrimSpace(umatBasic.Nama)
			nama = &n
		case userID != nil && strings.TrimSpace(userName) != "":
			n := strings.TrimSpace(userName)
			nama = &n
		case umatBasic != nil && strings.TrimSpace(umatBasic.Nama) != "":
			n := strings.TrimSpace(umatBasic.Nama)
			nama = &n
		}
	}
	return userID, umatID, nama, nil
}

func validateOptionalJabatanDates(rawStart, rawEnd *string) (*string, *string, error) {
	var startPtr, endPtr *string
	var startTime, endTime time.Time
	var hasStart, hasEnd bool

	if rawStart != nil {
		s := strings.TrimSpace(*rawStart)
		if s != "" {
			t, err := time.Parse("2006-01-02", s)
			if err != nil {
				return nil, nil, domain.ValidationError("startDate", "INVALID_DATE", "startDate must be formatted as YYYY-MM-DD.")
			}
			startPtr = &s
			startTime = t
			hasStart = true
		}
	}
	if rawEnd != nil {
		e := strings.TrimSpace(*rawEnd)
		if e != "" {
			t, err := time.Parse("2006-01-02", e)
			if err != nil {
				return nil, nil, domain.ValidationError("endDate", "INVALID_DATE", "endDate must be formatted as YYYY-MM-DD.")
			}
			endPtr = &e
			endTime = t
			hasEnd = true
		}
	}
	if hasStart && hasEnd && startTime.After(endTime) {
		return nil, nil, domain.ValidationError("endDate", "INVALID_DATE_RANGE", "endDate must be on or after startDate.")
	}
	return startPtr, endPtr, nil
}

func toOrgJabatanResponse(j *ent.OrgJabatan) domain.OrgJabatanResponse {
	return domain.OrgJabatanResponse{
		ID:        j.ID,
		ParokiID:  j.ParokiID,
		OrgID:     j.OrgID,
		PeriodeID: j.PeriodeID,
		ParentID:  j.ParentID,
		Jabatan:   j.Jabatan,
		UserID:    j.UserID,
		UmatID:    j.UmatID,
		Nama:      j.Nama,
		StartDate: j.StartDate,
		EndDate:   j.EndDate,
		Urutan:    j.Urutan,
		Active:    j.Active,
		CreatedAt: j.CreatedAt,
		UpdatedAt: j.UpdatedAt,
	}
}
