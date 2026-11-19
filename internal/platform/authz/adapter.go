package authz

import (
	"context"
	"fmt"
	"strings"

	"entgo.io/ent/dialect/sql"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/casbinrule"
	"github.com/paroki/domus/ent/predicate"
)

// EntAdapter implements Casbin's persist.Adapter backed by Ent's CasbinRule schema.
type EntAdapter struct {
	client *ent.Client
}

var _ persist.Adapter = (*EntAdapter)(nil)

// NewEntAdapter creates a new Ent-backed Casbin persistence adapter.
func NewEntAdapter(client *ent.Client) *EntAdapter {
	return &EntAdapter{client: client}
}

// LoadPolicy loads all policy and grouping rules from the database into model.
func (a *EntAdapter) LoadPolicy(m model.Model) error {
	rules, err := a.client.CasbinRule.Query().All(context.Background())
	if err != nil {
		return fmt.Errorf("load casbin rules: %w", err)
	}
	for _, r := range rules {
		if err := persist.LoadPolicyLine(formatRuleLine(r), m); err != nil {
			return fmt.Errorf("load policy line: %w", err)
		}
	}
	return nil
}

// SavePolicy replaces all stored Casbin rules with the current model policies.
func (a *EntAdapter) SavePolicy(m model.Model) error {
	ctx := context.Background()
	tx, err := a.client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin save policy tx: %w", err)
	}

	if _, err := tx.CasbinRule.Delete().Exec(ctx); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("clear casbin rules: %w", err)
	}

	var builders []*ent.CasbinRuleCreate
	for _, sec := range []string{"p", "g"} {
		for ptype, ast := range m[sec] {
			for _, rule := range ast.Policy {
				v := packRule(rule)
				builders = append(builders, tx.CasbinRule.Create().
					SetPtype(ptype).
					SetV0(v[0]).
					SetV1(v[1]).
					SetV2(v[2]).
					SetV3(v[3]).
					SetV4(v[4]).
					SetV5(v[5]))
			}
		}
	}

	if len(builders) > 0 {
		if _, err := tx.CasbinRule.CreateBulk(builders...).Save(ctx); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("bulk save casbin rules: %w", err)
		}
	}

	return tx.Commit()
}

// AddPolicy persists a single policy or grouping rule.
func (a *EntAdapter) AddPolicy(_ string, ptype string, rule []string) error {
	v := packRule(rule)
	_, err := a.client.CasbinRule.Create().
		SetPtype(ptype).
		SetV0(v[0]).
		SetV1(v[1]).
		SetV2(v[2]).
		SetV3(v[3]).
		SetV4(v[4]).
		SetV5(v[5]).
		Save(context.Background())
	return err
}

// RemovePolicy deletes a single policy or grouping rule.
func (a *EntAdapter) RemovePolicy(_ string, ptype string, rule []string) error {
	v := packRule(rule)
	_, err := a.client.CasbinRule.Delete().
		Where(
			casbinrule.PtypeEQ(ptype),
			casbinrule.V0EQ(v[0]),
			casbinrule.V1EQ(v[1]),
			casbinrule.V2EQ(v[2]),
			casbinrule.V3EQ(v[3]),
			casbinrule.V4EQ(v[4]),
			casbinrule.V5EQ(v[5]),
		).
		Exec(context.Background())
	return err
}

// RemoveFilteredPolicy deletes rules matching the given ptype and field values.
func (a *EntAdapter) RemoveFilteredPolicy(_ string, ptype string, fieldIndex int, fieldValues ...string) error {
	preds := []predicate.CasbinRule{casbinrule.PtypeEQ(ptype)}
	cols := []string{
		casbinrule.FieldV0,
		casbinrule.FieldV1,
		casbinrule.FieldV2,
		casbinrule.FieldV3,
		casbinrule.FieldV4,
		casbinrule.FieldV5,
	}

	for i, val := range fieldValues {
		idx := fieldIndex + i
		if idx < 0 || idx >= len(cols) || val == "" {
			continue
		}
		col := cols[idx]
		v := val
		preds = append(preds, predicate.CasbinRule(func(s *sql.Selector) {
			s.Where(sql.EQ(s.C(col), v))
		}))
	}

	_, err := a.client.CasbinRule.Delete().Where(preds...).Exec(context.Background())
	return err
}

func packRule(rule []string) [6]string {
	var v [6]string
	for i := 0; i < len(rule) && i < len(v); i++ {
		v[i] = rule[i]
	}
	return v
}

func formatRuleLine(r *ent.CasbinRule) string {
	parts := []string{r.Ptype}
	for _, v := range []string{r.V0, r.V1, r.V2, r.V3, r.V4, r.V5} {
		if v != "" {
			parts = append(parts, v)
		}
	}
	return strings.Join(parts, ", ")
}
