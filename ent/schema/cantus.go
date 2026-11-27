package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// CantusMediaItem stores a single media attachment in cantus.media JSONB.
type CantusMediaItem struct {
	Type        string `json:"type"`
	Title       string `json:"title,omitempty"`
	ViewURL     string `json:"viewUrl,omitempty"`
	DownloadURL string `json:"downloadUrl,omitempty"`
	S3Key       string `json:"s3Key,omitempty"`
}

// Cantus stores the global cross-parish liturgical and profane song/psalm catalog.
type Cantus struct {
	ent.Schema
}

// Annotations returns the table-level SQL annotations for Cantus.
func (Cantus) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cantus"},
	}
}

// Mixin returns the shared mixins for Cantus.
func (Cantus) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
		SoftDeleteMixin{},
	}
}

// Fields returns the fields of the Cantus entity.
func (Cantus) Fields() []ent.Field {
	return []ent.Field{
		field.String("slug").NotEmpty(),
		field.Enum("jenis").Values("lagu", "mazmur", "profan"),
		field.String("sumber").Optional().Nillable(),
		field.String("nomor").Optional().Nillable(),
		field.String("judul").NotEmpty(),
		field.String("kategori").Optional().Nillable(),
		field.JSON("tag", []string{}).Default([]string{}),
		field.String("nada_dasar").Optional().Nillable(),
		field.Text("lirik").Optional().Nillable(),
		field.JSON("kredit", map[string]string{}).Optional(),
		field.JSON("media", []CantusMediaItem{}).Optional(),
		field.JSON("meta", map[string]any{}).Optional(),
		field.Bool("active").Default(true),
		field.UUID("created_by", uuid.UUID{}).Optional().Nillable(),
	}
}

// Indexes returns the indexes for the Cantus entity.
func (Cantus) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("slug").
			Unique().
			Annotations(entsql.IndexWhere("deleted_at IS NULL")).
			StorageKey("cantus_slug_active_unique"),
		index.Fields("sumber", "nomor"),
		index.Fields("jenis", "sumber", "active"),
		index.Fields("judul"),
	}
}
