package services

import (
	models "backend/Models"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MenuService interface {
	GetMenu() ([]*models.MenuModel, error)
	Upsert(m models.MenuModel) error
	DeleteByPageName(pageName string) error
	GetByPageName(pageName string) (models.MenuModel, error)
}

type MenuServiceImplementation struct {
	collection *mongo.Collection
	ctx context.Context
}

func NewMenuService(ctx context.Context, col *mongo.Collection) MenuService {
	return &MenuServiceImplementation{
		collection: col,
		ctx: ctx,
	}
}

func (msi *MenuServiceImplementation) GetMenu() ([]*models.MenuModel, error) {
	var returnMenu []*models.MenuModel
	cursor, err := msi.collection.Find(msi.ctx, bson.D{{}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(msi.ctx)

	if err := cursor.All(msi.ctx, &returnMenu); err != nil {
		return nil, err
	}
	return returnMenu, nil
}

// Upsert writes the display metadata for a page's menu entry. The page owns its
// Path now (see PageService), so the menu document carries only Name/Caption.
func (msi *MenuServiceImplementation) Upsert(m models.MenuModel) error {
	if m.Name == "" {
		m.Name = m.PageName
	}
	opts := options.Update().SetUpsert(true)
	_, err := msi.collection.UpdateOne(msi.ctx,
		bson.D{{Key: "PageName", Value: m.PageName}},
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "Name", Value: m.Name},
			{Key: "Caption", Value: m.Caption},
			{Key: "PageName", Value: m.PageName},
		}}},
		opts,
	)
	return err
}

func (msi *MenuServiceImplementation) DeleteByPageName(pageName string) error {
	_, err := msi.collection.DeleteMany(msi.ctx, bson.D{{Key: "PageName", Value: pageName}})
	return err
}

func (msi *MenuServiceImplementation) GetByPageName(pageName string) (models.MenuModel, error) {
	var m models.MenuModel
	err := msi.collection.FindOne(msi.ctx, bson.D{{Key: "PageName", Value: pageName}}).Decode(&m)
	return m, err
}
