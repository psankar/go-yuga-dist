package store

import (
	"context"
	"go-yuga-dist/app/model"

	"github.com/jackc/pgx/v4/pgxpool"
)

type DB struct {
	pool *pgxpool.Pool
}

func NewDB(pool *pgxpool.Pool) *DB {
	return &DB{pool: pool}
}

func (db *DB) CreateUser(ctx context.Context, user *model.User) error {
	_, err := db.pool.Exec(ctx,
		"INSERT INTO users (name, email_address, password, region) VALUES ($1, $2, $3, $4)",
		user.Name, user.EmailAddress, user.Password, user.Region)
	return err
}

func (db *DB) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	user := &model.User{}
	err := db.pool.QueryRow(ctx,
		"SELECT id, name, email_address, password, region, created_at FROM users WHERE email_address = $1",
		email).Scan(&user.ID, &user.Name, &user.EmailAddress, &user.Password, &user.Region, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (db *DB) GetUserByID(ctx context.Context, userID string) (*model.User, error) {
	user := &model.User{}
	err := db.pool.QueryRow(ctx,
		"SELECT id, name, email_address, password, region, created_at FROM users WHERE id = $1",
		userID).Scan(&user.ID, &user.Name, &user.EmailAddress, &user.Password, &user.Region, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (db *DB) CreatePost(ctx context.Context, post *model.Post) (string, error) {
	var postID string
	err := db.pool.QueryRow(ctx,
		"INSERT INTO posts (user_id, content, region) VALUES ($1, $2, $3) RETURNING id",
		post.UserID, post.Content, post.Region).Scan(&postID)
	if err != nil {
		return "", err
	}
	return postID, nil
}

func (db *DB) GetPostsByUserID(ctx context.Context, userID string) ([]*model.Post, error) {
	rows, err := db.pool.Query(ctx,
		"SELECT id, user_id, content, region, created_at FROM posts WHERE user_id = $1",
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*model.Post
	for rows.Next() {
		post := &model.Post{}
		err := rows.Scan(&post.ID, &post.UserID, &post.Content, &post.Region, &post.CreatedAt)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, nil
}

func (db *DB) GetPostByID(ctx context.Context, postID string) (*model.Post, error) {
	post := &model.Post{}
	var authorName string
	err := db.pool.QueryRow(ctx,
		"SELECT p.id, p.user_id, p.content, p.region, p.created_at, u.name FROM posts p JOIN users u ON p.user_id = u.id WHERE p.id = $1",
		postID).Scan(&post.ID, &post.UserID, &post.Content, &post.Region, &post.CreatedAt, &authorName)
	if err != nil {
		return nil, err
	}
	post.AuthorName = authorName
	return post, nil
}

func (db *DB) CreateSession(ctx context.Context, session *model.Session) error {
	_, err := db.pool.Exec(ctx,
		"INSERT INTO sessions (token, user_id, expires_at) VALUES ($1, $2, $3)",
		session.Token, session.UserID, session.ExpiresAt)
	return err
}

func (db *DB) GetSessionByToken(ctx context.Context, token string) (*model.Session, error) {
	session := &model.Session{}
	err := db.pool.QueryRow(ctx,
		"SELECT token, user_id, expires_at FROM sessions WHERE token = $1",
		token).Scan(&session.Token, &session.UserID, &session.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return session, nil
}
