package searchpick

import (
  "reflect"
  "strings"
  "log"
)

func (s *Searchpick) SettingCharFilter(filter []string) map[string]interface{} {
  tokenizer := s.SettingTokenizer("standard", filter) //.(map[string]interface{})
  tokenizer["char_filter"] = []string{ "ampersand" }
  return tokenizer
}

func (s *Searchpick) SettingTokenizer(tokenizer string, filter []string) map[string]interface{} {
  defaultFilter := []string{}

  if s.CaseSensitive {
    if s.SpecialCharacters != "false" { defaultFilter = []string{"asciifolding"} }
  } else if s.SpecialCharacters != "false" {
    defaultFilter = []string{"lowercase", "asciifolding"}
  } else {
    defaultFilter = []string{"lowercase"}
  }

  return map[string]interface{}{
    "type": "custom",
    "tokenizer": tokenizer,
    "filter": append(defaultFilter, filter...),
  }
}

func (s *Searchpick) CharsFilter(chars []string) []string {
  if s.Stem != "false" {
    return append(chars, "searchpick_stemmer")
  } else {
    return chars
  }
}

func (s *Searchpick) SettingSynonyms() {
  newSynonyms := []interface{}{}
  if !reflect.ValueOf(s.Synonyms).IsZero() && len(s.Synonyms) > 0 {
    for _, synonym := range s.Synonyms {
      typeSynonym := reflect.TypeOf(synonym).Kind()

      if typeSynonym == reflect.Slice {
        syns := synonym.([]interface{})
        for _, syn := range syns {
          if !reflect.ValueOf(syn).IsZero(){
            newSynonyms = append( newSynonyms, strings.ReplaceAll(strings.ToLower(syn.(string)), " ", "") )
          }
        }
      } else if typeSynonym == reflect.String && !reflect.ValueOf(synonym).IsZero() {
        newSynonyms = append(newSynonyms, strings.ToLower(synonym.(string)))
      }
    }
  }
  s.Synonyms = newSynonyms
}

func (s *Searchpick) SettingAnalyzer(defaultAnalyzer string) map[string]interface{} {
  keywordFilter := []string{ "lowercase" }
  
  if s.StemConversions && s.Stem != "false" {
    keywordFilter = append(keywordFilter, "searchpick_stemmer")
  }

  wordStart := []string{"searchpick_edge_ngram"}
  wordMiddle := []string{"searchpick_ngram"}
  wordEnd := []string{"reverse", "searchpick_edge_ngram", "reverse"}
  charsDefaultAnalyzer := []string{"searchpick_index_shingle"}

  if len(s.Synonyms) > 0 {
    wordStart = append([]string{"searchpick_synonym"}, wordStart...)
    wordMiddle = append([]string{"searchpick_synonym"}, wordMiddle...)
    wordEnd = append([]string{"searchpick_synonym"}, wordEnd...)
    charsDefaultAnalyzer = append([]string{"searchpick_synonym"}, charsDefaultAnalyzer...)
  }

  keywordValue := s.SettingTokenizer("keyword", keywordFilter)
  keywordValue["filter"] = keywordFilter

  analyzer := map[string]interface{}{
    "searchpick_keyword": keywordValue,
    "searchpick_search": s.SettingCharFilter(s.CharsFilter([]string{"searchpick_search_shingle"})),
    "searchpick_search2": s.SettingCharFilter(s.CharsFilter([]string{})),
    "searchpick_autocomplete_search": s.SettingTokenizer("keyword", []string{}),
    "searchpick_word_search": s.SettingTokenizer("standard", []string{}),
    "searchpick_suggest_index": s.SettingTokenizer("standard", []string{"searchpick_suggest_shingle"}),
    "searchpick_text_start_index": s.SettingTokenizer("keyword", []string{"searchpick_edge_ngram"}),
    "searchpick_text_middle_index": s.SettingTokenizer("keyword", []string{"searchpick_ngram"}),
    "searchpick_text_end_index": s.SettingTokenizer("keyword", []string{"reverse", "searchpick_edge_ngram", "reverse"}),
    "searchpick_word_start_index": s.SettingTokenizer("standard", wordStart),
    "searchpick_word_middle_index": s.SettingTokenizer("standard", wordMiddle),
    "searchpick_word_end_index": s.SettingTokenizer("standard", wordEnd),
  }

  analyzer[defaultAnalyzer] = s.SettingCharFilter(s.CharsFilter(charsDefaultAnalyzer))
  
  return analyzer
}

func (s *Searchpick) SettingFilter() map[string]interface{} {
  filter := map[string]interface{}{
    "searchpick_index_shingle": map[string]interface{}{"type": "shingle", "token_separator": ""},
    "searchpick_search_shingle": map[string]interface{}{
      "type": "shingle",
      "token_separato": "",
      "output_unigrams": false,
      "output_unigrams_if_no_shingles": true,
    },
    "searchpick_suggest_shingle": map[string]interface{}{
      "type": "shingle",
      "max_shingle_size": 5,
    },
    "searchpick_edge_ngram": map[string]interface{}{
      "type": "edgeNGram",
      "min_gram": 1,
      "max_gram": 50,
    },
    "searchpick_ngram": map[string]interface{}{
      "type": "nGram",
      "min_gram": 1,
      "max_gram": 50,
    },
  }

  if s.Stem != "false" {
    filter["searchpick_stemmer"] = map[string]interface{}{
      "type": "snowball",
      "language": "English",
    }
  }

  return filter
}

// https://github.com/ankane/searchpick/blob/master/lib/searchpick/index_options.rb
func (s *Searchpick) SettingsMappings() map[string]interface{} {
  s.CheckIndexName().CheckIndexType()
  defaultType := "text"
  defaultAnalyzer := "searchpick_index"
  keywordMapping := map[string]interface{}{
    "type": "keyword",
  }

  indexTrueValue := true
  indexFalseValue := false

  if reflect.ValueOf(s.IgnoreAbove).IsZero() {
    keywordMapping["ignore_above"] = 30000
  } else {
    keywordMapping["ignore_above"] = s.IgnoreAbove
  }

  settings := map[string]interface{}{
    "index": map[string]interface{}{
      "max_ngram_diff": 49,
      "max_shingle_diff": 4,
    },
  }

  s.SettingSynonyms()
  
  analyzer := s.SettingAnalyzer(defaultAnalyzer)
  filter := s.SettingFilter()
  
  charFilter := map[string]interface{}{
    "ampersand":  map[string]interface{}{
      "type": "mapping",
      "mappings": []string{"&=> and "},
    },
  }

  // add similarity as settings
  if !reflect.ValueOf(s.Similarity).IsZero() {
    settings["similarity"] = map[string]interface{}{
      "default": map[string]interface{}{
        "type": s.Similarity,
      },
    }
  }
  
  // add synonyms to filters
  if len(s.Synonyms) > 0 {
    filter["searchpick_synonym"] = map[string]interface{}{
      "type": "synonym",
      "synonyms": s.Synonyms,
    }
  }

  mappings := map[string]interface{}{}
 
  if !reflect.ValueOf(s.Mappings).IsZero() && !s.MergeMappings {
    mappings[s.IndexType] = s.Mappings
    settings = s.Settings
    if reflect.ValueOf(settings).IsZero() {
      settings = map[string]interface{}{}
    }
    log.Println("asdasdas")
  } else {
    // add add conversions list to mappings
    if reflect.TypeOf(s.Conversions).Kind() == reflect.Slice {
      for _, conversion := range s.Conversions {
        mappings[conversion] = map[string]interface{}{
          "type": "nested",
          "properties": map[string]interface{}{
            "query": map[string]interface{}{
              "type": defaultType,
              "analyzer": "searchpick_keyword",
            },
            "count": map[string]interface{}{
              "type": "integer",
            },
          },
        }
      }
    }
    
    isWord := true
    
    // if (s.Match != "" || s.Match == "word") {
    //   isWord = true
    // }

    analyzedFieldOptions := map[string]interface{}{
      "type": defaultType,
      "index": indexTrueValue,
      "analyzer": defaultAnalyzer,
    }

    mappingOptions := map[string]interface{}{}
    mappingValues := []string{}
    mappingMatches := []string{}

    if !reflect.ValueOf(s.Suggest).IsZero() && len(s.Suggest) > 0 {
      mappingOptions["suggest"] = s.Suggest
      mappingValues = append(mappingValues, s.Suggest...)
      mappingMatches = append(mappingMatches, "suggest")
    }
    
    mappingOptions["is_word"] = false
    if !reflect.ValueOf(s.Word).IsZero() && len(s.Word) > 0 {
      mappingOptions["word"] = s.Word
      mappingValues = append(mappingValues, s.Word...)
      mappingOptions["is_word"] = true
    }
    
    mappingOptions["is_text_start"] = false
    if !reflect.ValueOf(s.TextStart).IsZero() && len(s.TextStart) > 0 {
      mappingOptions["text_start"] = s.TextStart
      mappingValues = append(mappingValues, s.TextStart...)
      mappingOptions["is_text_start"] = true
      mappingMatches = append(mappingMatches, "text_start")
    }
    
    mappingOptions["is_text_middle"] = false
    if !reflect.ValueOf(s.TextMiddle).IsZero() && len(s.TextMiddle) > 0 {
      mappingOptions["text_middle"] = s.TextMiddle
      mappingValues = append(mappingValues, s.TextMiddle...)
      mappingOptions["is_text_middle"] = true
      mappingMatches = append(mappingMatches, "text_middle")
    }
    
    mappingOptions["is_text_end"] = false
    if !reflect.ValueOf(s.TextEnd).IsZero() && len(s.TextEnd) > 0 {
      mappingOptions["text_end"] = s.TextEnd
      mappingValues = append(mappingValues, s.TextEnd...)
      mappingOptions["is_text_end"] = true
      mappingMatches = append(mappingMatches, "text_end")
    }

    mappingOptions["is_word_start"] = false
    if !reflect.ValueOf(s.WordStart).IsZero() && len(s.WordStart) > 0 {
      mappingOptions["word_start"] = s.WordStart
      mappingValues = append(mappingValues, s.WordStart...)
      mappingOptions["is_word_start"] = true
      mappingMatches = append(mappingMatches, "word_start")
    }
    
    mappingOptions["is_word_end"] = false
    if !reflect.ValueOf(s.WordEnd).IsZero() && len(s.WordEnd) > 0 {
      mappingOptions["word_end"] = s.WordEnd
      mappingValues = append(mappingValues, s.WordEnd...)
      mappingOptions["is_word_end"] = true
      mappingMatches = append(mappingMatches, "word_end")
    }
    
    mappingOptions["is_word_middle"] = false
    if !reflect.ValueOf(s.WordMiddle).IsZero() && len(s.WordMiddle) > 0 {
      mappingOptions["word_middle"] = s.WordMiddle
      mappingValues = append(mappingValues, s.WordMiddle...)
      mappingOptions["is_word_middle"] = true
      mappingMatches = append(mappingMatches, "word_middle")
    }

    mappingOptions["is_highlight"] = false
    if !reflect.ValueOf(s.Highlight).IsZero() && len(s.Highlight) > 0 {
      mappingOptions["highlight"] = s.Highlight
      mappingValues = append(mappingValues, s.Highlight...)
      mappingOptions["is_highlight"] = true
    }
    
    mappingOptions["is_searchable"] = false
    if !reflect.ValueOf(s.Searchable).IsZero() && len(s.Searchable) > 0 {
      searchables := []string{}

      for _, searchable := range s.Searchable {
        if searchable != "_all" {
          searchables = append(searchables, searchable)
        }
      }

      if len(searchables) > 0 {
        mappingOptions["searchable"] = searchables
        mappingOptions["is_searchable"] = true
      }
    }
    
    mappingOptions["is_filterable"] = false
    if !reflect.ValueOf(s.Filterable).IsZero() && len(s.Filterable) > 0 {
      mappingOptions["filterable"] = s.Filterable
      mappingOptions["is_filterable"] = true
      mappingValues = append(mappingValues, s.Filterable...)
    }

    uniqMappingValues := SliceUniqString(mappingValues)
    
    var field string
    for _, field = range uniqMappingValues {
      fields := map[string]interface{}{}

      if mappingOptions["is_filterable"].(bool) && !SliceContainsString(mappingOptions["filterable"].([]string), field) {
        fields[field] = map[string]interface{}{
          "type": defaultType,
          "index": indexFalseValue,
        }
      } else {
        fields[field] = analyzedFieldOptions
      }

      if !mappingOptions["is_searchable"].(bool) || SliceContainsString(mappingOptions["searchable"].([]string), field) {
        if isWord {
          fields["analyzed"] = analyzedFieldOptions

          if mappingOptions["is_highlight"].(bool) && SliceContainsString(mappingOptions["highlight"].([]string), field) {
            termVector := fields["analyzed"].(map[string]interface{})
            termVector["term_vector"] = "with_positions_offsets"
          }
        }

        var mm string
        for _, mm = range mappingMatches {
          if s.Match == mm || SliceContainsString(mappingOptions[mm].([]string), field){
            fields[mm] = map[string]interface{}{
              "type": defaultType,
              "index": indexTrueValue,
              "analyzer": "searchpick_"+mm+"_index",
            }
          }
        }

        fieldFields := fields[field]
        exceptFields := map[string]interface{}{}

        for fk, fv := range fields {
          if fk != field {
            exceptFields[fk] = fv
          }
        }
        MapMerge(&fieldFields, map[string]interface{}{ "fields": exceptFields })
        mappings[field] = fieldFields 
      }
    }

    if !reflect.ValueOf(s.Locations).IsZero() && len(s.Locations) > 0 {
      for _, l := range s.Locations {
        mappings[l] = map[string]interface{}{
          "type": "geo_point",
        }
      }
    }

    if !reflect.ValueOf(s.GeoShape).IsZero() && len(s.GeoShape) > 0 {
      for _, g := range s.GeoShape {
        mappings[g] = map[string]interface{}{
          "type": "geo_shape",
        }
      }
    }

    if s.Inheritance {
      mappings["type"] = keywordMapping  
    }

    routing := map[string]interface{}{}
    if s.Routing != "" {
      routing["required"] = true
      if s.Routing == "false" {
        routing["path"] = "false"
      }  
    }

    dynamicFields := map[string]interface{}{
      "{name}": keywordMapping,
    }

    if mappingOptions["is_filterable"].(bool) {
      dynamicFields["{name}"] = map[string]interface{}{
        "type": defaultType,
        "index": indexTrueValue,
      }
    }

    if !mappingOptions["is_searchable"].(bool) {
      if s.Match != "" && s.Match != "word" {
        matchAnalyzer := "searchpick_"+s.Match+"_index"
        dynamicFields[s.Match] = map[string]interface{}{
          "type": defaultType,
          "index": indexTrueValue,
          "analyzer": matchAnalyzer,
        }
      }



      if isWord {
        dynamicFields["analyzed"] = analyzedFieldOptions
      }
    }

    dynamicFieldsWithoutName := map[string]interface{}{}
    for dfK, dfV := range dynamicFields {
      log.Println(dfK)
      // dfKString := dfK.(string)
      if dfK != "{name}" {
        dynamicFieldsWithoutName[dfK] = dfV
      }
    }

    nameDynamicFields := dynamicFields["{name}"]
    multiFields := map[string]interface{}{
      "fields": dynamicFieldsWithoutName,
    }
    MapMerge(&multiFields, nameDynamicFields)
    
    dynamicTemplateItem := map[string]interface{}{
      "string_template": map[string]interface{}{
        "match": "*",
        "match_mapping_type": "string",
        "mapping": multiFields,
      },
    }

    if !reflect.ValueOf(s.Mappings).IsZero() && s.MergeMappings {
      MapMerge(&mappings, s.Mappings)
    }

    formattedMappings := map[string]interface{}{
      "properties": mappings,
      "_routing": routing,
      "dynamic_templates": []interface{}{dynamicTemplateItem},
    }

    indexMappings := map[string]interface{}{}
    indexMappings[s.IndexType] = formattedMappings
    settings["analysis"] = map[string]interface{}{
      "analyzer": analyzer,
      "filter": filter,
      "char_filter": charFilter,
    }
    mappings = indexMappings
  }

  return map[string]interface{}{
    "settings": settings,
    "mappings": mappings,
  }

}