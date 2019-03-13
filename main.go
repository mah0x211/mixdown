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
// Created by Masatoshi Fukunaga on 19/01/09
//

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/mah0x211/mixdown/file"
	"github.com/mah0x211/mixdown/rex"
	"github.com/mah0x211/mixdown/theme"
	"github.com/mah0x211/mixdown/util"
)

type Mixdown struct {
	// configuration parameters
	OutDir       string `json:"outdir,omitempty"`
	UseEpochname bool   `json:"use_epochname,omitempty"`
	Extname      string `json:"extname,omitempty"`
	NArchive     int    `json:"narchive,omitempty"`

	ThemeDir  string              `json:"-"`
	Theme     *theme.Theme        `json:"-"`
	Hashtags  []string            `json:"-"`
	Documents []*file.TrackedFile `json:"-"`
	Resources []*file.TrackedFile `json:"-"`
	Readme    *file.TrackedFile   `json:"-"`
}

const MixdownDotDir string = ".mixdown/"

const (
	pageTypeHome = iota + 1
	pageTypeArticle
	pageTypeArchive
	pageTypeTag
)

func createMixdown() *Mixdown {
	return &Mixdown{
		OutDir:       "docs",
		UseEpochname: false,
		Extname:      "html",
		NArchive:     40,

		ThemeDir: filepath.Join(MixdownDotDir, "theme"),
	}
}

// render tags
func (m *Mixdown) renderTags() error {
	type stTag struct {
		PageType int
		Page     int
		NPage    *int
		Readme   *file.TrackedFile
		Hashtags []string
		Href     string
		Pathname string
		Subject  string
		Docs     []*file.TrackedFile
		Newer    *stTag
		Older    *stTag
	}

	// grouping files with hashtags
	tags := make(map[string]*stTag)
	ndoc := m.NArchive
	tagExists := make(map[string]bool)
	for _, doc := range m.Documents {
		// maintain references for readme.html
		if strings.HasPrefix(doc.Source, "README.") {
			m.Readme = doc
		}

		// grouping with hashtags
		for _, hashtag := range doc.Hashtags {
			tagName := hashtag[1:]
			href := filepath.Join("t", url.PathEscape(tagName)) + "/"
			doc.Summary = strings.Replace(
				doc.Summary, hashtag,
				fmt.Sprintf("<a href=%q>%s</a>", href, hashtag), 1,
			)

			// insert hashtag into list
			if _, ok := tagExists[hashtag]; !ok {
				tagExists[hashtag] = true
				m.Hashtags = append(m.Hashtags, tagName)
			}

			if tag, ok := tags[hashtag]; ok {
				// create next page
				if len(tag.Docs) == ndoc {
					*tag.NPage++
					pageName := strconv.Itoa(*tag.NPage) + "." + m.Extname
					tag.Older = &stTag{
						PageType: pageTypeTag,
						Page:     *tag.NPage,
						NPage:    tag.NPage,
						Href:     filepath.Join(href, pageName),
						Pathname: filepath.Join("t", tagName, pageName),
						Subject:  hashtag,
						Newer:    tag,
					}
					tag = tag.Older
				}
				tag.Docs = append(tag.Docs, doc)

			} else {
				page := 1
				tags[hashtag] = &stTag{
					PageType: pageTypeTag,
					Page:     page,
					NPage:    &page,
					Href:     filepath.Join(href, "index."+m.Extname),
					Pathname: filepath.Join("t", tagName, "index."+m.Extname),
					Subject:  hashtag,
					Docs:     []*file.TrackedFile{doc},
				}
			}
		}
	}

	// sort hashtags by name
	sort.Slice(m.Hashtags, func(i, j int) bool {
		return m.Hashtags[i] < m.Hashtags[j]
	})

	// render tags
	for tagName, tag := range tags {
		for tag != nil {
			tag.Readme = m.Readme
			tag.Hashtags = m.Hashtags
			pathname := filepath.Join(m.OutDir, tag.Pathname)
			log.Printf("%q -> %q", tagName, pathname)

			// render
			if ofile, err := util.CreateFile(pathname); err != nil {
				return fmt.Errorf("error util.CreateFile(): %s", err)
			} else if err = m.Theme.Execute(ofile, "tag", tag); err != nil {
				return fmt.Errorf("error Template.Execute(): %s", err)
			} else {
				ofile.Close()
			}
			tag = tag.Older
		}
	}

	return nil
}

// render articles
func (m *Mixdown) renderArticles() error {
	type stArticle struct {
		*file.TrackedFile
		PageType int
		Readme   *file.TrackedFile
		Hashtags []string
	}

	for _, doc := range m.Documents {
		if err := doc.Load(); err != nil {
			return fmt.Errorf("error File.Load(): %s", err)
		}

		article := stArticle{doc, pageTypeArticle, m.Readme, m.Hashtags}
		pathname := filepath.Join(m.OutDir, doc.Pathname)
		log.Printf("%q -> %q", doc.Source, pathname)

		// render
		if ofile, err := util.CreateFile(pathname); err != nil {
			return fmt.Errorf("error util.CreateFile(): %s", err)
		} else if err = m.Theme.Execute(ofile, "article", article); err != nil {
			return fmt.Errorf("error Template.Execute(): %s", err)
		} else {
			ofile.Close()
		}
		doc.Unload()
	}

	return nil
}

// render archives into archive/ directory
func (m *Mixdown) renderArchives() error {
	type stArchive struct {
		PageType int
		Page     int
		NPage    *int
		Readme   *file.TrackedFile
		Hashtags []string
		Href     string
		Pathname string
		Subject  string
		Docs     []*file.TrackedFile
		First    *file.TrackedFile
		Last     *file.TrackedFile
		Newer    *stArchive
		Older    *stArchive
	}

	// collect files to archive
	page := 1
	arc := &stArchive{
		PageType: pageTypeArchive,
		Page:     page,
		NPage:    &page,
		Pathname: filepath.Join("archive", "index."+m.Extname),
	}
	arc.Href = arc.Pathname
	head := arc
	ndoc := m.NArchive
	for _, doc := range m.Documents {
		// create next page
		if len(arc.Docs) == ndoc {
			arc.First, arc.Last = arc.Docs[0], arc.Docs[ndoc-1]
			page++
			arc.Older = &stArchive{
				PageType: pageTypeArchive,
				Page:     page,
				NPage:    &page,
				Pathname: filepath.Join("archive", strconv.Itoa(page)+"."+m.Extname),
				Newer:    arc,
			}
			arc = arc.Older
			arc.Href = arc.Pathname
		}
		arc.Docs = append(arc.Docs, doc)
	}

	if len(arc.Docs) > 0 {
		arc.First, arc.Last = arc.Docs[0], arc.Docs[len(arc.Docs)-1]
		// render archives
		arc = head
		for arc != nil {
			arc.Readme = m.Readme
			arc.Hashtags = m.Hashtags
			pathname := filepath.Join(m.OutDir, arc.Pathname)
			log.Printf("%q -> %q", arc.Pathname, pathname)

			// render
			if ofile, err := util.CreateFile(pathname); err != nil {
				return fmt.Errorf("error util.CreateFile(): %s", err)
			} else if err = m.Theme.Execute(ofile, "archive", arc); err != nil {
				return fmt.Errorf("error Template.Execute(): %s", err)
			} else {
				ofile.Close()
			}
			arc = arc.Older
		}
	}

	return nil
}

// render home
func (m *Mixdown) renderHome() error {
	type stHome struct {
		PageType int
		Readme   *file.TrackedFile
		Hashtags []string
		Subject  string
		Docs     []*file.TrackedFile
	}

	home := stHome{pageTypeHome, m.Readme, m.Hashtags, "", m.Documents}
	pathname := filepath.Join(m.OutDir, "index."+m.Extname)
	log.Printf("index -> %q", pathname)

	// render
	if ofile, err := util.CreateFile(pathname); err != nil {
		return fmt.Errorf("error util.CreateFile(): %s", err)
	} else if err = m.Theme.Execute(ofile, "home", home); err != nil {
		return fmt.Errorf("error Template.Execute(): %s", err)
	} else {
		ofile.Close()
	}

	return nil
}

// render resource
func (m *Mixdown) renderResources() error {
	for _, rsrc := range m.Resources {
		pathname := filepath.Join(m.OutDir, rsrc.Pathname)
		if err := util.CopyFile(rsrc.Pathname, pathname); err != nil {
			return fmt.Errorf("error util.CopyFile(): %s", err)
		}
		rsrc.Unload()
	}

	return nil
}

// render
func (m *Mixdown) render(target string) error {
	switch target {
	case "tag":
		return m.renderTags()
	case "article":
		return m.renderArticles()
	case "archive":
		return m.renderArchives()
	case "home":
		return m.renderHome()
	case "resources":
		return m.renderResources()
	default:
		return fmt.Errorf("unknown target %q", target)
	}
}

func main() {
	m := createMixdown()

	// load mixdown config file
	cfgFile := filepath.Join(MixdownDotDir, "config.json")
	if f, err := os.Open(cfgFile); err == nil {
		log.Println(strings.Repeat("*", 80))
		log.Printf("LOAD CONFIG FILE %q", cfgFile)
		if err = json.NewDecoder(f).Decode(m); err != nil {
			log.Fatalf("failed to load %q: %s", cfgFile, err)
		}
		f.Close()

		// verify options
		if strings.HasPrefix(m.OutDir, MixdownDotDir) {
			log.Fatalf("error invalid outdir configuration %q - cannot be output to the %q directory", m.OutDir, MixdownDotDir)
		} else if !rex.Extname.MatchString(m.Extname) {
			log.Fatalf("error invalid extname configuration %q - extname must be [0-9a-zA-Z_]+", m.Extname)
		} else if m.NArchive < 1 {
			log.Fatalf("error invalid narchive configuration %d - narchive must be greater than 0", m.NArchive)
		}
		log.Println(strings.Repeat("*", 80))

	} else if !os.IsNotExist(err) {
		log.Fatalf("failed to load %q: %s", cfgFile, err)
	}

	// parse command-line parameters
	flag.StringVar(&m.OutDir, "outdir", m.OutDir, "pathname of output directory. if not specified, automatically generate a temporary name.")
	flag.BoolVar(&m.UseEpochname, "use-epochname", m.UseEpochname, "use epoch time of file creation time as filename. (default \"false\")")
	flag.StringVar(&m.Extname, "extname", m.Extname, "extension name of the output file.")
	flag.IntVar(&m.NArchive, "narchive", m.NArchive, "number of articles in archive.")
	flag.Parse()

	// verify outdir
	if m.OutDir == "" {
		// generate temporary name
		if tmpdir, err := ioutil.TempDir("./", "mixdown-"); err != nil {
			log.Fatalf("failed to ioutil.TempDir(): %s", err)
		} else {
			m.OutDir = tmpdir
		}
	}

	// verify options
	if strings.HasPrefix(m.OutDir, MixdownDotDir) {
		log.Fatalf("error invalid outdir %q - cannot be output to the %q directory", m.OutDir, MixdownDotDir)
	} else if !rex.Extname.MatchString(m.Extname) {
		log.Fatalf("error invalid extname %q - extname must be [0-9a-zA-Z_]+", m.Extname)
	} else if m.NArchive < 1 {
		log.Fatalf("error invalid narchive %d - narchive must be greater than 0", m.NArchive)
	}

	log.Println("mixdown with following options;")
	log.Printf("  -outdir    : %q", m.OutDir)
	log.Printf("  -use-epochname: %t", m.UseEpochname)
	log.Printf("  -extname      : %q", m.Extname)
	log.Printf("  -narchive     : %d", m.NArchive)

	// remove existing output-dir
	if err := os.RemoveAll(m.OutDir); err != nil && !os.IsNotExist(err) {
		log.Fatalf("failed to os.RemoveAll(): %s", err)
	}

	// create outdir
	log.Println(strings.Repeat("*", 80))
	log.Printf("CREATE OUTPUT DIRECTORY %q", m.OutDir)
	if err := util.Mkdir(m.OutDir); err != nil {
		log.Fatalf("failed to util.Mkdir(): %s", err)
	}

	// load theme files
	log.Println(strings.Repeat("*", 80))
	log.Println("LOAD THEME FILES")
	if t, err := theme.New(m.ThemeDir); err != nil {
		log.Fatalf("failed to theme.New(): %s", err)
	} else {
		m.Theme = t
	}

	// load tracked files
	log.Println(strings.Repeat("*", 80))
	log.Println("LOAD TRACKED FILES")
	if docs, rsrc, err := file.GetTrackedFiles(m.UseEpochname, m.Extname); err != nil {
		log.Fatalf("failed to file.GetTrackedFiles(): %s", err)
	} else {
		m.Documents = docs
		m.Resources = rsrc
	}

	// render
	for _, target := range []string{
		"tag", "article", "archive", "home", "resources",
	} {
		log.Println(strings.Repeat("*", 80))
		log.Printf("RENDER %q", strings.Title(target))
		if err := m.render(target); err != nil {
			log.Fatalf("failed to render(): %s", err)
		}
	}

	// export assets
	log.Println(strings.Repeat("*", 80))
	log.Println("EXPORT ASSETS DIRECTORIES")
	if err := m.Theme.ExportAssets(m.OutDir); err != nil {
		log.Fatalf("failed to theme.ExportAssets(): %s", err)
	}

	log.Println(strings.Repeat("*", 80))
	log.Println("goodbye")
}
