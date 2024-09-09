module github.com/foomo/bluge-sled

go 1.22.2

replace (
	github.com/blugelabs/bluge => github.com/zincsearch/bluge v1.1.5
	github.com/blugelabs/bluge_segment_api => github.com/zincsearch/bluge_segment_api v1.0.0
	github.com/blugelabs/ice => github.com/zincsearch/ice v1.1.3
)

require (
	github.com/a-h/templ v0.2.747
	github.com/blugelabs/bluge v0.1.9
	github.com/brianvoe/gofakeit/v7 v7.0.2
	github.com/cespare/xxhash v1.1.0
	github.com/google/uuid v1.3.1
	github.com/mozillazg/go-unidecode v0.2.0
	github.com/pkg/errors v0.9.1
	github.com/samber/lo v1.39.0
	github.com/stretchr/testify v1.8.4
	golang.org/x/sync v0.3.0
)

require (
	github.com/RoaringBitmap/roaring v0.9.4 // indirect
	github.com/axiomhq/hyperloglog v0.0.0-20230201085229-3ddf4bad03dc // indirect
	github.com/bits-and-blooms/bitset v1.2.2 // indirect
	github.com/blevesearch/go-porterstemmer v1.0.3 // indirect
	github.com/blevesearch/mmap-go v1.0.4 // indirect
	github.com/blevesearch/segment v0.9.1 // indirect
	github.com/blevesearch/snowballstem v0.9.0 // indirect
	github.com/blevesearch/vellum v1.0.10 // indirect
	github.com/blugelabs/bluge_segment_api v0.2.0 // indirect
	github.com/blugelabs/ice v1.0.0 // indirect
	github.com/caio/go-tdigest v3.1.0+incompatible // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-metro v0.0.0-20211217172704-adc40b04c140 // indirect
	github.com/klauspost/compress v1.17.3 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/mschoch/smat v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	golang.org/x/exp v0.0.0-20240222234643-814bf88cf225 // indirect
	golang.org/x/sys v0.21.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
