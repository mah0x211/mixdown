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
// Created by Masatoshi Fukunaga on 19/02/28
//

package theme

import (
	"fmt"
	"net/url"
	"reflect"
	"text/template"
)

func indirectInterface(v reflect.Value) reflect.Value {
	if v.Kind() != reflect.Interface {
		return v
	}
	if v.IsNil() {
		return reflect.Value{}
	}
	return v.Elem()
}

func fnIndirect(arg reflect.Value) reflect.Value {
	return reflect.Indirect(arg)
}

func fnSlice(arg reflect.Value, s, e int) (reflect.Value, error) {
	v := indirectInterface(arg)
	if !v.IsValid() {
		return reflect.Value{}, fmt.Errorf("index of untyped nil")
	}

	switch v.Kind() {
	case reflect.Array, reflect.Slice, reflect.String:
		n := v.Len()
		if s < 0 {
			s = 0
		} else if s > n {
			s = n
		}

		if e < 0 {
			e = n
		} else if e > n {
			e = n
		}

		return v.Slice(s, e), nil

	default:
		return reflect.Value{}, fmt.Errorf("can't index item of type %s", v.Type())
	}
}

func fnEscapePath(str string) string {
	return url.PathEscape(str)
}

var defaultFuncMap = template.FuncMap{
	"indirect":   fnIndirect,
	"slice":      fnSlice,
	"escapePath": fnEscapePath,
}
