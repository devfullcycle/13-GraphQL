package dataloader

import (
	"context"
	"net/http"

	"github.com/LucianTavares/comunicacao_entre_sistemas/graphql/graph/model"
	"github.com/LucianTavares/comunicacao_entre_sistemas/graphql/internal/database"
	"github.com/graph-gophers/dataloader"
	gopher_dataloader "github.com/graph-gophers/dataloader"
)

type ctxKey string

const (
	loadersKey = ctxKey("dataloaders")
)

type DataLoader struct {
	categoryLoader *dataloader.Loader
}

type categoryBatcher struct {
	category *database.Category
}

func (i *DataLoader) GetCategory(ctx context.Context, categoryID string) (*model.Category, error) {

	thunk := i.categoryLoader.Load(ctx, gopher_dataloader.StringKey(categoryID))
	result, err := thunk()

	if err != nil {
		return nil, err
	}
	return result.(*model.Category), nil
}

func NewDataLoader(categoryModel *database.Category) *DataLoader {
	categories := &categoryBatcher{category: categoryModel}
	return &DataLoader{
		categoryLoader: dataloader.NewBatchedLoader(categories.get),
	}
}

func Middleware(loader *DataLoader, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCtx := context.WithValue(r.Context(), loadersKey, loader)
		r = r.WithContext(nextCtx)
		next.ServeHTTP(w, r)
	})
}

func For(ctx context.Context) *DataLoader {
	return ctx.Value(loadersKey).(*DataLoader)
}

func (c *categoryBatcher) get(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {

	categoryIDs := make([]string, len(keys))
	for ix, key := range keys {
		categoryIDs[ix] = key.String()
	}

	categories, err := c.category.FindByIds(categoryIDs)

	if err != nil {
		return []*dataloader.Result{{Data: nil, Error: err}}
	}
	if len(categories) != len(keys) {
		return []*dataloader.Result{{Data: nil, Error: nil}}
	}

	var categoriesModel []*model.Category
	for _, category := range categories {

		categoriesModel = append(categoriesModel, &model.Category{
			ID:          category.ID,
			Name:        category.Name,
			Description: category.Description,
		})
	}

	results := make([]*dataloader.Result, len(keys))

	for index, categoryId := range keys {
		for _, category := range categoriesModel {
			if categoryId.String() == category.ID {
				results[index] = &dataloader.Result{Data: category, Error: nil}
			}
		}
	}

	return results
}
