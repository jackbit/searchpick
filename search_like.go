package searchpick

import (
  "strings"
  "regexp"
)

// user.Searchpick().Search(&sp.SearchOption{
//   Where: map[string]interface{}{
//     "category": map[string]interface{}{
//       "like": "%frozen%",
//     },
//   },
// })
func (sFilter *SearchFilter) CheckQueryLike(sq *SearchQuery) bool {

  if sq.Operator != "like" {  return false }

  reserved := []string{".", "?", "+", "*", "|", "{", "}", "[", "]", "(", ")", "\"", "\\"}
  text := sq.OperatorQuery.(string)

  for _, reserve := range reserved {
    text = strings.ReplaceAll(text, reserve, "\\"+reserve)
  }

  r1, _ := regexp.Compile("%")
  r2, _ := regexp.Compile("_")

  text = strings.ReplaceAll(r1.ReplaceAllString(text, ".*"),"\\.*", "\\%")
  text = strings.ReplaceAll(r2.ReplaceAllString(text, ".*"),"\\.*", "\\_")
  text = strings.ReplaceAll(text, "\\\\%", "%")
  text = strings.ReplaceAll(text, "\\\\_", "_")

  expQuery := map[string]interface{}{}
  expQuery[sq.Field] = map[string]interface{}{
    "value": text,
  }

  sFilter.Filters = append(sFilter.Filters, map[string]interface{}{ "regexp": expQuery})
  return true
}