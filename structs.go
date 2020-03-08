package searchpick

type SearchOption struct {
  Term            string
  Fields          []string
  Select          []string
  Exclude         []interface{}
  Operator        string
  Page            int64
  PerPage         int64
  Limit           int64
  Padding         int64
  Offset          int64
  Order           map[string]interface{}
  Where           map[string]interface{}
  Similar         bool
  Match           string
  Body            map[string]interface{}
  BodyJson        string
  Misspellings    map[string]interface{}
  Conversions     string
  ConversionsTerm string
  Aggs            interface{} //Slice or Map
  SmartAggs       string //"false" or "true"
  BoostBy         interface{}
  Boost           string
  BoostWhere      map[string]interface{}
  BoostByDistance map[string]interface{}
  BoostByRecency  map[string]interface{}
  Explain         bool
  IndicesBoost    map[string]interface{}
  Suggest         bool
}


type BoostField struct {
  Fields   []string
  Boosts   map[string]interface{}
  MustNots []interface{}
  Shoulds  []interface{}
  Queries  []interface{}
}

type SearchQuery struct {
  Query         map[string]interface{}
  Field         string
  Operator      string
  OperatorQuery interface{}
}

type SearchFilter struct {
  Filters         []interface{}
  Where           map[string]interface{}
  Field           string
  Payloads        map[string]interface{}
  CustomFilters   []interface{}
  MultiplyFilters []interface{}
}

type SearchResult struct {
  Params   *SearchOption
  Results  []interface{}
}

type SearchData struct {
  Id       string
  BodyJson string
}

type Searchpick struct {
  Name              string
  BatchSize         float64
  Callbacks         string
  CaseSensitive     bool
  Conversions       []string
  DefaultFields     []string
  Filterable        []string
  GeoShape          []string
  Highlight         []string
  IndexName         string
  IndexType         string
  IndexPrefix       string
  IgnoreAbove       float64
  Locations         []string
  Mappings          map[string]interface{}
  Match             string
  MergeMappings     bool
  Routing           string // "true", "false", default: ""
  Searchable        []string
  Settings          map[string]interface{}
  Similarity        string
  SpecialCharacters string // "true", "false", default: ""
  Stem              string // "true", "false", default: ""
  StemConversions   bool
  Suggest           []string
  Synonyms          []interface{}
  TextEnd           []string
  TextMiddle        []string
  TextStart         []string
  Word              []string
  WordEnd           []string
  WordMiddle        []string
  WordStart         []string
  FinalMappings     map[string]interface{}
  Error             error
  Client            *elastic.Client
  SearchData        *SearchData
  Version           string
}