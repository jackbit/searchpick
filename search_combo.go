package searchpick

// result := user.Searchpick().Search(&sp.SearchOption{
//   Where: map[string]interface{}{
//     "category": map[string]interface{}{
//       "prefix": "frozen",
//     },
//   },
// })
func (sFilter *SearchFilter) CheckQueryPrefix(sq *SearchQuery) bool {
  if sq.Operator != "prefix" { return false }
  filters := map[string]interface{}{}
  filters[sq.Field] = sq.OperatorQuery

  sFilter.Filters = append(sFilter.Filters, map[string]interface{}{"prefix": filters})
  return true
}

// result := user.Searchpick().Search(&sp.SearchOption{
//   Where: map[string]interface{}{
//     "category": map[string]interface{}{
//       "regexp": "/frozen .+/",
//     },
//   },
// })
func (sFilter *SearchFilter) CheckQueryRegex(sq *SearchQuery) bool {
  if sq.Operator != "regexp" { return false }
  filters := map[string]interface{}{}
  filters[sq.Field] = map[string]interface{}{
    "value": sq.OperatorQuery,
  }

  sFilter.Filters = append(sFilter.Filters, map[string]interface{}{"regexp": filters})
  return true
}

// result := user.Searchpick().Search(&sp.SearchOption{
//   Where: map[string]interface{}{
//     "category": map[string]interface{}{
//       "not": []interface{}{ 25, 40 },
//     },
//   },
// })
//
// result := user.Searchpick().Search(&sp.SearchOption{
//   Where: map[string]interface{}{
//     "category": map[string]interface{}{
//       "not": 20,
//     },
//   },
// })
func (sFilter *SearchFilter) CheckQueryNot(sq *SearchQuery) bool {
  if sq.Operator != "not" && sq.Operator != "_not" { return false }

  sFilter.Filters = append(sFilter.Filters, map[string]interface{}{
    "bool": map[string]interface{}{
      "must_not": sFilter.OperatorFilters( sq.Field, sq.OperatorQuery ).Where,
    },
  })

  return true
}

// result := user.Searchpick().Search(&sp.SearchOption{
//   Where: map[string]interface{}{
//     "category": map[string]interface{}{
//       "all": []interface{}{ 25, 40 },
//     },
//   },
// })
func (sFilter *SearchFilter) CheckQueryAll(sq *SearchQuery) bool {
  if sq.Operator != "all" { return false }
  collections := sq.OperatorQuery.([]interface{})

  for _, v := range collections {
    sFilter.Filters = append(sFilter.Filters, sFilter.OperatorFilters(sq.Field, v).Where)
  }
  
  return true
}

// result := user.Searchpick().Search(&sp.SearchOption{
//   Where: map[string]interface{}{
//     "category": map[string]interface{}{
//       "exists": true,
//     },
//   },
// })
func (sFilter *SearchFilter) CheckQueryExists(sq *SearchQuery) bool {
  if sq.Operator != "exists" { return false }

  query := map[string]interface{}{
    "exists": map[string]interface{}{ "field": sq.Field },
  }

  sFilter.Filters = append(sFilter.Filters, query)
  return true
}