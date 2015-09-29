package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

func main() {
	fs, _ := ioutil.ReadDir("./templates/")
	out, _ := os.Create("templates.go")
	out.Write([]byte("package rest\n\nconst(\n"))
	for _, f := range fs {
		if strings.HasSuffix(f.Name(), ".html") {
			out.Write([]byte(strings.TrimSuffix(f.Name(), ".html") + " = `"))
			f, err := os.Open("./templates/" + f.Name())
			if err != nil {
				fmt.Println(err)
			}
			io.Copy(out, f)
			out.Write([]byte("`\n"))
		}
	}
	out.Write([]byte(")\n"))
}
