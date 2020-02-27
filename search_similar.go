package searchpick

// Find in a model which have similar value in field address and name
// Search with SearchOption:
// s := &searchpick.SearchOption {
//   Term: "*",
//   Similar: true,
//   Fields: []string{"name"},
//   Where: map[string]interface{}{"size": "12 oz"}
// }
// s.Search()
// 
// Search with model:
// user := &models.User{}
// search := &searchpick.SearchOption{
//   Fields: []string{"name"},
//   Where: map[string]interface{}{"size": "12 oz"}  
// }
// results, err := user.Similar(search)
func (sq *SearchQuery) SetSimilar(term string, sf *BoostField) {
  sq.Query["more_like_this"] = map[string]interface{}{
    "like": term,
    "min_doc_freq": 1,
    "min_term_freq": 1,
    "analyzer": "elastiq_search2",
    "fields": sf.Fields,
  }
}
