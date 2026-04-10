package mongodb

import (
	"context"
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (c *Client) SaveHistoryCommand(history db_models.CliOperationHistory) error {
	_, err := c.col(colOperationHistory).InsertOne(context.Background(), toMOperationHistory(history))
	return err
}

func (c *Client) GetDailyOperationHistory(date time.Time) ([]db_models.CliOperationHistory, error) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.Add(24 * time.Hour)

	filter := bson.M{
		"created_date": bson.M{
			"$gte": start,
			"$lt":  end,
		},
	}
	opts := options.Find().SetSort(bson.D{{Key: "ne_name", Value: 1}, {Key: "created_date", Value: 1}})

	ctx := context.Background()
	cur, err := c.col(colOperationHistory).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var results []db_models.CliOperationHistory
	for cur.Next(ctx) {
		var m mOperationHistory
		if err := cur.Decode(&m); err != nil {
			return nil, err
		}
		results = append(results, fromMOperationHistory(&m))
	}
	return results, cur.Err()
}

func (c *Client) DeleteHistoryBefore(cutoff time.Time) (int64, error) {
	filter := bson.M{"created_date": bson.M{"$lt": cutoff}}
	result, err := c.col(colOperationHistory).DeleteMany(context.Background(), filter)
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}
