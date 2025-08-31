package main

import (
	"errors"
	"flag"
	"log"
)

var (
	flagTpufRegion = flag.String(
		"tpuf-region",
		"",
		"turbopuffer region (see: turbopuffer.com/docs/regions; pick the one closest to you)",
	)
	flagTpufApiKey = flag.String(
		"tpuf-api-key",
		"",
		"your turbopuffer API key",
	)
	flagBuildIndex = flag.String(
		"build-index",
		"",
		"name of the index to build. will create a json file with this name in current directory",
	)
	flagDeleteIndex = flag.String(
		"delete-index",
		"",
		"name of the index to delete both locally and from turbopuffer",
	)
	flagServeIndex = flag.String(
		"serve-index",
		"",
		"name of the index to serve via an HTTP server",
	)
	flagSet = flag.String(
		"set",
		"",
		"which mtgjson set to download and index (vintage, standard, pioneer, pauper, modern)",
	)
)

func tpufApiKey() (string, error) {
	if *flagTpufApiKey != "" {
		return *flagTpufApiKey, nil
	}
	return "", errors.New("missing --tpuf-api-key flag")
}

func tpufRegion() string {
	if *flagTpufRegion != "" {
		return *flagTpufRegion
	}
	log.Println("no --tpuf-region flag provided, defaulting to gcp-us-central1")
	return "gcp-us-central1"
}

func mtgSet() (Set, error) {
	set := Set(*flagSet)
	if !set.Valid() {
		return "", errors.New(
			"invalid set, must be one of vintage, standard, pioneer, pauper, modern",
		)
	}
	return set, nil
}

// Set is an enumeration of supported mtgjson MTG sets.
type Set string

// List of supported sets. Notably, we care about the atomic version of the set.
var (
	Vintage  Set = "vintage"
	Standard Set = "standard"
	Pioneer  Set = "pioneer"
	Pauper   Set = "pauper"
	Modern   Set = "modern"
)

// DownloadURL returns the mtgjson download URL for the given set.
// Specifically, downloads the atomic version of the set, i.e. unique cards only,
// ignoring reprints and variations.
func (s Set) DownloadURL() (string, error) {
	switch s {
	case Vintage:
		return "https://mtgjson.com/api/v5/LegacyAtomic.json", nil
	case Standard:
		return "https://mtgjson.com/api/v5/StandardAtomic.json", nil
	case Pioneer:
		return "https://mtgjson.com/api/v5/PioneerAtomic.json", nil
	case Pauper:
		return "https://mtgjson.com/api/v5/PauperAtomic.json", nil
	case Modern:
		return "https://mtgjson.com/api/v5/ModernAtomic.json", nil
	default:
		return "", errors.New("unknown set")
	}
}

func (s Set) Valid() bool {
	switch s {
	case Vintage, Standard, Pioneer, Pauper, Modern:
		return true
	default:
		return false
	}
}
