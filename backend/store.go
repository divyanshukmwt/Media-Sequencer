package main

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ErrNotFound = errors.New("not found")

type Store interface {
	ListWindows(ctx context.Context) ([]Window, error)
	GetWindow(ctx context.Context, id string) (*Window, error)
	CreateWindow(ctx context.Context, w Window) error
	AddMedia(ctx context.Context, windowID string, item MediaItem) (*Window, error)
	RemoveMedia(ctx context.Context, windowID string, mediaID string) (*Window, error)
	RenameWindow(ctx context.Context, windowID string, name string) (*Window, error)
	DeleteWindow(ctx context.Context, windowID string) error
	FindMediaByID(ctx context.Context, mediaID string) (*MediaItem, error)
	GetSyncState(ctx context.Context) (SyncState, error)
	SetSyncState(ctx context.Context, s SyncState) error
	CountWindows(ctx context.Context) (int64, error)
}

type MongoStore struct {
	windows *mongo.Collection
	sync    *mongo.Collection
}

func ConnectMongo(ctx context.Context, uri string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}
	return client, nil
}

func NewMongoStore(client *mongo.Client, dbName string) *MongoStore {
	db := client.Database(dbName)
	return &MongoStore{
		windows: db.Collection("windows"),
		sync:    db.Collection("sync_state"),
	}
}

func (s *MongoStore) ListWindows(ctx context.Context) ([]Window, error) {
	cur, err := s.windows.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	windows := []Window{}
	for cur.Next(ctx) {
		var w Window
		if err := cur.Decode(&w); err != nil {
			return nil, err
		}
		windows = append(windows, w)
	}
	return windows, cur.Err()
}

func (s *MongoStore) GetWindow(ctx context.Context, id string) (*Window, error) {
	var w Window
	err := s.windows.FindOne(ctx, bson.M{"_id": id}).Decode(&w)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (s *MongoStore) CreateWindow(ctx context.Context, w Window) error {
	now := time.Now().UTC()
	w.CreatedAt = now
	w.UpdatedAt = now
	if w.CycleStartAt.IsZero() {
		w.CycleStartAt = now
	}
	if w.Playlist == nil {
		w.Playlist = []MediaItem{}
	}
	_, err := s.windows.InsertOne(ctx, w)
	return err
}

func (s *MongoStore) AddMedia(ctx context.Context, windowID string, item MediaItem) (*Window, error) {
	filter := bson.M{"_id": windowID}
	update := bson.M{
		"$push": bson.M{"playlist": item},
		"$set":  bson.M{"updated_at": time.Now().UTC()},
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var w Window
	err := s.windows.FindOneAndUpdate(ctx, filter, update, opts).Decode(&w)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (s *MongoStore) RemoveMedia(ctx context.Context, windowID string, mediaID string) (*Window, error) {
	filter := bson.M{"_id": windowID}
	update := bson.M{
		"$pull": bson.M{"playlist": bson.M{"id": mediaID}},
		"$set":  bson.M{"updated_at": time.Now().UTC()},
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var w Window
	err := s.windows.FindOneAndUpdate(ctx, filter, update, opts).Decode(&w)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (s *MongoStore) RenameWindow(ctx context.Context, windowID string, name string) (*Window, error) {
	filter := bson.M{"_id": windowID}
	update := bson.M{
		"$set": bson.M{"name": name, "updated_at": time.Now().UTC()},
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var w Window
	err := s.windows.FindOneAndUpdate(ctx, filter, update, opts).Decode(&w)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (s *MongoStore) DeleteWindow(ctx context.Context, windowID string) error {
	res, err := s.windows.DeleteOne(ctx, bson.M{"_id": windowID})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *MongoStore) FindMediaByID(ctx context.Context, mediaID string) (*MediaItem, error) {
	var w Window
	err := s.windows.FindOne(ctx, bson.M{"playlist.id": mediaID}).Decode(&w)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	for _, m := range w.Playlist {
		if m.ID == mediaID {
			return &m, nil
		}
	}
	return nil, ErrNotFound
}

func (s *MongoStore) GetSyncState(ctx context.Context) (SyncState, error) {
	var state SyncState
	err := s.sync.FindOne(ctx, bson.M{"_id": SyncStateDocID}).Decode(&state)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return SyncState{ID: SyncStateDocID, Active: false}, nil
	}
	if err != nil {
		return SyncState{}, err
	}
	return state, nil
}

func (s *MongoStore) SetSyncState(ctx context.Context, state SyncState) error {
	state.ID = SyncStateDocID
	opts := options.Replace().SetUpsert(true)
	_, err := s.sync.ReplaceOne(ctx, bson.M{"_id": SyncStateDocID}, state, opts)
	return err
}

func (s *MongoStore) CountWindows(ctx context.Context) (int64, error) {
	return s.windows.CountDocuments(ctx, bson.M{})
}