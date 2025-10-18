package store

import (
	"context"
	"go-yuga-dist/app/model"
)

type Store interface {
	CreateUser(ctx context.Context, user *model.User) error
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	GetUserByID(ctx context.Context, userID string) (*model.User, error)
	CreatePost(ctx context.Context, post *model.Post) (string, error)
	GetPostsByUserID(ctx context.Context, userID string) ([]*model.Post, error)
	GetPostByID(ctx context.Context, postID string) (*model.Post, error)
	CreateSession(ctx context.Context, session *model.Session) error
	GetSessionByToken(ctx context.Context, token string) (*model.Session, error)
}
