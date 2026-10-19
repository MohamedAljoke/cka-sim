package docs

// Page is one page of the Kubernetes docs, cut into the chunks a search answers with. Its hash changes whenever
// its text does, so an unchanged page keeps the vectors it already has.
type Page struct {
	Path   string
	Hash   string
	Chunks []Chunk
}

type Chunk struct {
	Title string
	URL   string
	Text  string
}

// Index is the docs with a vector per chunk, from one embedding model.
type Index struct {
	Model string
	Ref   string
	Pages []IndexedPage
}

type IndexedPage struct {
	Path   string
	Hash   string
	Chunks []IndexedChunk
}

type IndexedChunk struct {
	Chunk
	Vector []float32
}
