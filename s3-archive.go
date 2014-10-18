/*
The MIT License (MIT)

Copyright (c) 2014 DutchCoders [https://github.com/dutchcoders/]

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package main

import (
	"crypto/md5"
	_ "encoding/hex"
        "hash"
        "encoding/hex"
	"flag"
	"github.com/goamz/goamz/aws"
        "github.com/dutchcoders/goamz/s3"
        "errors"
	"io"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var config struct {
	AWS_ACCESS_KEY string
	AWS_SECRET_KEY string
	BUCKET         string
}

func init() {
	config.AWS_ACCESS_KEY = os.Getenv("AWS_ACCESS_KEY")
	config.AWS_SECRET_KEY = os.Getenv("AWS_SECRET_KEY")
	config.BUCKET = os.Getenv("BUCKET")
}

func getBucket() (*s3.Bucket, error) {
	auth, err := aws.GetAuth(config.AWS_ACCESS_KEY, config.AWS_SECRET_KEY, "", time.Time{})
	if err != nil {
		return nil, err
	}

	conn := s3.New(auth, aws.Regions["eu-west-1"])

	b := conn.Bucket(config.BUCKET)
	return b, nil
}

type md5reader struct {
	io.Reader
        hasher hash.Hash 
}

func (r *md5reader)Read(p []byte) (n int, err error) {
    n, err = r.Reader.Read(p)
    r.hasher.Write(p[:n])
    return
}

func (r *md5reader) Hash() string {
    return (hex.EncodeToString(r.hasher.Sum(nil)))
}


func MD5Reader(reader io.Reader) * md5reader {
	hasher := md5.New()
        return &md5reader{reader, hasher}
}

func walk(path string, f os.FileInfo, err error) error {
	if f.IsDir() {
		return nil
	}

	baseTime := time.Now()

	date := baseTime.AddDate(0, -6, 0)

	if !(f.ModTime().Before(date)) {
		return nil
	}

	// var err error

	// upload
        file, err := os.Open(path)

	defer file.Close()

	contentLength := f.Size()

	contentType := mime.TypeByExtension(filepath.Ext(path))

	if contentType == "" {
		contentType = "application/binary"
	}

	var relpath string
	if relpath, err = filepath.Rel(basepath, path); err != nil {
		return err
	}

	key := destination

	for _, p := range strings.Split(relpath, string(filepath.Separator)) {
		key = filepath.Join(key, p)
	}

	key = filepath.ToSlash(key)

	log.Printf("Uploading %s -> %s %d %s", path, key, contentLength, contentType)

	var b *s3.Bucket

	b, err = getBucket()
	if err != nil {
		log.Print(err.Error())
		return err
	}

        reader := MD5Reader(file)
        resp, err := b.PutReaderWithResponse(key, reader, contentLength, contentType, s3.Private, s3.Options{})

	if err != nil {
		log.Print(err.Error())
		return err
	}

        if (resp.Header["Etag"][0][1:33] != reader.Hash()) {
            return errors.New("Invalid hash")
        }

	if !move {
		return nil
	}

	file.Close()

	log.Printf("Deleting %s\n", path)
	defer os.Remove(path)

	return nil
}

var basepath string
var destination string
var move bool = false

func main() {
    	fmt.Println("S3-archive: archive folders to S3.")
	fmt.Println("")
        fmt.Println("Usage:")
        fmt.Println("go run s3-archive.go --destination=data --move=false /tmp/")
	fmt.Println("")
        fmt.Println("Made with <3 by DutchCoders (http://dutchcoders.io/)")
	fmt.Println("----------------------------------------------------")

        basepath = flag.Arg(0)

	if basepath == "" {
		log.Panic("Path not set")
	}

	flag.StringVar(&destination, "destination", "", "")
	flag.BoolVar(&move, "move", false, "")
	flag.Parse()

	filepath.Walk(basepath, walk)
}
