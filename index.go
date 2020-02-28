package searchpick

import (
  "context"
  "encoding/json"
  "log"
  "fmt"
)

func (s *Searchpick) BuildSearchData(id string, data interface{}) *SearchData {
  j, _ := json.Marshal(data)
  searchData := &SearchData{
    Id: id,
    BodyJson: string(j),
  }
  return searchData
}

func (s *Searchpick) GetIndexName() string {
  if s.IndexPrefix == "" {
    return s.IndexName
  }

  return s.IndexPrefix + "_" + s.IndexName
}

func (s *Searchpick) UnsetupIndex() *Searchpick {
  ctx := context.Background()
  _, err := s.Client.DeleteIndex(s.GetIndexName()).Do(ctx)

  if err != nil {
    s.Error = err
    log.Println(err)
  }

  return s
}

func (s *Searchpick) SetupIndex() *Searchpick {
  s.Hook()
  if s.Error != nil { return s }
  
  if s.FinalMappings == nil { s.FinalMappings = s.SettingsMappings() }

  if s.Error != nil { return s }

  ctx := context.Background()
  exists, _ := s.Client.IndexExists(s.GetIndexName()).Do(ctx)

  if !exists {
    j, _ := json.Marshal(s.FinalMappings)
    log.Println(string(j))
    createIndex, err := s.Client.CreateIndex(s.GetIndexName()).Body(string(j)).Do(ctx)

    if err != nil {
      s.Error = err
      log.Println(err)
      return s
    }

    if !createIndex.Acknowledged {
      log.Println(s.Error)
    }
  }

  return s
}

func (s *Searchpick) Reindex() error {
  s.SetupIndex()
  if s.Error != nil { return s.Error }
  ctx := context.Background()

  _, err := s.Client.Index().
    Index(s.GetIndexName()).
    Type(s.IndexType).
    Id(fmt.Sprintf("%v", s.SearchData.Id)).
    BodyJson(s.SearchData.BodyJson).
    Do(ctx)

  if err != nil {
    log.Println(err)
  }

  return err
}

func (s *Searchpick) IndexExists(id interface{}) error {
  s.SetupIndex()
  if s.Error != nil { return s.Error }

  ctx := context.Background()
  _, err := s.Client.Get().
    Index(s.GetIndexName()).
    Type(s.IndexType).
    Id(fmt.Sprintf("%v", id)).
    Do(ctx)

  if err != nil {
    log.Println(err)
  }

  return err
}

func (s *Searchpick) IndexDelete(id interface{}) error {
  s.SetupIndex()
  if s.Error != nil { return s.Error }

  ctx := context.Background()
  _, err := s.Client.Delete().
    Index(s.GetIndexName()).
    Type(s.IndexType).
    Id(fmt.Sprintf("%v", id)).
    Do(ctx)

  if err != nil {
    log.Println(err)
  }

  return err
}
