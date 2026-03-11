package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmhobbs/go-pbo"
)

func main() {
	f, err := os.Open("ViralSuppresor.pbo")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	bank, err := pbo.Load(f)
	if err != nil {
		panic(err)
	}

	for _, file := range bank.Files {
		var dir string
		segments := strings.Split(file.Filename, "\\")
		filename := segments[len(segments)-1]
		if len(segments) == 1 {
			dir = "."
		} else {
			dir = filepath.Join(segments[:len(segments)-1]...)
		}

		log.Println(file.Filename, dir, filename)

		err := os.MkdirAll(filepath.Join("output", dir), 0755)
		if err != nil {
			panic(err)
		}
		f, err := os.Create(filepath.Join("output", dir, filename))
		if err != nil {
			panic(err)
		}
		defer f.Close()

		reader, err := file.Reader()
		if err != nil {
			panic("Reader(): " + err.Error())
		}

		_, err = io.Copy(f, reader)
		if err != nil {
			panic("Copy(): " + err.Error())
		}
	}
}
