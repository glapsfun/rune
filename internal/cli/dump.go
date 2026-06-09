package cli

import (
	"encoding/json"
	"fmt"

	"github.com/rune-task-runner/rune/internal/ast"
)

// dumpFile emits the parsed Runefile: canonical text (default) or JSON
// (--format json), for machine consumption (FR-023).
func dumpFile(opts Options, f *ast.File) error {
	if opts.DumpFormat == "json" {
		data, err := json.MarshalIndent(toDTO(f), "", "  ")
		if err != nil {
			return &UsageError{Err: err}
		}
		fmt.Fprintln(opts.Stdout, string(data))
		return nil
	}
	fmt.Fprint(opts.Stdout, ast.Dump(f))
	return nil
}

type fileDTO struct {
	Settings    map[string]any `json:"settings"`
	Assignments map[string]any `json:"assignments"`
	Tasks       []taskDTO      `json:"tasks"`
}

type taskDTO struct {
	Name       string     `json:"name"`
	Doc        string     `json:"doc,omitempty"`
	Executor   string     `json:"executor,omitempty"`
	Private    bool       `json:"private"`
	Params     []paramDTO `json:"params,omitempty"`
	Deps       []string   `json:"deps,omitempty"`
	PostHooks  []string   `json:"postHooks,omitempty"`
	Attributes []string   `json:"attributes,omitempty"`
	Body       []string   `json:"body,omitempty"`
}

type paramDTO struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
}

func toDTO(f *ast.File) fileDTO {
	dto := fileDTO{
		Settings:    map[string]any{},
		Assignments: map[string]any{},
	}
	for _, s := range f.Settings {
		dto.Settings[s.Name] = settingValue(s)
	}
	for _, a := range f.Assignments {
		dto.Assignments[a.Name] = exprText(a.Expr)
	}
	for _, t := range f.Tasks {
		td := taskDTO{
			Name:     t.Name,
			Doc:      t.Doc,
			Executor: t.Executor,
			Private:  t.IsPrivate(),
		}
		for _, p := range t.Params {
			td.Params = append(td.Params, paramDTO{Name: p.Name, Kind: paramKind(p.Kind)})
		}
		for _, d := range t.Deps {
			td.Deps = append(td.Deps, d.Name)
		}
		for _, d := range t.PostHooks {
			td.PostHooks = append(td.PostHooks, d.Name)
		}
		for _, a := range t.Attributes {
			td.Attributes = append(td.Attributes, a.Kind)
		}
		for _, bl := range t.Body {
			td.Body = append(td.Body, bl.Raw)
		}
		dto.Tasks = append(dto.Tasks, td)
	}
	return dto
}

func settingValue(s *ast.Setting) any {
	switch {
	case s.Bool:
		return true
	case s.List != nil:
		vals := make([]string, len(s.List))
		for i, e := range s.List {
			vals[i] = exprText(e)
		}
		return vals
	default:
		return exprText(s.Value)
	}
}

func paramKind(k ast.ParamKind) string {
	switch k {
	case ast.ParamRequired:
		return "required"
	case ast.ParamDefaulted:
		return "defaulted"
	case ast.ParamVariadicPlus:
		return "variadic+"
	case ast.ParamVariadicStar:
		return "variadic*"
	default:
		return "unknown"
	}
}

// exprText renders an expression's source-ish form via the shared dumper.
func exprText(e ast.Expr) string {
	if e == nil {
		return ""
	}
	return ast.DumpExpr(e)
}
