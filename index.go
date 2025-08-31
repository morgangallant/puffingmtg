package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/turbopuffer/turbopuffer-go"
)

// Index is a metadata object, serialized to JSON, which describes a built index.
type Index struct {
	// Name is the name of the index, as provided by the user.
	Name string `json:"name"`

	// Namespace is the name of the associated turbopuffer namespace for the index.
	Namespace string `json:"namespace"`

	// CreatedAt is the timestamp of when the index was created.
	CreatedAt time.Time `json:"created_at"`

	// Checksum is the hex SHA256 of the source file used to build the index. Used to detect changes
	// to the source file to know if a rebuild is necessary.
	Checksum string `json:"checksum"`

	// The set that was indexed.
	Set Set `json:"set"`
}

// LoadIndex loads an Index with the given name. If the index doesn't exist, returns nil.
// The index is expected to be in a file named <name>.json in the current directory.
func LoadIndex(name string) (*Index, error) {
	fp := indexFilepath(name)
	f, err := os.Open(fp)
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("opening index file %q: %w", fp, err)
	}
	defer f.Close()

	var index Index
	if err := json.NewDecoder(f).Decode(&index); err != nil {
		return nil, fmt.Errorf("decoding index file %q: %w", fp, err)
	}

	if index.Name != name {
		return nil, fmt.Errorf("index name mismatch: expected %q, got %q", name, index.Name)
	}

	return &index, nil
}

// Delete deletes the index both from turbopuffer and from local disk.
func (idx *Index) Delete(ctx context.Context, tpuf *turbopuffer.Client) error {
	ns := tpuf.Namespace(idx.Namespace)

	var tpufError *turbopuffer.Error
	if _, err := ns.DeleteAll(ctx, turbopuffer.NamespaceDeleteAllParams{}); err != nil {
		if errors.As(err, &tpufError) && tpufError.StatusCode == http.StatusNotFound {
			return nil
		}
		return fmt.Errorf("deleting all rows in namespace %q: %w", idx.Namespace, err)
	}

	fp := indexFilepath(idx.Name)
	if err := os.Remove(fp); err != nil {
		return fmt.Errorf("deleting index file %q: %w", fp, err)
	}

	return nil
}

// NewIndex creates a new Index file with a given name, indexing a particular set.
// If the index file already exists, returns an error.
func NewIndex(ctx context.Context, tpuf *turbopuffer.Client, name string, set Set) (*Index, error) {
	fp := indexFilepath(name)
	f, err := os.OpenFile(fp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if os.IsExist(err) {
		return nil, fmt.Errorf("index file %q already exists", fp)
	} else if err != nil {
		return nil, fmt.Errorf("creating index file %q: %w", fp, err)
	}
	defer f.Close()

	start := time.Now()
	log.Printf("downloading set %q from mtgjson...", set)

	setObj, checksum, err := downloadSet(ctx, set)
	if err != nil {
		return nil, fmt.Errorf("downloading set %q: %w", set, err)
	}
	log.Printf("downloaded set %q in %s", set, time.Since(start))
	log.Printf("computed checksum: %s", checksum)

	var (
		nsName = turbopufferNamespace(name, checksum)
		ns     = tpuf.Namespace(nsName)
	)
	if err := ensureNamespaceDoesntExist(ctx, ns); err != nil {
		return nil, fmt.Errorf("ensuring namespace %q doesn't exist: %w", nsName, err)
	}
	log.Printf("using turbopuffer namespace %q", nsName)

	if err := upsertSet(ctx, ns, setObj); err != nil {
		return nil, fmt.Errorf("uploading set to turbopuffer: %w", err)
	}
	log.Printf("uploaded set to turbopuffer namespace %q", nsName)

	index := &Index{
		Name:      name,
		Namespace: nsName,
		CreatedAt: time.Now().UTC(),
		Checksum:  checksum,
		Set:       set,
	}
	if err := json.NewEncoder(f).Encode(index); err != nil {
		return nil, fmt.Errorf("writing index file %q: %w", fp, err)
	}
	log.Printf("wrote index file %q", fp)

	return index, nil
}

func indexFilepath(name string) string {
	return fmt.Sprintf("%s.json", name)
}

func downloadSet(ctx context.Context, set Set) (*AtomicSet, string, error) {
	url, err := set.DownloadURL()
	if err != nil {
		return nil, "", fmt.Errorf("getting download URL for set %q: %w", set, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("creating request for %q: %w", url, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("downloading set from %q: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf(
			"downloading set from %q: unexpected status code %d (status text: %s)",
			url,
			resp.StatusCode,
			resp.Status,
		)
	}

	// As we're reading the response body for JSON deserialization, we'll compute a rolling
	// checksum of the underlying data. We'll store this in the index object.
	var (
		hasher = sha256.New()
		tee    = io.TeeReader(resp.Body, hasher)
	)

	var atomicSet AtomicSet
	if err := json.NewDecoder(tee).Decode(&atomicSet); err != nil {
		return nil, "", fmt.Errorf("decoding set from %q: %w", url, err)
	}

	checksum := hex.EncodeToString(hasher.Sum(nil))

	return &atomicSet, checksum, nil
}

func turbopufferNamespace(name, checksum string) string {
	now := time.Now().UTC().Format("20060102-150405")
	return fmt.Sprintf("mtg_%s_%s_%s", name, now, checksum[:8])
}

func ensureNamespaceDoesntExist(ctx context.Context, ns turbopuffer.Namespace) error {
	var tpufError *turbopuffer.Error
	meta, err := ns.Metadata(ctx, turbopuffer.NamespaceMetadataParams{})
	if err != nil {
		if errors.As(err, &tpufError) && tpufError.StatusCode == http.StatusNotFound {
			return nil
		}
		return fmt.Errorf("checking namespace metadata: %w", err)
	}
	return fmt.Errorf("namespace %q already exists (created at %s)", ns.ID(), meta.CreatedAt)
}

func upsertSet(ctx context.Context, ns turbopuffer.Namespace, set *AtomicSet) error {
	const (
		targetBatchSize  = 128 << 20 // 128MB
		estimatedRowSize = 1 << 10   // 1KB
	)
	batch := make([]turbopuffer.RowParam, 0, targetBatchSize/estimatedRowSize)
	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		if _, err := ns.Write(ctx, turbopuffer.NamespaceWriteParams{
			UpsertRows: batch,
			Schema:     turbopufferSchema(),
		}); err != nil {
			return fmt.Errorf("writing batch of %d rows: %w", len(batch), err)
		}
		batch = batch[:0]
		return nil
	}
	var (
		numCards   int
		numFlushes int
	)
	for _, cards := range set.Data {
		for _, card := range cards {
			batch = append(batch, buildRow(card))
			numCards += 1
			if len(batch)*estimatedRowSize >= targetBatchSize {
				if err := flush(); err != nil {
					return err
				}
				numFlushes += 1
			}
		}
	}
	if err := flush(); err != nil {
		return err
	}
	numFlushes += 1

	log.Printf("uploaded %d cards (%d flushes)", numCards, numFlushes)

	return nil
}

func buildRow(card AtomicCard) turbopuffer.RowParam {
	return turbopuffer.RowParam{
		"id":                  uuid.NewString(),
		"types":               card.Types,
		"power":               card.Power,
		"toughness":           card.Toughness,
		"name":                card.Name,
		"edhrec_rank":         card.EdhrecRank,
		"edhrec_saltiness":    card.EdhrecSaltiness,
		"colors":              card.Colors,
		"converted_mana_cost": card.ManaValue,
		"mana_cost":           card.ManaCost,
		"rulings":             card.Rulings.AsTexts(),
		"starting_loyalty":    card.Loyalty,
		"text":                card.Text,
	}
}

func turbopufferSchema() map[string]turbopuffer.AttributeSchemaConfigParam {
	return map[string]turbopuffer.AttributeSchemaConfigParam{
		"id": {
			Type: turbopuffer.Opt(turbopuffer.AttributeType("uuid")),
		},
		"types": {
			Type: turbopuffer.Opt(turbopuffer.AttributeType("[]string")),
		},
		"power": {
			Type: turbopuffer.Opt(turbopuffer.AttributeType("string")),
		},
		"toughness": {
			Type: turbopuffer.Opt(turbopuffer.AttributeType("string")),
		},
		"name": {
			Type: turbopuffer.Opt(turbopuffer.AttributeType("string")),
			FullTextSearch: &turbopuffer.FullTextSearchConfigParam{
				RemoveStopwords: turbopuffer.Bool(false),
			},
			Filterable: turbopuffer.Bool(true),
		},
		"edhrec_rank": {
			Type: turbopuffer.Opt(turbopuffer.AttributeType("uint")),
		},
		"edhrec_saltiness": {
			Type: turbopuffer.Opt(turbopuffer.AttributeType("float")),
		},
		"colors": {
			Type: turbopuffer.Opt(turbopuffer.AttributeType("[]string")),
		},
		"converted_mana_cost": {
			Type: turbopuffer.Opt(turbopuffer.AttributeType("uint")),
		},
		"mana_cost": {
			Type: turbopuffer.Opt(turbopuffer.AttributeType("string")),
		},
		"rulings": {
			Type: turbopuffer.Opt(turbopuffer.AttributeType("[]string")),
			FullTextSearch: &turbopuffer.FullTextSearchConfigParam{
				Stemming: turbopuffer.Bool(true),
			},
		},
		"starting_loyalty": {
			Type: turbopuffer.Opt(turbopuffer.AttributeType("string")),
		},
		"text": {
			Type: turbopuffer.Opt(turbopuffer.AttributeType("string")),
			FullTextSearch: &turbopuffer.FullTextSearchConfigParam{
				Stemming:        turbopuffer.Bool(true),
				RemoveStopwords: turbopuffer.Bool(false),
			},
		},
	}
}

// Search performs a search query against the index, returning up to topk results.
func (idx *Index) Search(
	ctx context.Context,
	tpuf *turbopuffer.Client,
	query string,
	topk int,
) ([]turbopuffer.Row, error) {
	ns := tpuf.Namespace(idx.Namespace)
	resp, err := ns.Query(ctx, turbopuffer.NamespaceQueryParams{
		RankBy: turbopuffer.NewRankByTextSum([]turbopuffer.RankByText{
			turbopuffer.NewRankByTextProduct(2.0, turbopuffer.NewRankByTextBM25("name", query)),
			turbopuffer.NewRankByTextProduct(1.0, turbopuffer.NewRankByTextBM25("text", query)),
		}),
		TopK: turbopuffer.Int(int64(topk)),
		IncludeAttributes: turbopuffer.IncludeAttributesParam{
			StringArray: []string{"name", "mana_cost", "text"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("querying namespace %q: %w", idx.Namespace, err)
	}
	return resp.Rows, nil
}
