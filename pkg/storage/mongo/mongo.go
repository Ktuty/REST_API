package mongo

import (
	"GoNews/pkg/storage"
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Storage struct {
	db *mongo.Client
}

func NewStorage(uri string) (*Storage, error) {
	mongoOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.Background(), mongoOpts)

	if err != nil {
		return nil, err
	}

	//defer client.Disconnect(context.Background())

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	s := Storage{
		db: client,
	}
	return &s, nil
}

const (
	databaseName   = "pst"
	collectionName = "posts"
	counterName    = "counter"
)

func (s *Storage) Posts() ([]storage.Post, error) {
	collection := s.db.Database(databaseName).Collection(collectionName)
	var posts []storage.Post

	cursor, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	err = cursor.All(context.Background(), &posts)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

// func (s *Storage) AddPost(post storage.Post) error {
// 	collection := s.db.Database(databaseName).Collection(collectionName)
// 	_, err := collection.InsertOne(context.Background(), post)
// 	if err != nil {
// 		return err
// 	}

//		return nil
//	}
func (s *Storage) AddPost(post storage.Post) error {
	id, err := s.getNextSequence(collectionName)
	if err != nil {
		log.Printf("Failed to get next sequence: %v", err)
		return err
	}
	post.ID = id

	// Создание новой коллекции.
	collection := s.db.Database(databaseName).Collection(collectionName)

	// Вставка нового поста.
	_, err = collection.InsertOne(context.Background(), post)
	if err != nil {
		log.Printf("Failed to insert post: %v", err)
		return err
	}

	log.Println("Post inserted successfully")
	return nil
}

type Counter struct {
	ID  string `bson:"_id"`
	Seq int    `bson:"seq"`
}

func (s *Storage) getNextSequence(name string) (int, error) {
	collection := s.db.Database(databaseName).Collection(counterName)
	update := bson.M{
		"$inc": bson.M{
			"seq": 1,
		},
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After).SetUpsert(true)
	var counter Counter
	err := collection.FindOneAndUpdate(context.Background(), bson.M{"_id": name}, update, opts).Decode(&counter)
	if err != nil {
		return 0, err
	}
	return counter.Seq, nil
}
func (s *Storage) UpdatePost(post storage.Post) error {
	collection := s.db.Database(databaseName).Collection(collectionName)
	filter := bson.M{"_id": post.ID}
	update := bson.M{
		"$set": bson.M{
			"author_id":   post.AuthorID,
			"title":       post.Title,
			"content":     post.Content,
			"created_at":  post.CreatedAt,
			"author_name": post.AuthorName,
		},
	}

	_, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) DeletePost(post storage.Post) error {
	collection := s.db.Database(databaseName).Collection(collectionName)
	filter := bson.M{"_id": post.ID}
	_, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return err
	}
	return nil
}
