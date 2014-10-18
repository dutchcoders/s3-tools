s3-tools
=======

Collection of multiple s3 tools

s3-copy
=======

s3-copy will copy the content of Amazon S3 buckets from one account to the same or another account. 

```
go get github.com/goamz/goamz/aws
go get github.com/goamz/goamz/s3

go run s3-copy.go s3://ACCESS_KEY:SECRET_KEY@source s3://ACCESS_KEY:SECRET_KEY@destination
```

s3-delete
=======

s3-delete will bulk delete the contents of an Amazon S3 bucket

```
go get github.com/goamz/goamz/aws
go get github.com/goamz/goamz/s3

go run s3-delete.go s3://ACCESS_KEY:SECRET_KEY@source
```

s3-archive
=======

s3-archive will archive files older then 1 month to Amazon S3 bucket with optional move

```
go get github.com/goamz/goamz/aws
go get github.com/goamz/goamz/s3

export AWS_ACCESS_KEY={AWS_ACCESS_KEY}
export AWS_SECRET_KEY={AWS_SECRET_KEY}
export BUCKET={BUCKET}

go run s3-archive --destination=dest --move=false /tmp/
```


If your password contains special characters, you need to url encode them. For example / will be %2F.

TODO
====

* support other regions
* support for paths
* support for wildcards
* add progress
* add statistics (upload speed)
* etc

CONTRIBUTIONS
============

Contributions are welcome.
