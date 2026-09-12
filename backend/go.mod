module mediasequencer

go 1.22

require go.mongodb.org/mongo-driver v1.17.1

require (
	github.com/golang/snappy v0.0.4 // indirect
	github.com/klauspost/compress v1.13.6 // indirect
	github.com/montanaflynn/stats v0.7.1 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	golang.org/x/crypto v0.26.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/text v0.17.0 // indirect
)

replace go.mongodb.org/mongo-driver => github.com/mongodb/mongo-go-driver v1.17.1

replace golang.org/x/text => github.com/golang/text v0.14.0

replace golang.org/x/crypto => github.com/golang/crypto v0.26.0

replace golang.org/x/sync => github.com/golang/sync v0.8.0

replace golang.org/x/tools => github.com/golang/tools v0.6.0

replace golang.org/x/mod => github.com/golang/mod v0.8.0

replace golang.org/x/net => github.com/golang/net v0.25.0

replace golang.org/x/sys => github.com/golang/sys v0.23.0

replace golang.org/x/term => github.com/golang/term v0.23.0
