develop:
	go build

standalone:
	go-bindata-assetfs data/...
	go build -tags "bindata"

dep:
	go get -v github.com/jteeuwen/go-bindata/...
	go get -v github.com/elazarl/go-bindata-assetfs/...
