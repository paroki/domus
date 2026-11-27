package handler

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/cantus/domain"
	"github.com/paroki/domus/internal/cantus/service"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/httpx"
)

// Handler exposes HTTP endpoints for the global cantus catalog.
type Handler struct {
	service *service.Service
}

// New constructs a Handler backed by svc.
func New(svc *service.Service) *Handler {
	return &Handler{service: svc}
}

// Register mounts /cantus endpoints under the /api router group.
func (h *Handler) Register(
	api fiber.Router,
	auth, requireAuth, requireRead, requireWrite fiber.Handler,
) {
	group := api.Group("/cantus", auth, requireAuth)
	group.Post("/search", requireRead, h.Search)
	group.Get("/:id", requireRead, h.Get)
	group.Post("/", requireWrite, h.Create)
	group.Patch("/:id", requireWrite, h.Update)
	group.Delete("/:id", requireWrite, h.Delete)
}

// Search searches the global cantus catalog.
//
//	@Summary		Search cantus
//	@Description	Searches global cantus entries with query, filters, and keyset cursor pagination.
//	@ID				searchCantus
//	@Tags			Cantus
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.SearchCantusRequest	false	"Search parameters"
//	@Success		200		{object}	httpx.Envelope[[]domain.CantusResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Router			/cantus/search [post]
func (h *Handler) Search(c fiber.Ctx) error {
	var req domain.SearchCantusRequest
	if err := decodeEmptyOrStrictJSON(c, &req); err != nil {
		return err
	}
	searchHash := computeSearchHash(req.Q, req.Filter)
	cursorT, cursorID, limit, err := parseSearchCursorAndLimit(req.Cursor, req.Limit, searchHash)
	if err != nil {
		return err
	}
	items, hasMore, err := h.service.Search(c.Context(), actorIDFromCtx(c), req, cursorT, cursorID, limit)
	if err != nil {
		return err
	}
	var nextCursor string
	if hasMore && len(items) > 0 {
		last := items[len(items)-1]
		nextCursor = encodeSearchCursor(last.CreatedAt, last.ID.String(), searchHash)
	}
	return httpx.List(c, items, httpx.Meta{
		NextCursor: nextCursor,
		HasMore:    hasMore,
		Limit:      limit,
	})
}

// Get fetches a single cantus entry by ID.
//
//	@Summary		Get cantus
//	@Description	Returns a single non-deleted cantus entry by ID.
//	@ID				getCantus
//	@Tags			Cantus
//	@Produce		json
//	@Param			id	path		string	true	"Cantus UUID"
//	@Success		200	{object}	httpx.Envelope[domain.CantusResponse]
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/cantus/{id} [get]
func (h *Handler) Get(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid cantus ID.")
	}
	res, err := h.service.Get(c.Context(), actorIDFromCtx(c), id)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// Create creates a new cantus entry or reactivates a soft-deleted entry with the same slug.
//
//	@Summary		Create cantus
//	@Description	Creates a global cantus entry or reactivates a soft-deleted entry sharing the same slug.
//	@ID				createCantus
//	@Tags			Cantus
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.CreateCantusRequest	true	"Cantus creation payload"
//	@Success		201		{object}	httpx.Envelope[domain.CantusResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/cantus [post]
func (h *Handler) Create(c fiber.Ctx) error {
	var req domain.CreateCantusRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.Create(c.Context(), actorIDFromCtx(c), req)
	if err != nil {
		return err
	}
	return httpx.Created(c, res)
}

// Update partially updates a cantus entry.
//
//	@Summary		Update cantus
//	@Description	Partially updates a non-deleted cantus entry by ID.
//	@ID				updateCantus
//	@Tags			Cantus
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string						true	"Cantus UUID"
//	@Param			body	body		domain.UpdateCantusRequest	true	"Cantus update payload"
//	@Success		200		{object}	httpx.Envelope[domain.CantusResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/cantus/{id} [patch]
func (h *Handler) Update(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid cantus ID.")
	}
	var req domain.UpdateCantusRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.Update(c.Context(), actorIDFromCtx(c), id, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// Delete soft-deletes a cantus entry if not in use.
//
//	@Summary		Delete cantus
//	@Description	Soft-deletes a cantus entry after checking in-use references.
//	@ID				deleteCantus
//	@Tags			Cantus
//	@Param			id	path	string	true	"Cantus UUID"
//	@Success		204
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Failure		409	{object}	httpx.Envelope[any]
//	@Router			/cantus/{id} [delete]
func (h *Handler) Delete(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid cantus ID.")
	}
	if err := h.service.Delete(c.Context(), id); err != nil {
		return err
	}
	return httpx.NoContent(c)
}

func actorIDFromCtx(c fiber.Ctx) uuid.UUID {
	if p, ok := authx.UserContext(c.Context()); ok {
		return p.UserID
	}
	return uuid.Nil
}

func decodeStrictJSON(c fiber.Ctx, dst any) error {
	body := bytes.TrimSpace(c.Body())
	if len(body) > 0 {
		dec := json.NewDecoder(bytes.NewReader(body))
		dec.DisallowUnknownFields()
		if err := dec.Decode(dst); err != nil {
			return httpx.BadRequest("Invalid request body.")
		}
		if err := dec.Decode(&struct{}{}); err != io.EOF {
			return httpx.BadRequest("Invalid request body.")
		}
	}
	if v := c.App().Config().StructValidator; v != nil {
		if err := v.Validate(dst); err != nil {
			return err
		}
	}
	return nil
}

func decodeEmptyOrStrictJSON(c fiber.Ctx, dst any) error {
	if len(bytes.TrimSpace(c.Body())) == 0 {
		return nil
	}
	return decodeStrictJSON(c, dst)
}

type cantusCursor struct {
	CreatedAt time.Time `json:"t"`
	ID        string    `json:"id"`
	Hash      string    `json:"h"`
}

func computeSearchHash(q string, filter *domain.CantusSearchFilter) string {
	var f domain.CantusSearchFilter
	if filter != nil {
		f = *filter
		f.Slug = strings.ToLower(strings.TrimSpace(f.Slug))
		f.Jenis = strings.ToLower(strings.TrimSpace(f.Jenis))
		f.Sumber = strings.TrimSpace(f.Sumber)
		f.Kategori = strings.TrimSpace(f.Kategori)
		f.Tag = strings.TrimSpace(f.Tag)
		f.NadaDasar = strings.TrimSpace(f.NadaDasar)
	}
	payload, _ := json.Marshal(struct {
		Q      string                    `json:"q"`
		Filter domain.CantusSearchFilter `json:"filter"`
	}{
		Q:      strings.TrimSpace(q),
		Filter: f,
	})
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:8])
}

func encodeSearchCursor(createdAt time.Time, id, hash string) string {
	raw, _ := json.Marshal(cantusCursor{
		CreatedAt: createdAt.UTC(),
		ID:        id,
		Hash:      hash,
	})
	return base64.RawURLEncoding.EncodeToString(raw)
}

func parseSearchCursorAndLimit(rawCursor string, rawLimit int, expectedHash string) (*time.Time, *uuid.UUID, int, error) {
	limit := rawLimit
	if limit <= 0 {
		limit = httpx.DefaultLimit
	}
	if limit > httpx.MaxLimit {
		limit = httpx.MaxLimit
	}
	if rawCursor == "" {
		return nil, nil, limit, nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(rawCursor)
	if err != nil {
		return nil, nil, 0, httpx.InvalidCursor("Invalid pagination cursor.")
	}
	var cur cantusCursor
	if err := json.Unmarshal(decoded, &cur); err != nil || cur.ID == "" || cur.CreatedAt.IsZero() {
		return nil, nil, 0, httpx.InvalidCursor("Invalid pagination cursor.")
	}
	if cur.Hash != expectedHash {
		return nil, nil, 0, httpx.InvalidCursor("Pagination cursor does not match search query.")
	}
	id, err := uuid.Parse(cur.ID)
	if err != nil {
		return nil, nil, 0, httpx.InvalidCursor("Invalid pagination cursor.")
	}
	t := cur.CreatedAt
	return &t, &id, limit, nil
}
