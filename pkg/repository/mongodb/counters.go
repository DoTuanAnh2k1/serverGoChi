package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const colCounters = "cli_counters"

// nextID returns an atomically-incremented int64 id for the given name by
// issuing findAndModify($inc seq) on the counters collection. It replaces the
// auto-increment behavior that MySQL/Postgres give for free.
func (c *Client) nextID(ctx context.Context, name string) (int64, error) {
	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)
	var result struct {
		Seq int64 `bson:"seq"`
	}
	err := c.col(colCounters).FindOneAndUpdate(ctx,
		bson.M{"_id": name},
		bson.M{"$inc": bson.M{"seq": int64(1)}},
		opts,
	).Decode(&result)
	if err != nil {
		return 0, fmt.Errorf("nextID(%s): %w", name, err)
	}
	return result.Seq, nil
}
