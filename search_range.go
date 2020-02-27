package searchpick

func (sFilter *SearchFilter) AddRange(sq *SearchQuery) {
  hasRange := false
  for _, item := range sFilter.Filters {
    newItem := item.(map[string]interface{})
    _, ok := newItem["range"]
    if ok {
      hasRange = true
      rangeObj := newItem["range"].(map[string]interface{})
      rangeObj[sq.Field] = sq.Query
      break
    }
  }

  if !hasRange {
    newRange := map[string]interface{}{}
    newRange[sq.Field] = sq.Query
    sFilter.Filters = append(sFilter.Filters, map[string]interface{}{"range":  newRange})
  }
}

// user.Searchpick().Search(&sp.SearchOption{
//   Where: map[string]interface{}{
//     "price": map[string]interface{}{
//       "gt": 100, 
//       "lt": 200,
//     },
//   },
// })
func (sFilter *SearchFilter) CheckQueryRange(sq *SearchQuery) bool {
  rangeQuery := map[string]interface{}{ "from": sq.OperatorQuery }
  isRange := true

  switch sq.Operator {
    case "gt":
      rangeQuery["include_lower"] = false
    case "gte":
      rangeQuery["include_lower"] = true
    case "lt":
      rangeQuery["include_upper"] = false
    case "lte":
      rangeQuery["include_upper"] = true
    default:
      isRange = false
  }

  if isRange {
    sq.Query = rangeQuery
    sFilter.AddRange(sq)
  }

  return isRange
}