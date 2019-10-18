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
// Created by Masatoshi Fukunaga on 19/02/19
//

package file

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mah0x211/mixdown/rex"
	"github.com/mah0x211/mixdown/util"

	blackfriday "gopkg.in/russross/blackfriday.v2"
)

type mdExtractor struct {
	Subject     string
	Summary     string
	done        bool
	skipSubject bool
	container   *blackfriday.Node
	buf         bytes.Buffer
}

func (e *mdExtractor) Extract(node *blackfriday.Node, entering bool) (skip bool) {
	if e.done {
		return
	}

	// ignore document
	if node.Type == blackfriday.Document {
		return
	}

	// select container
	if e.container == nil {
		// not container node
		if !entering {
			return
		}

		// extract subject from h1 node
		if !e.skipSubject {
			if node.Type == blackfriday.Heading &&
				node.HeadingData.Level == 1 {
				e.container = node
				// skip rendering as body contents
				skip = true
				return
			}
			// skip subject extraction if non-h1 node
			e.skipSubject = true
		}

		// extract summary from p node
		if node.Type == blackfriday.Paragraph {
			e.container = node
		} else {
			// skip summary extraction if non-p node
			e.done = true
		}
		return
	}

	// end container walking
	if node == e.container {
		e.container = nil

		// use extracted literals as subject
		if !e.skipSubject {
			e.skipSubject = true
			e.Subject = e.buf.String()
			e.buf.Reset()
			// skip rendering as body contents
			skip = true
			return
		}

		// use extracted literals as summary
		e.done = true
		e.Summary = e.buf.String()
		return
	}

	// extract literal
	if node.Literal != nil {
		e.buf.Write(node.Literal)
	}

	// skip rendering as body contents
	if !e.skipSubject {
		skip = true
	}

	return
}

type TrackedFile struct {
	isMarkdown bool
	Href       string
	Pathname   string
	Source     string
	Name       string
	Author     string
	Cdate      string
	Ctime      string
	Mtime      string
	Subject    string
	Summary    string
	Hashtags   []string
	Content    string
	Newer      *TrackedFile
	Older      *TrackedFile
}

func epoch2iso8601(epoch string) (string, error) {
	if i64, err := strconv.ParseInt(epoch, 10, 64); err != nil {
		return "", err
	} else {
		const fmtISO8601 = "20060102T150405Z"
		return time.Unix(i64, 0).Format(fmtISO8601), nil
	}
}

// GetTrackedFiles ...
func GetTrackedFiles(baseURL string, useEpochname bool, extname string) ([]*TrackedFile, []*TrackedFile, error) {
	// read tracked files of git
	out, err := util.ExecCommand("git", "ls-files", "-z")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to util.ExecCommand(): %s", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\000")
	docs := make([]*TrackedFile, 0)
	rsrc := make([]*TrackedFile, 0)
	for _, src := range lines {
		// skip EOF, LICENSE\..* and dotfiles
		if src == "" || src == "LICENSE" || strings.HasPrefix(src, "LICENSE.") ||
			strings.HasPrefix(src, ".") {
			continue
		}

		// get last commit-log with following command;
		// 	git log -n 1 --format=%ae/%cd/%s/%b -- ${file}
		// 	  %ae: author email
		//    %ct: committer date, UNIX timestamp
		//    %s : subject
		//    %b : body
		// 	for more details: https://git-scm.com/docs/git-log
		out, err = util.ExecCommand("git", "log", "--format=%ae%x00%ct%x00%s%x00%b%x00", "--", src)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to util.ExecCommand(): %s", err)
		}

		// extract segments
		logs := strings.Split(string(out), "\000\n")
		log.Printf("%q - %q", src, logs[0])
		info := strings.Split(logs[0], "\000")
		f := &TrackedFile{
			isMarkdown: strings.HasSuffix(src, ".md"),
			Source:     src,
			Href:       src,
			Pathname:   src,
			Name:       util.Basename(src),
			Author:     strings.SplitN(info[0], "@", 2)[0], // without domain name
			Ctime:      info[1],
			Mtime:      info[1],
			Subject:    strings.TrimSpace(info[2]),
			Summary:    strings.TrimSpace(info[3]),
		}

		// set first-commit time to ctime
		if len(logs) > 1 {
			info = strings.SplitN(logs[len(logs)-1], "\000", 3)
			f.Ctime = info[1]
		}

		// convert ctime to cdate
		if f.Cdate, err = epoch2iso8601(f.Ctime); err != nil {
			return nil, nil, fmt.Errorf("failed to epoch2iso8601(): %s", err)
		}

		// preprocess
		if f.isMarkdown {
			// extract hashtags
			for _, match := range rex.Hashtag.FindAllStringIndex(f.Summary, -1) {
				f.Hashtags = append(f.Hashtags, f.Summary[match[0]:match[1]])
			}

			// create pathname
			if useEpochname {
				f.Pathname = filepath.Join(f.Cdate[:4], f.Ctime+"."+extname)
				f.Href = filepath.Join(baseURL, f.Pathname)
			} else if src == "README.md" {
				f.Pathname = f.Name + "." + extname
				f.Href = filepath.Join(baseURL, f.Pathname)
			} else {
				f.Pathname = filepath.Join(f.Cdate[:4], f.Name+"."+extname)
				f.Href = filepath.Join(
					baseURL, f.Cdate[:4], url.PathEscape(f.Name)+"."+extname,
				)
			}

			if err = f.Load(); err != nil {
				return nil, nil, fmt.Errorf("error File.Load(): %s", err)
			}

			docs = append(docs, f)
		} else {
			rsrc = append(rsrc, f)
		}
	}

	// sort by date in descending order
	sort.Slice(docs, func(i, j int) bool {
		return docs[i].Ctime > docs[j].Ctime
	})
	// set links
	for i, n := 0, len(docs)-1; i < n; i++ {
		docs[i].Older, docs[i+1].Newer = docs[i+1], docs[i]
	}

	return docs, rsrc, nil
}

// Load ...
func (f *TrackedFile) Load() error {
	// render markdown
	if f.isMarkdown {
		out, err := ioutil.ReadFile(f.Source)
		if err != nil {
			return fmt.Errorf("error ioutil.ReadFile(): %s", err)
		}
		out = bytes.TrimSpace(out)

		r := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
			Flags: blackfriday.CommonHTMLFlags,
		})
		parser := blackfriday.New(
			blackfriday.WithExtensions(blackfriday.CommonExtensions),
		)
		ast := parser.Parse(out)

		// extract subject and summary
		var buf bytes.Buffer
		extractor := &mdExtractor{}
		ast.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
			skipRender := extractor.Extract(node, entering)
			if !skipRender {
				return r.RenderNode(&buf, node, entering)
			}
			return blackfriday.GoToNext
		})

		if extractor.Subject != "" {
			f.Subject = extractor.Subject
		}
		if extractor.Summary != "" {
			f.Summary = extractor.Summary
		}
		f.Content = buf.String()
	}

	return nil
}

// Unload ...
func (f *TrackedFile) Unload() {
	f.Content = ""
}
