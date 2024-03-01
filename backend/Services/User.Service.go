package services

import (
	models "backend/Models"
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Login(string, string) error
}

type UserServiceImplementation struct{
	collection *mongo.Collection
	ctx context.Context
}

func NewUserService(ctx context.Context, col *mongo.Collection) (UserService, TokenService) {
	return &UserServiceImplementation{
		collection: col,
		ctx: ctx,
	}, &TokenServiceImplementation{
		ctx: ctx,
	}
}

func (usi *UserServiceImplementation) Login(uName string, uPass string) error {
	var User models.UserModel
	querry := bson.D{bson.E{Key: "Username", Value: uName}};
	if err := usi.collection.FindOne(usi.ctx, querry).Decode(&User); err != nil {
		return errors.New("invalid Username and Password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(User.Password), []byte(uPass)); err != nil {
		return errors.New("invalid Username and Password")
	}
	return nil
}
