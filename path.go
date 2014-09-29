// Copyright (c) 2014 Datacratic. All rights reserved.

package rest

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// PathItem represents a single item in a path which can either be a constant or
// a positional argument. The name of a positional arguments is used for
// documentation and the position is used to determine which argument of the
// handler it represents.
type PathItem struct {
	Name string
	Pos  int
}

// IsPositional returns true if the item is a positional argument.
func (item PathItem) IsPositional() bool {
	return item.Pos >= 0
}

// String returns the string representation of the item.
func (item PathItem) String() string {
	if item.IsPositional() {
		return fmt.Sprintf("{%d:%s}", item.Pos, item.Name)
	}
	return item.Name
}

// Path is an array of PathItem which represents the templated path of an HTTP
// query.
//
// The format of a path is a series of item seperated by '/' characters. An item
// can either a string literal or a positional argument which takes the
// following form:
//
//     {%d:%s}
//
// Where the integral component is the position and the string argument is the
// description.
//
// A full path examples looks like the following:
//
//    /a/{0:b}/c
//
type Path []PathItem

func splitPath(path string) (split []string) {
	if trimmed := strings.Trim(path, "/"); len(trimmed) > 0 {
		split = strings.Split(trimmed, "/")
	}
	return
}

func joinPath(a, b string) string {
	return strings.TrimRight(a, "/") + "/" + strings.TrimLeft(b, "/")
}

// NewPath breaks up the given path into PathItem to form a new Path object. It
// panics if it's unable to parse the path.
func NewPath(rawPath string) (path Path) {
	for _, item := range splitPath(rawPath) {
		if item[0] != '{' {
			if n := strings.IndexAny(item, "{}"); n >= 0 {
				panic(fmt.Sprintf("invalid '{' or '}' char detected path '%s': %s", path, item))
			}

			path = append(path, PathItem{item, -1})
			continue

		}

		item = item[1 : len(item)-1]

		n := strings.Index(item, ":")
		if n < 0 {
			n = len(item)
		}

		pos, err := strconv.ParseUint(item[0:n], 10, 8)
		if err != nil {
			panic(fmt.Sprintf("unable to parse position: '%s' -> '%s'", item[0:n], err))
		}

		var name string
		if n != len(item) {
			name = item[n+1:]
		}

		path = append(path, PathItem{name, int(pos)})
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
