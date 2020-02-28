# Searchpick {Work in Progress}

Inspired by [Searchkick](https://github.com/ankane/searchkick), make elasticsearch easy in golang. 

## Contents

- [Getting Started](#getting-started)
- [Indexing](#indexing)
- [Querying](#querying)

## Getting Started

[Install Elasticsearch](https://www.elastic.co/guide/en/elasticsearch/reference/current/setup.html). For Homebrew, use:

```sh
brew install elasticsearch
brew services start elasticsearch
```

Add this library to your golang sources:

```sh
go get github.com/jackbit/searchpick
```
Add searchpick to models you want to search.

```golang
package models
import sp "github.com/jackbit/searchpick"

type Product struct {
    ID string `json:"id" db:"id"` // it is used as identifier
    Tags []string `json:"tags" db:"tags"`
    Price float64 `json:"price" db:"price"`
    Color string `json:"color" db:"color"`
    StoreName string `json:"store_name", db:"-"`
    StoreId float64 `json:"store_id" db:"store_id"`
    //...
}

func (p *Product) Searchpick() *sp.Searchpick {
  return &sp.Searchpick{
    Name: "products", // name of elasticsearch index
  }
}
```
## Indexing

Define search data for index.
```golang
func (p *Product) SearchData() *sp.Searchpick {
  data := map[string]interface{} {
    "keyword": fmt.Sprintf("%s %s %s", u.Name, u.Category, u.StoreName),
    "price": p.Price,
    "tags": p.Tags,
    "store_id": p.StoreId,
    "color": p.Color,
    "created_at": u.CreatedAt,
    "updated_at": u.UpdatedAt,
  }
  searchPick := p.Searchpick()
  searchPick.SearchData = searchPick.BuildSearchData(u.ID.String(), data)
  return searchPick
}
```

Add data to the search index per record

```ruby
product := &Product{}
if err := product.SearchData().Reindex(); err != nil {
    log.Println("Reindex is failed")
}
```

Remove data to from index.
```ruby
product := &Product{}
if err := product.Searchpick().IndexDelete(product.ID); err != nil {
    log.Println("Product can not be deleted")
}
```
Check data if indexed.
```ruby
product := &Product{}
if err := product.Searchpick().IndexExists(product.ID); err != nil {
    log.Println("Product is not exist")
}
```

## Querying

Query like SQL

```ruby
options := SearchOption{
    Term: "*",
    Where: map[string]interface{}{
        "expired_at": map[string]interface{}{"gt": time.Now()},
        "category_id": []interface{}{ 25, 30 },
        "size": map[string]interface{}{
            "all": []interface{}{ 
                "s", 
                "m", 
                "l", 
                "xl",
            },
        },
        "title": map[string]interface{}{
            "like": "%converse%",
        },
        "tag": map[string]interface{}{
            "regexp": "/shoe .+/",
        },
        "supplier_id": map[string]interface{}{
            "exists": true,
        },
        "payment_method": map[string]interface{}{
            "not": []interface{}{ 
                "cod", 
                "cash",
            },
        },
        "_or": []interface{}{
          map[string]interface{}{
              "in_stock": true,
              "backordered": true,
          },
        },
        "_and": []interface{}{
          map[string]interface{}{
              "city_id": []interface{}{ 1, 2, 3, 4},
          },
          map[string]interface{}{
              "shipping_method_id": []interface{}{ 10, 20, 30, 40},
          },
        },
        "_not": map[string]interface{}{
            "store_id": 1001,
        },
    },
    Page: 1,
    PerPage: 20,
    Order: map[string]interface{}{
        "price": "asc",
    }
}
response, err := product.Search(options)
if err != nil {
    log.Println(err.Error())
}
log.Println(response)
```

Query like JSON

```ruby
query := `{
    "expired_at": {"gt": "2020-12-24"},
    "category_id": [25, 30],
    "size": {"all": ["s", "m", "l", "xl"]},
    "title": {"like": "%converse%"},
    "tag": {"regexp": "/shoe .+/"},
    "supplier_id": {"exists": true},
    "payment_method": {"not": ["cod", "cash"]},
    "_or": [ {"in_stock": true}, {"backordered": true} ],
    "_and": [ {"city_id": [1, 2, 3, 4]}, {"shipping_method_id": [10, 20, 30, 40]} ],
    "_not": {"store_id": 1}
}`
jsonQuery, _ := json.Marshal(query)
whereQuery := jsonQuery.( map[string]interface{} )
options := SearchOption{
    Term: "*",
    Where: whereQuery,
    Page: 1,
    PerPage: 20,
    Order: map[string]interface{}{ "price": "asc" }
}
response, err := product.Search(options)
if err != nil {
    log.Println(err.Error())
}
log.Println(response)
```
Available options for Searching:
```golang
type SearchOption struct {
  Term     string
  Fields   []string
  Operator string
  Page     int64
  PerPage  int64
  Limit    int64
  Padding  int64
  Offset   int64
  Order    map[string]interface{}
  Where    map[string]interface{}
  Similar  bool
  Match    string
}
```

Order / Sort
[All of these sort options are supported](https://www.elastic.co/guide/en/elasticsearch/reference/current/search-request-sort.html). The most relevant first is the default

```ruby
Order: map[string]interface{}{"_score": "desc"} 
```
Limit / offset

```ruby
Limit: 20, 
Offset: 40,
```

Select

```ruby
Select: []string{"name"}
```
