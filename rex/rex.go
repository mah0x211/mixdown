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
// Created by Masatoshi Fukunaga on 19/03/07
//

package rex

import (
	"regexp"
)

var (
	// Extname is pattern of extension-name
	Extname = regexp.MustCompile(`^\w+`)

	// Hashtag is pattern of hashtags
	Hashtag = regexp.MustCompile(
		// single white space character pattern:
		// 	[^ \f\n\r\t\v\u00a0\u1680\u2000-\u200a\u2028\u2029\u202f\u205f\u3000\ufeff]
		//
		// above pattern defined at the following website:
		// 	https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/RegExp
		`\B#[^ \f\n\r\t\v` + "\u00a0\u1680\u2000-\u200a\u2028\u2029\u202f\u205f\u3000\ufeff" + `]+`,
	)

	// ThemeFile is pattern of names of theme-file
	ThemeFile = regexp.MustCompile(
		// <fname>.page.<ext>[@<fname>.<ext>]
		`^(\w+(?:\.\w+)*)\.mix\.\w+(?:@(\w+(?:\.\w+)+))?$`,
	)

	// TemplateAction is pattern of sub-template directive
	TemplateAction = regexp.MustCompile(
		// {{template "@name" .}}
		`\{{2}\s*template\s+"@(\w+(?:\.\w+)+?)"[^}]+}{2}`,
	)
)
