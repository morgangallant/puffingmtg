package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/turbopuffer/turbopuffer-go"
	"github.com/turbopuffer/turbopuffer-go/option"
)

func main() {
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	tpuf, err := newTurbopufferClient()
	if err != nil {
		log.Fatalf("failed to create turbopuffer client: %v", err)
	}

	switch {
	case *flagBuildIndex != "":
		if err := buildIndex(ctx, tpuf, *flagBuildIndex); err != nil {
			log.Fatalf("failed to build index %q: %v", *flagBuildIndex, err)
		}
	case *flagDeleteIndex != "":
		if err := deleteIndex(ctx, tpuf, *flagDeleteIndex); err != nil {
			log.Fatalf("failed to delete index %q: %v", *flagDeleteIndex, err)
		}
	case *flagServeIndex != "":
		if err := serveIndex(ctx, tpuf, *flagServeIndex); err != nil {
			log.Fatalf("failed to serve index %q: %v", *flagServeIndex, err)
		}
	default:
		log.Println(
			"no action specified, you must pass one of: -build-index, -delete-index or -serve-index",
		)
		log.Println("available flags:")
		flag.PrintDefaults()
	}
}

func buildIndex(ctx context.Context, tpuf *turbopuffer.Client, name string) error {
	fp := fmt.Sprintf("%s.json", name)
	if existing, err := LoadIndex(fp); err != nil {
		return fmt.Errorf("checking for existing index: %w", err)
	} else if existing != nil {
		log.Printf("index %q already exists, not overwriting.", fp)
		log.Printf("to delete this index fully (including from turbopuffer), use -delete-index %q", name)
		return nil
	}

	set, err := mtgSet()
	if err != nil {
		return fmt.Errorf("choosing mtg set: %w", err)
	}

	index, err := NewIndex(ctx, tpuf, name, set)
	if err != nil {
		return fmt.Errorf("creating new index: %w", err)
	}

	log.Printf("successfully created index %q (backed by tpuf namespace %q)", name, index.Namespace)
	log.Printf("to serve this index, use -serve-index %q", name)

	return nil
}

func deleteIndex(ctx context.Context, tpuf *turbopuffer.Client, name string) error {
	index, err := LoadIndex(name)
	if err != nil {
		return fmt.Errorf("loading index %q: %w", name, err)
	} else if index == nil {
		log.Printf("index %q does not exist, nothing to do", name)
		return nil
	}

	if err := index.Delete(ctx, tpuf); err != nil {
		return fmt.Errorf("deleting index %q: %w", name, err)
	}

	log.Printf("successfully deleted index %q (from turbopuffer and local disk)", name)

	return nil
}

func serveIndex(ctx context.Context, tpuf *turbopuffer.Client, name string) error {
	index, err := LoadIndex(name)
	if err != nil {
		return fmt.Errorf("loading index %q: %w", name, err)
	} else if index == nil {
		return fmt.Errorf("index %q does not exist, cannot serve. run -build-index first", name)
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("enter your query: ")
		query, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("reading query from stdin: %w", err)
		}
		query = strings.Trim(query, " \r\n")

		start := time.Now()
		results, err := index.Search(ctx, tpuf, query, 10)
		if err != nil {
			return fmt.Errorf("searching index %q: %w", name, err)
		}

		log.Printf("found %d results in %d ms:", len(results), time.Since(start).Milliseconds())
		for i, row := range results {
			log.Printf("%d: %s (%s)\n%s", i+1, row["name"], row["mana_cost"], row["text"])
		}
	}
}

func newTurbopufferClient() (*turbopuffer.Client, error) {
	apiKey, err := tpufApiKey()
	if err != nil {
		return nil, fmt.Errorf("getting turbopuffer api key: %w", err)
	}
	client := turbopuffer.NewClient(
		option.WithAPIKey(apiKey),
		option.WithRegion(tpufRegion()),
	)
	return &client, nil
}
