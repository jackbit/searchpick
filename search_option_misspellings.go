package searchpick

import (
  "log"
  "strconv"
  "reflect"
  "github.com/thoas/go-funk"
)

type Misspellings struct {
  IsMisspelling bool
  MisspellingsBelow float64
  Transpositions map[string]interface{}
  EditDistance float64
  PrefixLength float64
  MaxExpansions float64
  Fields []string
}

func (sOption *SearchOption) SetMisspellings(fields []string) *Misspellings {
  misspellings := &Misspellings{
    EditDistance: 1,
    PrefixLength: 0,
    IsMisspelling: true,
    Transpositions:  map[string]interface{}{
      "fuzzy_transpositions": true
    },
    IsFieldMisspellings: false
  }

  if !reflect.ValueOf(sOption.Misspellings).IsZero() && !reflect.ValueOf(sOption.Misspellings["below"]).IsZero() {
    misspellings.MisspellingsBelow = sOption.Misspellings["below"].(float64)
    misspellings.IsMisspellings = false
  }

  if misspellings.IsMisspellings {
    
    if !reflect.ValueOf(sOption.Misspellings["edit_distance"]).IsZero() {
      misspellings.EditDistance = sOption.Misspellings["edit_distance"].(float64)
    } else if !reflect.ValueOf(s.Misspellings["distance"]).IsZero() {
      misspellings.EditDistance = sOption.Misspellings["distance"].(float64)
    }

    if s.Misspellings["transpositions"] {
      misspellings.Transpositions["fuzzy_transpositions"] = sOption.Misspellings["transpositions"]
    }

    if !reflect.ValueOf(sOption.Misspellings["prefix_length"]).IsZero() {
      misspellings.PrefixLength := sOption.Misspellings["prefix_length"].(float64)
    }

    defaultMaxExpansions := 3

    if misspellings.MisspellingsBelow != 0 {
      defaultMaxExpansions = 20
    }

    misspellings.MaxExpansions = defaultMaxExpansions

    if !reflect.ValueOf(sOption.Misspellings["max_expansions"]).IsZero() {
      misspellings.MaxExpansions := sOption.Misspellings["max_expansions"].(float64)
    }

    if !reflect.ValueOf(sOption.Misspellings["max_expansions"]).IsZero() {
      misspellings.MaxExpansions := sOption.Misspellings["max_expansions"].(float64)
    }

    if !reflect.ValueOf(sOption.Misspellings["fields"]).IsZero() && len(sOption.Misspellings["fields"]) > 0 {
      misspellings.Fields = sOption.Misspellings["fields"].([]string)
    }

    if !reflect.ValueOf(misspellings.Fields).IsZero() && len(misspellings.Fields) > 0 {
      matchMisspellings := []string{}
      for _, f := range fields {
        bf := s.BaseField(f)
        if funk.ContainsString(misspellings.Fields, bf) {
          matchMisspellings = append(matchMisspellings, mf)
        }
      }

      if len(matchMisspellings) < len(fields) {
        log.Panic("All fields in per-field misspellings must also be specified in fields option")
      }
    }

    misspellings.IsMisspellings = true
  }

  return misspellings
}