package defs

import (
	"fmt"

	"github.com/matthewmueller/golly/internal/dom/def"
	"github.com/matthewmueller/golly/internal/dom/index"
	"github.com/matthewmueller/golly/internal/dom/raw"
	"github.com/matthewmueller/golly/internal/gen"
	"github.com/pkg/errors"
)

var _ Dictionary = (*dict)(nil)

// NewDictionary create a callback
func NewDictionary(index index.Index, data *raw.Dictionary) Dictionary {
	return &dict{
		data:  data,
		index: index,
	}
}

// Dictionary interface
type Dictionary interface {
	def.Definition
}

// dict struct
type dict struct {
	data *raw.Dictionary
	pkg  string
	file string

	index index.Index
}

// ID fn
func (d *dict) ID() string {
	return d.data.Name
}

// Name fn
func (d *dict) Name() string {
	return d.data.Name
}

// Kind fn
func (d *dict) Kind() string {
	return "DICTIONARY"
}

func (d *dict) Type() (string, error) {
	return d.index.Coerce(d.data.Name)
}

func (d *dict) SetPackage(pkg string) {
	d.pkg = pkg
}
func (d *dict) GetPackage() string {
	return d.pkg
}

func (d *dict) SetFile(file string) {
	d.file = file
}
func (d *dict) GetFile() string {
	return d.file
}

// Parents fn
func (d *dict) Parents() (parents []def.Definition, err error) {
	if d.data.Extends != "" && d.data.Extends != "Object" {
		parent, isset := d.index[d.data.Extends]
		if !isset {
			return parents, fmt.Errorf("extends doesn't exist %s on %s", d.data.Extends, d.data.Name)
		}
		parents = append(parents, parent)
	}
	return parents, nil
}

// // Ancestors fn
// func (d *dict) Ancestors() []def.Definition {
// 	return nil
// }

// Children fn
func (d *dict) Dependencies() (defs []def.Definition, err error) {
	for _, member := range d.data.Members {
		if def := d.index.Find(member.Type); def != nil {
			defs = append(defs, def)
		}
	}
	return defs, nil
}

// Generate fn
func (d *dict) Generate() (string, error) {
	data := struct {
		Name    string
		Embeds  []string
		Members []string
	}{
		Name: gen.Capitalize(d.data.Name),
	}

	// Handle embeds
	parents, err := d.Parents()
	if err != nil {
		return "", errors.Wrapf(err, "error getting parents")
	}
	for _, parent := range parents {
		data.Embeds = append(data.Embeds, parent.Name())
	}

	for _, member := range d.data.Members {
		m, e := d.generateMember(member)
		if e != nil {
			return "", e
		}
		data.Members = append(data.Members, m)
	}

	return gen.Generate("dictionary/"+d.data.Name, data, `
		type {{ .Name }} struct {
			{{ range .Embeds }}
			{{ . }}
			{{ end }}

			{{ range .Members }}
			{{ . }}
			{{- end }}
		}
	`)
}

type memberData struct {
	Name string
	Type string
}

// Generate fn
func (d *dict) generateMember(m *raw.Member) (string, error) {
	member := gen.Vartype{
		Var: m.Name,
	}

	t, e := d.index.Coerce(m.Type)
	if e != nil {
		return "", e
	}
	member.Type = t

	// make the optional fields pointers
	if m.Nullable || !m.Required {
		member.Optional = true
	}

	return gen.Generate("member/"+m.Name, member, `{{ vt . }}`)
}
