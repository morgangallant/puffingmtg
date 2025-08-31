package main

type AtomicSet struct {
	Data map[string][]AtomicCard `json:"data"`
	Meta struct {
		Date    string `json:"date"`    // Example: "2025-08-30"
		Version string `json:"version"` // Example: "5.2.2+20250830"
	}
}

type AtomicCard struct {
	AsciiName               *string           `json:"asciiName,omitempty"`
	AttractionLights        []int             `json:"attractionLights,omitempty"`
	ColorIdentity           []string          `json:"colorIdentity"`
	ColorIndicator          []string          `json:"colorIndicator,omitempty"`
	Colors                  []string          `json:"colors"`
	ConvertedManaCost       float64           `json:"convertedManaCost"`
	Defense                 *string           `json:"defense,omitempty"`
	EdhrecRank              *int              `json:"edhrecRank,omitempty"`
	EdhrecSaltiness         *float64          `json:"edhrecSaltiness,omitempty"`
	FaceConvertedManaCost   *float64          `json:"faceConvertedManaCost,omitempty"`
	FaceManaValue           *float64          `json:"faceManaValue,omitempty"`
	FaceName                *string           `json:"faceName,omitempty"`
	FirstPrinting           *string           `json:"firstPrinting,omitempty"`
	ForeignData             []ForeignData     `json:"foreignData,omitempty"`
	Hand                    *string           `json:"hand,omitempty"`
	HasAlternativeDeckLimit *bool             `json:"hasAlternativeDeckLimit,omitempty"`
	Identifiers             Identifiers       `json:"identifiers"`
	IsFunny                 *bool             `json:"isFunny,omitempty"`
	IsGameChanger           *bool             `json:"isGameChanger,omitempty"`
	IsReserved              *bool             `json:"isReserved,omitempty"`
	Keywords                []string          `json:"keywords,omitempty"`
	Layout                  string            `json:"layout"`
	LeadershipSkills        *LeadershipSkills `json:"leadershipSkills,omitempty"`
	Legalities              Legalities        `json:"legalities"`
	Life                    *string           `json:"life,omitempty"`
	Loyalty                 *string           `json:"loyalty,omitempty"`
	ManaCost                *string           `json:"manaCost,omitempty"`
	ManaValue               float64           `json:"manaValue"`
	Name                    string            `json:"name"`
	Power                   *string           `json:"power,omitempty"`
	Printings               []string          `json:"printings,omitempty"`
	PurchaseUrls            PurchaseUrls      `json:"purchaseUrls"`
	RelatedCards            RelatedCards      `json:"relatedCards"`
	Rulings                 Rulings           `json:"rulings,omitempty"`
	Side                    *string           `json:"side,omitempty"`
	Subsets                 []string          `json:"subsets,omitempty"`
	Subtypes                []string          `json:"subtypes"`
	Supertypes              []string          `json:"supertypes"`
	Text                    *string           `json:"text,omitempty"`
	Toughness               *string           `json:"toughness,omitempty"`
	Type                    string            `json:"type"`
	Types                   []string          `json:"types"`
}

type ForeignData struct {
	FaceName    *string     `json:"faceName,omitempty"`
	FlavorText  *string     `json:"flavorText,omitempty"`
	Identifiers Identifiers `json:"identifiers"`
	Language    string      `json:"language"`
	Name        string      `json:"name"`
	Text        *string     `json:"text,omitempty"`
	Type        *string     `json:"type,omitempty"`
}

type Identifiers struct {
	AbuId                    *string `json:"abuId,omitempty"`
	CardKingdomEtchedId      *string `json:"cardKingdomEtchedId,omitempty"`
	CardKingdomFoilId        *string `json:"cardKingdomFoilId,omitempty"`
	CardKingdomId            *string `json:"cardKingdomId,omitempty"`
	CardsphereId             *string `json:"cardsphereId,omitempty"`
	CardsphereFoilId         *string `json:"cardsphereFoilId,omitempty"`
	CardtraderId             *string `json:"cardtraderId,omitempty"`
	CsiId                    *string `json:"csiId,omitempty"`
	McmId                    *string `json:"mcmId,omitempty"`
	McmMetaId                *string `json:"mcmMetaId,omitempty"`
	MiniaturemarketId        *string `json:"miniaturemarketId,omitempty"`
	MtgArenaId               *string `json:"mtgArenaId,omitempty"`
	MtgjsonFoilVersionId     *string `json:"mtgjsonFoilVersionId,omitempty"`
	MtgjsonNonFoilVersionId  *string `json:"mtgjsonNonFoilVersionId,omitempty"`
	MtgjsonV4Id              *string `json:"mtgjsonV4Id,omitempty"`
	MtgoFoilId               *string `json:"mtgoFoilId,omitempty"`
	MtgoId                   *string `json:"mtgoId,omitempty"`
	MultiverseId             *string `json:"multiverseId,omitempty"`
	ScgId                    *string `json:"scgId,omitempty"`
	ScryfallId               *string `json:"scryfallId,omitempty"`
	ScryfallCardBackId       *string `json:"scryfallCardBackId,omitempty"`
	ScryfallOracleId         *string `json:"scryfallOracleId,omitempty"`
	ScryfallIllustrationId   *string `json:"scryfallIllustrationId,omitempty"`
	TcgplayerProductId       *string `json:"tcgplayerProductId,omitempty"`
	TcgplayerEtchedProductId *string `json:"tcgplayerEtchedProductId,omitempty"`
	TntId                    *string `json:"tntId,omitempty"`
}

type Rulings []Ruling

func (r Rulings) AsTexts() []string {
	texts := make([]string, 0, len(r))
	for _, ruling := range r {
		texts = append(texts, ruling.Text)
	}
	return texts
}

type Ruling struct {
	Date string `json:"date"`
	Text string `json:"text"`
}

type LeadershipSkills struct {
	Brawl       bool `json:"brawl"`
	Commander   bool `json:"commander"`
	Oathbreaker bool `json:"oathbreaker"`
}

type Legalities struct {
	Alchemy         *string `json:"alchemy,omitempty"`
	Brawl           *string `json:"brawl,omitempty"`
	Commander       *string `json:"commander,omitempty"`
	Duel            *string `json:"duel,omitempty"`
	Explorer        *string `json:"explorer,omitempty"`
	Future          *string `json:"future,omitempty"`
	Gladiator       *string `json:"gladiator,omitempty"`
	Historic        *string `json:"historic,omitempty"`
	HistoricBrawl   *string `json:"historicbrawl,omitempty"`
	Legacy          *string `json:"legacy,omitempty"`
	Modern          *string `json:"modern,omitempty"`
	Oathbreaker     *string `json:"oathbreaker,omitempty"`
	Oldschool       *string `json:"oldschool,omitempty"`
	Pauper          *string `json:"pauper,omitempty"`
	PauperCommander *string `json:"paupercommander,omitempty"`
	Penny           *string `json:"penny,omitempty"`
	Pioneer         *string `json:"pioneer,omitempty"`
	Predh           *string `json:"predh,omitempty"`
	Premodern       *string `json:"premodern,omitempty"`
	Standard        *string `json:"standard,omitempty"`
	StandardBrawl   *string `json:"standardbrawl,omitempty"`
	Timeless        *string `json:"timeless,omitempty"`
	Vintage         *string `json:"vintage,omitempty"`
}
type PurchaseUrls struct {
	CardKingdom       *string `json:"cardKingdom,omitempty"`
	CardKingdomEtched *string `json:"cardKingdomEtched,omitempty"`
	CardKingdomFoil   *string `json:"cardKingdomFoil,omitempty"`
	Cardmarket        *string `json:"cardmarket,omitempty"`
	Tcgplayer         *string `json:"tcgplayer,omitempty"`
	TcgplayerEtched   *string `json:"tcgplayerEtched,omitempty"`
}

type RelatedCards struct {
	ReverseRelated []string `json:"reverseRelated,omitempty"`
	Spellbook      []string `json:"spellbook,omitempty"`
}
