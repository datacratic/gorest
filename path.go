// Copyright (c) 2014 Datacratic. All rights reserved.

package rest

import (
	"bytes"
	"strings"
)

// PathItem represents a single item in a path which can either be an argument
// or a constant. The name of an argument is used purely for documentation
// purposes while the name of a constant is its value.
type PathItem struct {
	Name  string
	IsArg bool
}

// String returns the string representation of the item.
func (item PathItem) String() string {
	if item.IsArg {
		return ":" + item.Name
	}
	return item.Name
}

// Path is an array of PathItem which represents the templated path of an HTTP
// query.
//
// The format of a path is a series of item seperated by '/' characters. An item
// can either be a constant or an argument which starts with leading ':'
// character. A full path examples looks like the following:
//
//    /a/:b/c
//
// Where a and c are both constants and b is an argument.
type Path []PathItem

// SplitPath breaks a REST path into its components.
func SplitPath(path string) (split []string) {
	if trimmed := strings.Trim(path, "/"); len(trimmed) > 0 {
		split = strings.Split(trimmed, "/")
	}
	return
}

// JoinPath joins two partial paths into a single path.
func JoinPath(a, b string) string {
	return strings.TrimRight(a, "/") + "/" + strings.TrimLeft(b, "/")
}

// NewPath breaks up the given path into PathItem to form a new Path object. It
// panics if it's unable to parse the path.
func NewPath(rawPath string) (path Path) {
	for _, item := range SplitPath(rawPath) {
		if item[0] != ':' {
			path = append(path, PathItem{item, false})
			continue
		}

		path = append(path, PathItem{item[1:], true})
	}

	return
}

// NumArgs returns the number of items in the path where IsArg is true.
func (path Path) NumArgs() (n int) {
	for _, item := range path {
		if item.IsArg {
			n++
		}
	}
	return
}

// String returns the string representation of the path.
func (path Path) String() string {
	buffer := new(bytes.Buffer)
	buffer.WriteString("/")

	for _, item := range path {
		buffer.WriteString(item.String())
		buffer.WriteString("/")
	}

	return buffer.String()
}
