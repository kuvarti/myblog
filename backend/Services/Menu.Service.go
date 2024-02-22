package services

import (
	models "backend/Models"
	"context"
	"encoding/json"
	"os"
)

type MenuService interface {
	GetMenu() ([]*models.MenuModel, error)
}

type MenuServiceImplementation struct {
	ctx context.Context
}

func NewMenuService(ctx context.Context) MenuService {
	return &MenuServiceImplementation{
		ctx: ctx,
	}
}

func (msi *MenuServiceImplementation) GetMenu() ([]*models.MenuModel, error) {
	var returnMenu []*models.MenuModel
	infile, err := os.ReadFile("temporaryDB/menu.json");
	if err != nil {
		return nil, err
	}
	json.Unmarshal(infile, &returnMenu);
	return returnMenu, nil
}
