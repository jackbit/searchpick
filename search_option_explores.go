package searchpick

import (
  "log"
  "strconv"
  "reflect"
)

func (sOption *SearchOption) ExploreFields(boostField *BoostField) {
  mustNot := []interface{}{}
  should := []interface{}{}
  queries := []interface{}{}
  misspellings := sOption.SetMisspellings(boostField.Fields)

  for _, f := range boostField.Fields {
    field := f.(string)
    queriesToAdd := []interface{}
    qS := []interface{}
    
    factor := 1

    if !reflect.ValueOf(boostField.Boosts[field]).IsZero() {
      factor := boostField.Boosts[field].(float64)
    }

    sharedOptions :- map[string]interface{}{
      "query": sOption.Term,
      "boost": 10 * factor,
    }
    
    matchType := "match"
    if strings.HasSuffix(field, ".phrase") {
      if field == "_all.phrase" {
        field = "_all"
      } else {
        field = strings.ReplaceAll(field, ".phrase", ".analyzed")
      }
      matchType = "match_phrase"
    }

    if matchType == "match" {
      sharedOptions["operator"] = operator
    }

    var excludeAnalyzer string
    excludeField := field

    fieldMisspellings := false

    if misspellings.IsMisspellings && SliceContainsString(misspellings.Fields, s.BaseField(field)) {
      fieldMisspellings = true
    }

    if field == "_all" || strings.HasSuffix(field, ".analyzed") {
      if operator != "and" && fieldMisspellings {
        sharedOptions["cutoff_frequency"] = 0.001
        dupSharedOptions := sharedOptions
        dupSharedOptions["analyzer"] = "searchpick_search"
        qs = append(qs, dupSharedOptions)
        dupSharedOptions["analyzer"] = "searchpick_search2"
        qs = append(qs, dupSharedOptions)
        excludeAnalyzer = "searchpick_search2"
      }
    } else if strings.HasSuffix(field, ".exact") {
      splitFields := strings.Split(field, ".")
      joinFields := strings.Join(splitFields[:len(splitFields)-1], ".")
      dupSharedOptions := sharedOptions
      dupSharedOptions["analyzer"] = "keyword"
      joinFieldsAnalyzer := map[string]interface{}{}
      joinFieldsAnalyzer[joinFields] = dupSharedOptions
      queriesToAdd = append(queriesToAdd, map[string]interface{}{"match": joinFieldsAnalyzer})
      excludeAnalyzer = "keyword"
    } else {
      analyzer := "searchpick_autocomplete_search"
      isMatchField, _ := regexp.MatchString("\\.word_(start|middle|end)$", field)
      if isMatchField { analyzer = "searchpick_word_search" }
      excludeAnalyzer = analyzer
    }

    if !fieldMisspellings && matchType == "match" {
      dupQS := qs
      for _, q := range dupQS {
        qMap := map[string]interface{}{}
        for qK, qV := range q.(map[string]interface{}) {
          if qK.(string) != "cutoff_frequency" {
            qMap[qK.(string)] = qV
          }
        }
        qMap["fuzziness"] = editDistance
        qMap["prefix_length"] = prefixLength
        qMap["max_expansions"] = maxExpansions
        qMap["boost"] = factor
        qMap["fuzzy_transpositions"] = transpositions["fuzzy_transpositions"]
        qs = append(qs, qMap)
      }
      reducers := map[string]interface{}{}
    }

    q2 := []interface{}{}
    if strings.HasPrefix(field, "*.") {
      multiMatchType := "best_fields"
      if matchType == "match_phrase" {
        multiMatchType = "phrase"
      }
      dupQS := qs
      q2 := []interface{}{}
      for _, q := range dupQS {
        qMap := map[string]interface{}{}
        qMap["fields"] = []string{ field.(string) }
        qMap["type"] = multiMatchType
        q2 = append(q2, map[string]interface{}{ "multi_match": qMap })
      }
    } else {
      for _, q := range qs {
        matchTypeField := map[string]interface{}{}
        qField := map[string]interface{}{}
        qField[field] = q
        matchTypeField[matchType] = qField
        q2 = append(q2, matchTypeField)
      }
    }
    
    //#boost exact matches more
    isMatchWord, _ := regexp.MatchString("\\.word_(start|middle|end)$", field)
    isSearchpickWord = !reflect.ValueOf(s.Word).IsZero() && len(s.Word) > 0
    if isMatchWord && isSearchpickWord {
      wordFieldRegex := regexp.Compile("\\.word_(start|middle|end)$/")
      wordField := wordFieldRegex.ReplaceAllString(field, ".analyzed")
      wordFieldMap := map[string]interface{}{}
      wordFieldMap[wordField] = qs[0]
      wordMatchMap := map[string]interface{}{}
      wordMatchMap[matchType] = wordFieldMap
      queriesToAdd = append(queriesToAdd, map[string]interface{}{
        "bool": map[string]interface{}{
          "must": map[string]interface{}{
            "bool": map[string]interface{}{
              "should": q2
            },
          },
          "should": wordMatchMap,
        },
      })
    } else {
      queriesToAdd = append(queriesToAdd, q2)
    }

    queries = append(queries, queriesToAdd)

    if !reflect.ValueOf(sOption.Exclude).IsZero() && len(sOption.Exclude) > 0 {
      mustNot = append(mustNot, sOption.SetExclude(excludeField, excludeAnalyzer))
    }
  }

  boostField.MustNots = mustNot
  boostField.Shoulds = should
  boostField.Queries = queries
}