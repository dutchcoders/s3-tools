package main

// Made with <3 by DutchCoders (dutchcoders.io)

import (
	"errors"
	"flag"
	"github.com/goamz/goamz/aws"
	"github.com/goamz/goamz/s3"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// https://bugs.launchpad.net/goamz/+bug/1087338/
// https://launchpadlibrarian.net/133464855/s3-keys-with-spaces.diff


func getBucketWithURL(urlStr string) (*s3.Bucket, error) {
	url, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	password, isset := url.User.Password()
	_ = isset

	return getBucket(url.Host, url.User.Username(), password)
}

func getBucket(bucket string, access_key string, secret_key string) (*s3.Bucket, error) {
	auth, err := aws.GetAuth(access_key, secret_key, "", time.Time{})
	if err != nil {
		return nil, err
	}

	conn := s3.New(auth, aws.Regions["eu-west-1"])
	b := conn.Bucket(bucket)
	return b, nil
}

func main() {
        threads := flag.Int("threads", "4", "number of threads, default: 4")
	flag.Parse()

	b, err := getBucketWithURL(flag.Arg(0))
	if err != nil {
		log.Panic(err.Error())
		return
	}

	destination, err := getBucketWithURL(flag.Arg(1))
	if err != nil {
		log.Panic(err.Error())
		return
	}

	sem := make(chan int, threads)

	marker := ""

	for {
		result, err := b.List("", "", marker, 20)
		if err != nil {
			log.Panic(err.Error())
			return
		}

		for _, key := range result.Contents {
			marker = key.Key

			sem <- 1

			go func(source *s3.Bucket, destination *s3.Bucket, path string) {

				log.Printf("Copying %s", path)

				defer func() {
					<-sem
				}()

				rc, err := source.GetResponse(path)
				if err != nil {
					log.Print(err.Error())
					return
				}

				defer rc.Body.Close()

				contentType := rc.Header.Get("Content-Type")
				eTag := rc.Header.Get("Etag")
				contentLength, err := strconv.ParseInt(rc.Header.Get("Content-Length"), 10, 0)

				options := s3.Options{}

				if err = destination.PutReader(path, rc.Body, contentLength, contentType, s3.PublicRead, options); err != nil {
					log.Print(err)
					return
				}

				var response *http.Response

				if response, err = destination.Head(path, nil); err != nil {
					log.Print(err)
					return
				}

				if response.Header.Get("Etag") != eTag {
					log.Print(errors.New("Signature doesnt match").Error())
                                        return 
				}

			}(b, destination, key.Key)
		}

		if !result.IsTruncated {
			break
		}
	}
}
