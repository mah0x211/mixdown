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
// Created by Masatoshi Fukunaga on 19/02/18
//

package util

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Basename strip directory and suffix from filenames
func Basename(pathname string) string {
	basename := filepath.Base(pathname)
	return basename[:len(basename)-len(filepath.Ext(basename))]
}

// IsDir returns a true if pathname is directory
func IsDir(pathname string) (bool, error) {
	if realpath, err := filepath.EvalSymlinks(pathname); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	} else if stat, err := os.Lstat(realpath); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	} else {
		return stat.IsDir(), nil
	}
}

// Mkdir create directory
func Mkdir(dirname string) error {
	if info, err := os.Lstat(dirname); err != nil {
		// got error
		if !os.IsNotExist(err) {
			return err
		} else if err = os.MkdirAll(dirname, 0766); err != nil {
			return err
		}
	} else if !info.IsDir() {
		return fmt.Errorf("%q already exists as a non-directory", dirname)
	}

	return nil
}

// CreateFile create file of pathname
func CreateFile(pathname string) (*os.File, error) {
	const flgs = os.O_WRONLY | os.O_CREATE | os.O_TRUNC

	if err := Mkdir(filepath.Dir(pathname)); err != nil {
		return nil, err
	} else if f, err := os.OpenFile(pathname, flgs, 0644); err != nil {
		return nil, err
	} else {
		return f, nil
	}
}

// CopyFile copy srcpath file to dstpath file
func CopyFile(srcpath, dstpath string) error {
	log.Printf("copy file %q to %q", srcpath, dstpath)
	if strings.HasPrefix(filepath.Base(srcpath), ".") {
		log.Printf("skip dotfile %q", srcpath)
		return nil
	}

	srcpath = filepath.Clean(srcpath)
	dstpath = filepath.Clean(dstpath)
	if err := os.MkdirAll(filepath.Dir(dstpath), 0766); err != nil {
		return fmt.Errorf("error os.MkdirAll(): %s", err)
	}

	if realpath, err := filepath.EvalSymlinks(srcpath); err != nil {
		return err
	} else {
		srcpath = realpath
	}

	src, err := os.Open(srcpath)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.OpenFile(dstpath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer dst.Close()

	buf := make([]byte, os.Getpagesize()*5)
	for {
		n, err := src.Read(buf)
		if err != nil && err != io.EOF {
			return err
		} else if n == 0 {
			break
		} else if _, err := dst.Write(buf[:n]); err != nil {
			return err
		}
	}

	return nil
}

// CopyDir copy srcdir directory to dstdir directory
func CopyDir(srcdir, dstdir string) error {
	var (
		ok  bool
		err error
		src *os.File
	)

	log.Printf("copy dir %q to %q", srcdir, dstdir)
	if strings.HasPrefix(filepath.Base(srcdir), ".") {
		log.Printf("skip dotfile %q", srcdir)
		return nil
	}

	if ok, err = IsDir(srcdir); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("srcdir must be directory")
	} else if err = Mkdir(dstdir); err != nil {
		return err
	}

	src, err = os.Open(srcdir)
	if err != nil {
		return err
	}
	defer src.Close()

	for {
		finfos, err := src.Readdir(1)
		if err != nil {
			// done
			if err == io.EOF {
				return nil
			}
			return err
		}

		srcname := filepath.Join(srcdir, finfos[0].Name())
		dstname := filepath.Join(dstdir, finfos[0].Name())
		if ok, err := IsDir(srcname); ok {
			if err = CopyDir(srcname, dstname); err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else if err = CopyFile(srcname, dstname); err != nil {
			if !os.IsNotExist(err) {
				return err
			}
			log.Printf("failed to copy file %q to %q - %s", srcname, dstname, err)
		}
	}
}

// ExecCommand execute a first argument as command with remaining arguments as
// command arguments.
func ExecCommand(args ...string) ([]byte, error) {
	cmd := exec.Command(args[0], args[1:]...)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error %q: %s", strings.Join(args, " "), err)
	}

	return bytes.TrimSpace(out), nil
}

// GenChecksum generate a SHA256 checksum of specified data
func GenChecksum(data []byte) string {
	chksum := sha256.Sum256(data)
	return hex.EncodeToString(chksum[:])
}
