package services

import (
	models "backend/Models"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MenuService interface {
	GetMenu() ([]*models.MenuModel, error)
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
