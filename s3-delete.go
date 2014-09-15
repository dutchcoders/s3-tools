package main

// Made with <3 by DutchCoders (dutchcoders.io)

import (
	"flag"
	"fmt"
	"github.com/goamz/goamz/aws"
	"github.com/dutchcoders/goamz/s3"
	"log"
	"net/url"
	"path/filepath"
	"sync"
	"runtime"
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

	conn := s3.New(auth, aws.Region{})
	b := conn.Bucket(bucket)
	return b, nil
}

type Excludes []string

func (self *Excludes) String() string {
	return fmt.Sprint(*self)
}

func (self *Excludes) Set(value string) error {
	*self = append(*self, value)
	return nil
}

func (self *Excludes) Match(value string) (bool, error) {
	for _, exclude := range *self {
		matched, err := filepath.Match(exclude, value)

		if err != nil {
			return false, err
		}

		if matched == true {
			return true, nil
		}
	}

	return false, nil
}

func main() {
        maxProcs := runtime.NumCPU()
        runtime.GOMAXPROCS(maxProcs)

	fmt.Println("S3-delete: fast bulk delete.")
	fmt.Println("")
        fmt.Println("Made with <3 by DutchCoders (http://dutchcoders.io/)")
	fmt.Println("----------------------------------------------------")

	threads := flag.Int("threads", 20, "number of threads, default: 20")

	var excludes Excludes

	flag.Var(&excludes, "exclude", "exclude pattern, default: none")
	flag.Parse()

	if flag.Arg(0) == "" {
		fmt.Println("Error: Source not set")
		return
	}

	b, err := getBucketWithURL(flag.Arg(0))
	if err != nil {
		log.Panic(err.Error())
		return
	}

	sem := make(chan int, *threads)

        var wg sync.WaitGroup

	marker := ""

	for {
		result, err := b.List("", "", marker, 1000)
		if err != nil {
			log.Panic(err.Error())
			return
		}


		for _, key := range result.Contents {
			if match, err := excludes.Match(key.Key); match || err != nil {
				log.Print(err.Error())
				continue
			}

                        wg.Add(1)

			sem <- 1


			go func(source *s3.Bucket, path string) {

				log.Printf("Deleting %s", path)

				defer func() {
                                        wg.Done()
					<-sem
				}()

				err := source.Del(path)
				if err != nil {
					log.Print(err.Error())
					return
				}
			}(b, key.Key)
		}

		if !result.IsTruncated {
			break
		}

                if (result.NextMarker != "") {
                    marker = result.NextMarker
                }
	}

        wg.Wait()
}
