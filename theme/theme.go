//
// Copyright (C) 2019 Masatoshi Fukunaga
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.  IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.
//
// Created by Masatoshi Fukunaga on 19/02/17
//

package theme

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/mah0x211/mixdown/rex"
	"github.com/mah0x211/mixdown/util"
)

// Theme is the representation of the manager of theme files and assets
type Theme struct {
	name   string
	assets map[string]string
	tmpls  map[string]*template.Template
}

func parseTemplate(tmpl *template.Template, src string) error {
	buf, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	} else if _, err = tmpl.Parse(string(buf)); err != nil {
		return err
	}

	// lookup nested-template actions
	dirname := filepath.Dir(src)
	matches := rex.TemplateAction.FindAllSubmatch(buf, -1)
	for _, match := range matches {
		// append nested-template
		if match[1] != nil {
			src = filepath.Join(dirname, string(match[1]))
			err = parseTemplate(tmpl, src)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// New allocate a instance of Theme
func New(themedir string) (*Theme, error) {
	tmpls := make(map[string]*template.Template)
	assets := make(map[string]string)

	// verify theme directory
	if ok, err := util.IsDir(themedir); err != nil {
		return nil, err
	} else if !ok {
		return nil, fmt.Errorf("theme %q is not found", themedir)
	}

	// load theme-files
	finfos, err := ioutil.ReadDir(themedir)
	if err != nil {
		return nil, err
	}

	for i, k := 0, len(finfos); i < k; i++ {
		fname := finfos[i].Name()

		if strings.HasPrefix(fname, ".") {
			// ignore dot-file
			continue
		} else if finfos[i].IsDir() {
			// holds asset dirinfo
			assets[fname] = filepath.Join(themedir, fname)
			log.Printf("asset %q", fname)
			continue
		}

		// ignore non-template-file
		if !rex.ThemeFile.MatchString(fname) {
			continue
		}

		// decompose filename
		matches := rex.ThemeFile.FindStringSubmatch(fname)
		basename, layout := matches[1], matches[2]
		log.Printf("template %q - %q", basename, fname)

		// parse template file
		src := filepath.Join(themedir, fname)
		tmpl := template.New(basename)
		tmpl.Funcs(defaultFuncMap)
		if err = parseTemplate(tmpl, src); err != nil {
			return nil, err
		}

		// decompose filename
		if layout != "" {
			err = parseTemplate(tmpl, filepath.Join(themedir, layout))
			if err != nil {
				return nil, err
			}
		}
		tmpls[basename] = tmpl
	}

	return &Theme{
		assets: assets,
		tmpls:  tmpls,
	}, nil
}

// Exists returns true if the specified named template exists
func (t *Theme) Exists(name string) bool {
	return t.tmpls[name] != nil
}

// Execute applies a parsed template to the specified data object,
// and writes the output to wr.
func (t *Theme) Execute(wr io.Writer, name string, data interface{}) error {
	if tmpl, ok := t.tmpls[name]; ok {
		return tmpl.Execute(wr, data)
	}
	return fmt.Errorf("template %q not found", name)
}

// ExportAssets copy asset directories into outdir
func (t *Theme) ExportAssets(outdir string) error {
	for name, srcdir := range t.assets {
		dstdir := filepath.Join(outdir, name)
		log.Printf("export %q %q -> %q", name, srcdir, dstdir)
		if err := util.CopyDir(srcdir, dstdir); err != nil {
			return err
		}
	}

	return nil
}
