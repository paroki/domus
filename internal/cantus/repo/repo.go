package repo

import (
	"context"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"
	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/cantus"
	"github.com/paroki/domus/ent/predicate"
	"github.com/paroki/domus/ent/schema"
	"github.com/paroki/domus/internal/cantus/domain"
)

// Repository encapsulates Ent persistence for the global cantus catalog.
type Repository struct {
	client *ent.Client
}

// New constructs a Repository backed by client.
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
		if rbErr := tx.Rollback(); rbErr != nil {
			return rbErr
		}
		return err
	}
	return tx.Commit()
}

// FindByID returns the non-deleted Cantus matching id.
func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (*ent.Cantus, error) {
	return r.client.Cantus.Query().
		Where(cantus.IDEQ(id), cantus.DeletedAtIsNil()).
		Only(ctx)
}

// FindBySlug returns the non-deleted Cantus matching slug.
func (r *Repository) FindBySlug(ctx context.Context, slug string) (*ent.Cantus, error) {
	return r.client.Cantus.Query().
		Where(cantus.SlugEQ(slug), cantus.DeletedAtIsNil()).
		Only(ctx)
}

// FindBySlugIncludeDeleted returns any Cantus matching slug, including soft-deleted rows.
func (r *Repository) FindBySlugIncludeDeleted(ctx context.Context, slug string) (*ent.Cantus, error) {
	return r.client.Cantus.Query().
		Where(cantus.SlugEQ(slug)).
		Order(ent.Desc(cantus.FieldUpdatedAt)).
		First(ctx)
}

// Search returns non-deleted Cantus entries matching q and filter ordered by (created_at DESC, id DESC).
func (r *Repository) Search(
	ctx context.Context,
	q string,
	filter domain.CantusSearchFilter,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]*ent.Cantus, error) {
	query := r.client.Cantus.Query().Where(cantus.DeletedAtIsNil())

	if q != "" {
		likePattern := "%" + escapeLike(q) + "%"
		query = query.Where(cantus.Or(
			cantus.JudulContainsFold(q),
			cantus.NomorContainsFold(q),
			cantus.LirikContainsFold(q),
			predicate.Cantus(func(s *sql.Selector) {
				s.Where(sql.P(func(b *sql.Builder) {
					b.WriteString("CAST(" + s.C(cantus.FieldKredit) + " AS TEXT) ILIKE ")
					b.Arg(likePattern)
				}))
			}),
		))
	}
	if filter.Slug != "" {
		query = query.Where(cantus.SlugEQ(filter.Slug))
	}
	if filter.Jenis != "" {
		query = query.Where(cantus.JenisEQ(cantus.Jenis(filter.Jenis)))
	}
	if filter.Sumber != "" {
		query = query.Where(cantus.SumberEqualFold(filter.Sumber))
	}
	if filter.Kategori != "" {
		query = query.Where(cantus.KategoriEqualFold(filter.Kategori))
	}
	if filter.NadaDasar != "" {
		query = query.Where(cantus.NadaDasarEqualFold(filter.NadaDasar))
	}
	if filter.Active != nil {
		query = query.Where(cantus.ActiveEQ(*filter.Active))
	}
	if filter.Tag != "" {
		query = query.Where(predicate.Cantus(func(s *sql.Selector) {
			s.Where(sqljson.ValueContains(s.C(cantus.FieldTag), filter.Tag))
		}))
	}
	if cursorCreatedAt != nil && cursorID != nil {
		query = query.Where(cantus.Or(
			cantus.CreatedAtLT(*cursorCreatedAt),
			cantus.And(cantus.CreatedAtEQ(*cursorCreatedAt), cantus.IDLT(*cursorID)),
		))
	}

	return query.Order(ent.Desc(cantus.FieldCreatedAt), ent.Desc(cantus.FieldID)).
		Limit(limit).
		All(ctx)
}

// CreateParams holds normalized fields for creating or reactivating a Cantus row.
type CreateParams struct {
	Slug      string
	Jenis     cantus.Jenis
	Sumber    *string
	Nomor     *string
	Judul     string
	Kategori  *string
	Tag       []string
	NadaDasar *string
	Lirik     *string
	Kredit    map[string]string
	Media     []schema.CantusMediaItem
	Meta      map[string]any
	Active    bool
	CreatedBy *uuid.UUID
}

// Create inserts a new Cantus row.
func (r *Repository) Create(ctx context.Context, p CreateParams) (*ent.Cantus, error) {
	cre := r.client.Cantus.Create().
		SetSlug(p.Slug).
		SetJenis(p.Jenis).
		SetJudul(p.Judul).
		SetTag(p.Tag).
		SetActive(p.Active)
	if p.Sumber != nil {
		cre.SetSumber(*p.Sumber)
	}
	if p.Nomor != nil {
		cre.SetNomor(*p.Nomor)
	}
	if p.Kategori != nil {
		cre.SetKategori(*p.Kategori)
	}
	if p.NadaDasar != nil {
		cre.SetNadaDasar(*p.NadaDasar)
	}
	if p.Lirik != nil {
		cre.SetLirik(*p.Lirik)
	}
	if p.Kredit != nil {
		cre.SetKredit(p.Kredit)
	}
	if p.Media != nil {
		cre.SetMedia(p.Media)
	}
	if p.Meta != nil {
		cre.SetMeta(p.Meta)
	}
	if p.CreatedBy != nil && *p.CreatedBy != uuid.Nil {
		cre.SetCreatedBy(*p.CreatedBy)
	}
	return cre.Save(ctx)
}

// Reactivate restores a soft-deleted Cantus row in-place with fresh payload fields.
func (r *Repository) Reactivate(ctx context.Context, id uuid.UUID, p CreateParams) (*ent.Cantus, error) {
	upd := r.client.Cantus.UpdateOneID(id).
		ClearDeletedAt().
		SetSlug(p.Slug).
		SetJenis(p.Jenis).
		SetJudul(p.Judul).
		SetTag(p.Tag).
		SetActive(p.Active)

	if p.Sumber != nil {
		upd.SetSumber(*p.Sumber)
	} else {
		upd.ClearSumber()
	}
	if p.Nomor != nil {
		upd.SetNomor(*p.Nomor)
	} else {
		upd.ClearNomor()
	}
	if p.Kategori != nil {
		upd.SetKategori(*p.Kategori)
	} else {
		upd.ClearKategori()
	}
	if p.NadaDasar != nil {
		upd.SetNadaDasar(*p.NadaDasar)
	} else {
		upd.ClearNadaDasar()
	}
	if p.Lirik != nil {
		upd.SetLirik(*p.Lirik)
	} else {
		upd.ClearLirik()
	}
	if p.Kredit != nil {
		upd.SetKredit(p.Kredit)
	} else {
		upd.ClearKredit()
	}
	if p.Media != nil {
		upd.SetMedia(p.Media)
	} else {
		upd.ClearMedia()
	}
	if p.Meta != nil {
		upd.SetMeta(p.Meta)
	} else {
		upd.ClearMeta()
	}
	if p.CreatedBy != nil && *p.CreatedBy != uuid.Nil {
		upd.SetCreatedBy(*p.CreatedBy)
	}
	return upd.Save(ctx)
}

// UpdateParams holds optional normalized fields for updating a Cantus row.
type UpdateParams struct {
	Slug      *string
	Jenis     *cantus.Jenis
	Sumber    *string
	Nomor     *string
	Judul     *string
	Kategori  *string
	Tag       []string
	NadaDasar *string
	Lirik     *string
	Kredit    map[string]string
	Media     []schema.CantusMediaItem
	Meta      map[string]any
	Active    *bool
}

// Update mutates mutable fields on a non-deleted Cantus row.
func (r *Repository) Update(ctx context.Context, id uuid.UUID, p UpdateParams) (*ent.Cantus, error) {
	upd := r.client.Cantus.UpdateOneID(id)
	if p.Slug != nil {
		upd.SetSlug(*p.Slug)
	}
	if p.Jenis != nil {
		upd.SetJenis(*p.Jenis)
	}
	if p.Sumber != nil {
		if *p.Sumber == "" {
			upd.ClearSumber()
		} else {
			upd.SetSumber(*p.Sumber)
		}
	}
	if p.Nomor != nil {
		if *p.Nomor == "" {
			upd.ClearNomor()
		} else {
			upd.SetNomor(*p.Nomor)
		}
	}
	if p.Judul != nil {
		upd.SetJudul(*p.Judul)
	}
	if p.Kategori != nil {
		if *p.Kategori == "" {
			upd.ClearKategori()
		} else {
			upd.SetKategori(*p.Kategori)
		}
	}
	if p.Tag != nil {
		upd.SetTag(p.Tag)
	}
	if p.NadaDasar != nil {
		if *p.NadaDasar == "" {
			upd.ClearNadaDasar()
		} else {
			upd.SetNadaDasar(*p.NadaDasar)
		}
	}
	if p.Lirik != nil {
		if *p.Lirik == "" {
			upd.ClearLirik()
		} else {
			upd.SetLirik(*p.Lirik)
		}
	}
	if p.Kredit != nil {
		if len(p.Kredit) == 0 {
			upd.ClearKredit()
		} else {
			upd.SetKredit(p.Kredit)
		}
	}
	if p.Media != nil {
		if len(p.Media) == 0 {
			upd.ClearMedia()
		} else {
			upd.SetMedia(p.Media)
		}
	}
	if p.Meta != nil {
		if len(p.Meta) == 0 {
			upd.ClearMeta()
		} else {
			upd.SetMeta(p.Meta)
		}
	}
	if p.Active != nil {
		upd.SetActive(*p.Active)
	}
	return upd.Save(ctx)
}

// SoftDelete sets deleted_at on a Cantus row.
func (r *Repository) SoftDelete(ctx context.Context, id uuid.UUID, at time.Time) error {
	return r.client.Cantus.UpdateOneID(id).
		SetDeletedAt(at).
		Exec(ctx)
}

func escapeLike(s string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(s)
}
