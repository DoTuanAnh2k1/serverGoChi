package mongodb

import (
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (c *Client) SaveHistoryCommand(history db_models.CliOperationHistory) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	if history.ID == 0 {
		id, err := c.nextID(ctx, colOperationHistory)
		if err != nil {
			return err
		}
		history.ID = int32(id)
	}
	_, err := c.col(colOperationHistory).InsertOne(ctx, toMOperationHistory(history))
	return err
}

func (c *Client) GetDailyOperationHistory(date time.Time) ([]db_models.CliOperationHistory, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.Add(24 * time.Hour)

	filter := bson.M{
		"created_date": bson.M{
			"$gte": start,
			"$lt":  end,
		},
	}
	opts := options.Find().SetSort(bson.D{{Key: "ne_name", Value: 1}, {Key: "created_date", Value: 1}})

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

func (c *Client) GetRecentHistory(limit int) ([]db_models.CliOperationHistory, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	opts := options.Find().SetSort(bson.D{{Key: "created_date", Value: -1}}).SetLimit(int64(limit))
	cur, err := c.col(colOperationHistory).Find(ctx, bson.M{}, opts)
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

func (c *Client) GetRecentHistoryFiltered(limit int, scope, neName, account string) ([]db_models.CliOperationHistory, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	filter := bson.M{}
	if scope != "" {
		filter["scope"] = scope
	}
	if neName != "" {
		filter["ne_name"] = neName
	}
	if account != "" {
		filter["account"] = account
	}
	opts := options.Find().SetSort(bson.D{{Key: "created_date", Value: -1}}).SetLimit(int64(limit))
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
	ctx, cancel := c.opCtx()
	defer cancel()
	filter := bson.M{"created_date": bson.M{"$lt": cutoff}}
	result, err := c.col(colOperationHistory).DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}
