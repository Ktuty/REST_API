package postgres

import (
	"GoNews/pkg/storage"
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

func NewStorage(connstr string) (*Storage, error) {
	db, err := pgxpool.Connect(context.Background(), connstr)
	if err != nil {
		return nil, err
	}

	s := Storage{
		db: db,
	}
	return &s, nil
}

func (s *Storage) Posts() ([]storage.Post, error) {
	rows, err := s.db.Query(context.Background(), `
	SELECT
		posts.id,
		posts.author_id,
		posts.title,
		posts.content,
		posts.created_at,
		authors.name AS author_name
	FROM
		posts
	JOIN
		authors ON posts.author_id = authors.id
	ORDER BY
		posts.id;
	`)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var posts []storage.Post

	for rows.Next() {
		var post storage.Post

		err = rows.Scan(
			&post.ID,
			&post.AuthorID,
			&post.Title,
			&post.Content,
			&post.CreatedAt,
			&post.AuthorName,
		)

		if err != nil {
			return nil, err
		}

		posts = append(posts, post)
	}

	return posts, rows.Err()
}

func (s *Storage) AddPost(post storage.Post) error {
	// Вставка автора
	_, err := s.db.Exec(context.Background(), `
	INSERT INTO
		authors (id, name)
	VALUES
		($1, $2)
	ON CONFLICT (id) DO NOTHING;
	`, post.AuthorID, post.AuthorName)

	if err != nil {
		return err
	}

	// Вставка поста с использованием последовательности
	_, err = s.db.Exec(context.Background(), `
	INSERT INTO
		posts (id, author_id, title, content, created_at)
	VALUES
		(nextval('posts_id_seq'), $1, $2, $3, $4);
	`, post.AuthorID, post.Title, post.Content, post.CreatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) UpdatePost(newPost storage.Post) error {
	_, err := s.db.Exec(context.Background(), `
	UPDATE posts
	SET title=$1, content=$2
	WHERE id=$3;
	`, newPost.Title, newPost.Content, newPost.ID)

	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) DeletePost(post storage.Post) error {
	_, err := s.db.Exec(context.Background(), `
	DELETE FROM posts
	WHERE id = $1;
	`, post.ID)

	if err != nil {
		return err
	}

	return nil
}
