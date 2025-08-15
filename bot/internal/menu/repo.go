package menu

//
//import (
//	"context"
//	"encoding/json"
//
//	"github.com/jackc/pgx/v5/pgxpool"
//)
//
//type Button struct {
//	Text string
//	Next string // code следующего узла
//}
//
//type Screen struct {
//	Title   string
//	Slug    *string
//	Buttons []Button
//}
//
//type Repo struct {
//	DB *pgxpool.Pool
//}
//
//func NewRepo(db *pgxpool.Pool) *Repo { return &Repo{DB: db} }
//
//func (r *Repo) GetScreen(ctx context.Context, code, lang string) (*Screen, error) {
//	const q = `
//WITH node AS (
//  SELECT id, code, title_key, slug
//  FROM "узлы_меню"
//  WHERE code=$1 AND active=true
//),
//title AS (
//  SELECT COALESCE(p.text, n.title_key) AS title, n.slug, n.id
//  FROM node n
//  LEFT JOIN "переводы" p ON p.key = n.title_key AND p.lang = $2
//),
//btns AS (
//  SELECT k.id, COALESCE(p.text, k.text_key) AS text, n2.code AS next_code, k."order"
//  FROM "кнопки" k
//  JOIN title t ON t.id = k.node_id
//  LEFT JOIN "переводы" p ON p.key = k.text_key AND p.lang = $2
//  LEFT JOIN "узлы_меню" n2 ON n2.id = k.next_node_id
//)
//SELECT (SELECT title FROM title) AS title,
//       (SELECT slug  FROM title) AS slug,
//       COALESCE((
//         SELECT json_agg(json_build_object('text', text, 'next', next_code) ORDER BY "order")
//         FROM btns
//       ), '[]'::json) AS buttons;
//`
//	var title string
//	var slug *string
//	var buttonsJSON []byte
//
//	if err := r.DB.QueryRow(ctx, q, code, lang).Scan(&title, &slug, &buttonsJSON); err != nil {
//		return nil, err
//	}
//
//	var raw []struct{ Text, Next string }
//	if err := json.Unmarshal(buttonsJSON, &raw); err != nil {
//		return nil, err
//	}
//
//	scr := &Screen{Title: title, Slug: slug}
//	for _, b := range raw {
//		if b.Next == "" {
//			continue
//		}
//		scr.Buttons = append(scr.Buttons, Button{Text: b.Text, Next: b.Next})
//	}
//	return scr, nil
//}
