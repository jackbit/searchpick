package searchpick

import (
  "log"
  "encoding/json"
  "reflect"
  "regexp"
)

func (sf BoostField) String() string {
  j, _ := json.Marshal(sf)
  return string(j)
}

func (eq SearchQuery) String() string {
  j, _ := json.Marshal(eq)
  return string(j)
}

func (s *Searchpick) SetConversions(sOption *SearchOption) []interface{} {
  conversionFields := []string{}
  if sOption.Conversions != "" && sOption.Conversions != "false" {
    conversionFields = append(conversionFields, sOption.Conversions)
  } else if !reflect.ValueOf(s.Conversions).IsZero() && len(s.Conversions) > 0 {
    conversionFields = s.Conversions
  }

  conversionTerm := s.Term
  if sOption.ConversionsTerm != "" {
    conversionTerm = sOption.ConversionsTerm
  }

  conversions := []interface{}{}
  if !reflect.ValueOf(s.Conversions).IsZero() && len(s.Conversions) > 0 {
    for _, conversionField := range conversionFields {
      matchKey := conversionField.(string)+".query"
      matchCount := conversionField.(string)+".count"
      nested := map[string]interface{}{
        "nested": map[string]interface{}{
          "path": conversionField,
          "score_mode": "sum",
          "query": map[string]interface{}{
            "function_score": map[string]interface{}{
              "boost_mode": "replace",
              "query": map[string]interface{}{
                "match": map[string]interface{}{
                  matchKey: conversionTerm,
                },
              },
              "field_value_factor": map[string]interface{}{
                "field": matchCount,
              },
            }, 
          },
        },
      }
      conversions := append(conversions, nested)
    }
    return conversions
  } else {
    return conversions
  }
}

func (s *Searchpick) BaseField(str string) string {
  exp := "\\.(analyzed|word_start|word_middle|word_end|text_start|text_middle|text_end|exact)$"
  r, _ := regexp.Compile(exp)
  return r.ReplaceAllString(str, "")
}

func (s *Searchpick) Search(sOption *SearchOption) SearchResult {
  sQuery := &SearchQuery{ Query: map[string]interface{}{} } 

  if sOption.Term == "" { sOption.Term = "*" }
  if sOption.Match == "" { sOption.Match = "word" }
  sOption.SetPagination()
  boostField := sOption.SetFields()

  isLoad = true
  operator := sOption.Operator
  if operator == "" { operator = "and" }
  isAll = false
  if sOption.Term == "*" {isAll = true}
  
  var payload map[string]interface{}

  if !reflect.ValueOf(sOption.Body).IsZero() {
    payload = sOption.Body
  } else if sOption.BodyJson != "" {
    _ = json.Unmarshal([]byte(sOption.BodyJson), &payload)
  } else {

    if sOption.Similar {
      sQuery.SetSimilar(sOption.Term, boostField)
    } else if isAll && !reflect.ValueOf(sOption.Exclude).IsZero() && len(sOption.Exclude) > 0 {
      sQuery.Query["match_all"] = map[string]interface{}{}
    } else {

      sOption.ExploreFields(boostField)

      if isAll {
        query :=  map[string]interface{}{"match_all": map[string]interface{}{}}
        boostField.Shoulds = []interface{}
      } else {
        queriesPayload = []map[string]interface{}{}

        for _, qs := range boostField.Queries {
          
          queriesPayload = append(queriesPayload, []map[string]interface{}{
            "dis_max": map[string]interface{}{
              "queries": qs,
            },
          })
        }

        payload = map[string]interface{}{
          "bool": map[string]interface{}{
            "should": queriesPayload,
          },
        }

        boostField.Shoulds = append(boostField.Shoulds s.SetConversions(sOption))
      }

      query = payload
    }

    //Note: Searchpick, skip inheritance of searchkick
    // start everything as efficient filters
    // move to post_filters as aggs demand

    sFilter := &SearchFilter{ Filters: []interface{}{}, Where: sOption.Where }
    sFilter.SetFilters().SetAggregation() //including post-filters
    
    // log.Println(sQuery.String())
    // log.Println(boostField.String())
    // log.Println(s.String())
    j, _ := json.Marshal(sFilter.Filters)
  }

  log.Println(string(j))
  esResult := SearchResult{}
  return esResult
}

