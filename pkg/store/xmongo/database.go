package xmongo

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	mopt "go.mongodb.org/mongo-driver/mongo/options"
)

type Database struct {
	name   string
	db     *mongo.Database
	config *Config
}

func newDatabase(name string, config *Config, db *mongo.Database) *Database {
	return &Database{name: name, config: config, db: db}
}

func (s *Database) Aggregate(ctx context.Context, coll string, pipeline interface{}, opts ...*mopt.AggregateOptions) (cur *mongo.Cursor, err error) {
	cur, err = newCollection(s.name, s.config, s.db.Collection(coll)).Aggregate(ctx, pipeline, opts...)
	return
}

func (s *Database) BulkWrite(ctx context.Context, coll string, models []mongo.WriteModel, opts ...*mopt.BulkWriteOptions) (res *mongo.BulkWriteResult, err error) {
	res, err = newCollection(s.name, s.config, s.db.Collection(coll)).BulkWrite(ctx, models, opts...)
	return
}

func (s *Database) CountDocuments(ctx context.Context, coll string, filter interface{}, opts ...*mopt.CountOptions) (count int64, err error) {
	count, err = newCollection(s.name, s.config, s.db.Collection(coll)).CountDocuments(ctx, filter, opts...)
	return
}

func (s *Database) DeleteMany(ctx context.Context, coll string, filter interface{}, opts ...*mopt.DeleteOptions) (res *mongo.DeleteResult, err error) {
	res, err = newCollection(s.name, s.config, s.db.Collection(coll)).DeleteMany(ctx, filter, opts...)
	return
}

func (s *Database) DeleteOne(ctx context.Context, coll string, filter interface{}, opts ...*mopt.DeleteOptions) (res *mongo.DeleteResult, err error) {
	res, err = newCollection(s.name, s.config, s.db.Collection(coll)).DeleteOne(ctx, filter, opts...)
	return
}

func (s *Database) Distinct(ctx context.Context, coll string, fieldName string, filter interface{}, opts ...*mopt.DistinctOptions) (val []interface{}, err error) {
	val, err = newCollection(s.name, s.config, s.db.Collection(coll)).Distinct(ctx, fieldName, filter, opts...)
	return
}

func (s *Database) EstimatedDocumentCount(ctx context.Context, coll string, opts ...*mopt.EstimatedDocumentCountOptions) (val int64, err error) {
	val, err = newCollection(s.name, s.config, s.db.Collection(coll)).EstimatedDocumentCount(ctx, opts...)
	return
}

func (s *Database) Find(ctx context.Context, coll string, filter interface{}, opts ...*mopt.FindOptions) (cur *mongo.Cursor, err error) {
	cur, err = newCollection(s.name, s.config, s.db.Collection(coll)).Find(ctx, filter, opts...)
	return
}

func (s *Database) FindOne(ctx context.Context, coll string, filter interface{}, opts ...*mopt.FindOneOptions) (res *mongo.SingleResult, err error) {
	res, err = newCollection(s.name, s.config, s.db.Collection(coll)).FindOne(ctx, filter, opts...)
	return
}

func (s *Database) FindOneAndDelete(ctx context.Context, coll string, filter interface{}, opts ...*mopt.FindOneAndDeleteOptions) (res *mongo.SingleResult, err error) {
	res, err = newCollection(s.name, s.config, s.db.Collection(coll)).FindOneAndDelete(ctx, filter, opts...)
	return
}

func (s *Database) FindOneAndReplace(ctx context.Context, coll string, filter interface{}, replacement interface{}, opts ...*mopt.FindOneAndReplaceOptions) (res *mongo.SingleResult, err error) {
	res, err = newCollection(s.name, s.config, s.db.Collection(coll)).FindOneAndReplace(ctx, filter, replacement, opts...)
	return
}

func (s *Database) FindOneAndUpdate(ctx context.Context, coll string, filter, update interface{},
	opts ...*mopt.FindOneAndUpdateOptions) (res *mongo.SingleResult, err error) {
	res, err = newCollection(s.name, s.config, s.db.Collection(coll)).FindOneAndUpdate(ctx, filter, update, opts...)
	return
}

func (s *Database) InsertMany(ctx context.Context, coll string, documents []interface{},
	opts ...*mopt.InsertManyOptions) (res *mongo.InsertManyResult, err error) {
	res, err = newCollection(s.name, s.config, s.db.Collection(coll)).InsertMany(ctx, documents, opts...)
	return
}

func (s *Database) InsertOne(ctx context.Context, coll string, document interface{},
	opts ...*mopt.InsertOneOptions) (res *mongo.InsertOneResult, err error) {
	res, err = newCollection(s.name, s.config, s.db.Collection(coll)).InsertOne(ctx, document, opts...)
	return
}

func (s *Database) ReplaceOne(ctx context.Context, coll string, filter, replacement interface{},
	opts ...*mopt.ReplaceOptions) (res *mongo.UpdateResult, err error) {
	res, err = newCollection(s.name, s.config, s.db.Collection(coll)).ReplaceOne(ctx, filter, replacement, opts...)
	return
}

func (s *Database) UpdateByID(ctx context.Context, coll string, id, update interface{},
	opts ...*mopt.UpdateOptions) (res *mongo.UpdateResult, err error) {
	res, err = newCollection(s.name, s.config, s.db.Collection(coll)).UpdateByID(ctx, id, update, opts...)
	return
}

func (s *Database) UpdateMany(ctx context.Context, coll string, filter, update interface{},
	opts ...*mopt.UpdateOptions) (res *mongo.UpdateResult, err error) {
	res, err = newCollection(s.name, s.config, s.db.Collection(coll)).UpdateMany(ctx, filter, update, opts...)
	return
}

func (s *Database) UpdateOne(ctx context.Context, coll string, filter, update interface{},
	opts ...*mopt.UpdateOptions) (res *mongo.UpdateResult, err error) {
	res, err = newCollection(s.name, s.config, s.db.Collection(coll)).UpdateOne(ctx, filter, update, opts...)
	return
}

func (s *Database) Drop(ctx context.Context, coll string) (err error) {
	return newCollection(s.name, s.config, s.db.Collection(coll)).Drop(ctx)
}
