s3-copy
=======

s3-copy will copy then contents Amazon s3 buckets from one account to the same or another account. 

```
go get github.com/goamz/goamz/aws
go get github.com/goamz/goamz/s3

go run copy.go s3://ACCESS_KEY:SECRET_KEY@source s3://ACCESS_KEY:SECRET_KEY@destination
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
