package xmongo

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	mopt "go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// mongodb method names
	aggregate              = "Aggregate"
	bulkWrite              = "BulkWrite"
	countDocuments         = "CountDocuments"
	deleteMany             = "DeleteMany"
	deleteOne              = "DeleteOne"
	distinct               = "Distinct"
	estimatedDocumentCount = "EstimatedDocumentCount"
	find                   = "Find"
	findOne                = "FindOne"
	findOneAndDelete       = "FindOneAndDelete"
	findOneAndReplace      = "FindOneAndReplace"
	findOneAndUpdate       = "FindOneAndUpdate"
	insertMany             = "InsertMany"
	insertOne              = "InsertOne"
	replaceOne             = "ReplaceOne"
	updateByID             = "UpdateByID"
	updateMany             = "UpdateMany"
	updateOne              = "UpdateOne"
	drop                   = "Drop"
)

type Collection struct {
	dbname string
	name   string
	conn   *mongo.Collection
	config *Config
}

func newCollection(dbname string, config *Config, conn *mongo.Collection) *Collection {
	return &Collection{dbname: dbname, name: conn.Name(), conn: conn, config: config}
}

func (c *Collection) Aggregate(ctx context.Context, pipeline interface{},
	opts ...*mopt.AggregateOptions) (cur *mongo.Cursor, err error) {
	// ctx, span := c.startSpan(ctx, aggregate)
	// defer func() {
	// 	c.endSpan(span, err)
	// }()

	cur, err = c.conn.Aggregate(ctx, pipeline, opts...)
	return
}

func (c *Collection) BulkWrite(ctx context.Context, models []mongo.WriteModel,
	opts ...*mopt.BulkWriteOptions) (res *mongo.BulkWriteResult, err error) {
	res, err = c.conn.BulkWrite(ctx, models, opts...)
	return
}

func (c *Collection) CountDocuments(ctx context.Context, filter interface{},
	opts ...*mopt.CountOptions) (count int64, err error) {
	count, err = c.conn.CountDocuments(ctx, filter, opts...)
	return
}

func (c *Collection) DeleteMany(ctx context.Context, filter interface{},
	opts ...*mopt.DeleteOptions) (res *mongo.DeleteResult, err error) {
	res, err = c.conn.DeleteMany(ctx, filter, opts...)
	return
}

func (c *Collection) DeleteOne(ctx context.Context, filter interface{},
	opts ...*mopt.DeleteOptions) (res *mongo.DeleteResult, err error) {
	res, err = c.conn.DeleteOne(ctx, filter, opts...)
	return
}

func (c *Collection) Distinct(ctx context.Context, fieldName string, filter interface{},
	opts ...*mopt.DistinctOptions) (val []interface{}, err error) {
	val, err = c.conn.Distinct(ctx, fieldName, filter, opts...)
	return
}

func (c *Collection) EstimatedDocumentCount(ctx context.Context,
	opts ...*mopt.EstimatedDocumentCountOptions) (val int64, err error) {
	val, err = c.conn.EstimatedDocumentCount(ctx, opts...)
	return
}

func (c *Collection) Find(ctx context.Context, filter interface{},
	opts ...*mopt.FindOptions) (cur *mongo.Cursor, err error) {
	cur, err = c.conn.Find(ctx, filter, opts...)
	return
}

func (c *Collection) FindOne(ctx context.Context, filter interface{},
	opts ...*mopt.FindOneOptions) (res *mongo.SingleResult, err error) {
	res = c.conn.FindOne(ctx, filter, opts...)
	err = res.Err()
	return
}

func (c *Collection) FindOneAndDelete(ctx context.Context, filter interface{},
	opts ...*mopt.FindOneAndDeleteOptions) (res *mongo.SingleResult, err error) {
	res = c.conn.FindOneAndDelete(ctx, filter, opts...)
	err = res.Err()
	return
}

func (c *Collection) FindOneAndReplace(ctx context.Context, filter interface{},
	replacement interface{}, opts ...*mopt.FindOneAndReplaceOptions) (
	res *mongo.SingleResult, err error) {
	res = c.conn.FindOneAndReplace(ctx, filter, replacement, opts...)
	err = res.Err()
	return
}

func (c *Collection) FindOneAndUpdate(ctx context.Context, filter, update interface{},
	opts ...*mopt.FindOneAndUpdateOptions) (res *mongo.SingleResult, err error) {
	res = c.conn.FindOneAndUpdate(ctx, filter, update, opts...)
	err = res.Err()
	return
}

func (c *Collection) InsertMany(ctx context.Context, documents []interface{},
	opts ...*mopt.InsertManyOptions) (res *mongo.InsertManyResult, err error) {
	res, err = c.conn.InsertMany(ctx, documents, opts...)
	return
}

func (c *Collection) InsertOne(ctx context.Context, document interface{},
	opts ...*mopt.InsertOneOptions) (res *mongo.InsertOneResult, err error) {
	res, err = c.conn.InsertOne(ctx, document, opts...)
	return
}

func (c *Collection) ReplaceOne(ctx context.Context, filter, replacement interface{},
	opts ...*mopt.ReplaceOptions) (res *mongo.UpdateResult, err error) {
	res, err = c.conn.ReplaceOne(ctx, filter, replacement, opts...)
	return
}

func (c *Collection) UpdateByID(ctx context.Context, id, update interface{},
	opts ...*mopt.UpdateOptions) (res *mongo.UpdateResult, err error) {
	res, err = c.conn.UpdateByID(ctx, id, update, opts...)
	return
}

func (c *Collection) UpdateMany(ctx context.Context, filter, update interface{},
	opts ...*mopt.UpdateOptions) (res *mongo.UpdateResult, err error) {
	res, err = c.conn.UpdateMany(ctx, filter, update, opts...)
	return
}

func (c *Collection) UpdateOne(ctx context.Context, filter, update interface{},
	opts ...*mopt.UpdateOptions) (res *mongo.UpdateResult, err error) {
	res, err = c.conn.UpdateOne(ctx, filter, update, opts...)
	return
}

func (c *Collection) Drop(ctx context.Context) (err error) {
	return c.conn.Drop(ctx)
}
