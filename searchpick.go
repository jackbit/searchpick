package searchpick

import (
  "log"
  "reflect"
  "errors"
  "github.com/gobuffalo/envy"
  "github.com/gertd/go-pluralize"
  "gopkg.in/olivere/elastic.v6"
  es_config "gopkg.in/olivere/elastic.v6/config"
)

type SearchData struct {
  Id string
  BodyJson string
}

type Searchpick struct {
  Name string
  BatchSize float64
  Callbacks string
  CaseSensitive bool
  Conversions []string
  DefaultFields []string
  Filterable []string
  GeoShape []string
  Highlight []string
  IndexName string
  IndexType string
  IndexPrefix string
  Inheritance bool
  IgnoreAbove float64
  Locations []string
  Mappings map[string]interface{}
  Match string
  MergeMappings bool
  Routing string // "true", "false", default: ""
  Searchable []string
  Settings map[string]interface{}
  Similarity string
  SpecialCharacters string // "true", "false", default: ""
  Stem string // "true", "false", default: ""
  StemConversions bool
  Suggest []string
  Synonyms []interface{}
  TextEnd []string
  TextMiddle []string
  TextStart []string
  Word []string
  WordEnd []string
  WordMiddle []string
  WordStart []string
  FinalMappings map[string]interface{}
  Error error
  Client *elastic.Client
  SearchData *SearchData
  Version string
}

var ES *elastic.Client

func (s *Searchpick) CheckIndexType() *Searchpick {
  if s.Error != nil { return s }

  if s.IndexType == "" {
    if s.Name == "" {
      s.Error = errors.New("Name is required or add IndexType in Model")
      log.Panic(s.Error)
    } else {
      s.IndexType = pluralize.NewClient().Singular(s.Name)
    }
  }
  return s
}

func (s *Searchpick) CheckIndexName() *Searchpick {
  if s.Error != nil { return s }
  
  if s.IndexName == "" {
    if s.Name == "" {
      s.Error = errors.New("Name is required or add IndexName in Model")
    } else {
      s.IndexName = pluralize.NewClient().Plural(s.Name)
    }
  }
  return s
}

func (s *Searchpick) Connect() *Searchpick {
  esURL := envy.Get("ES_URL", "http://localhost:9200")
  esUsername := envy.Get("ES_USERNAME", "")
  esPassword := envy.Get("ES_PASSWORD", "")

  c := es_config.Config{URL: esURL}

  if esUsername != "" && esPassword != "" {
    c.Username = esUsername
    c.Password = esPassword
  }

  client, err := elastic.NewClient(elastic.SetURL(esURL), elastic.SetBasicAuth(c.Username, c.Password), elastic.SetSniff(false))

  s.Client = client
  s.Error = err

  if err != nil {
    log.Println(err)
  }
  
  return s
}

func (s *Searchpick) Hook() *Searchpick {

  namespace := envy.Get("ES_NAMESPACE", "searchpick_dev")
  if s.IndexPrefix == "" && namespace != "" {
    s.IndexPrefix = namespace
  }

  if reflect.ValueOf(ES).IsNil() {
    s.Connect()
    if s.Error == nil {
      ES = s.Client
    }
  } else {
    s.Client = ES
  }

  return s
}

func init() {
  s := &Searchpick{}
  s.Hook()
  if s.Error == nil {
    ES = s.Client
  }
}
