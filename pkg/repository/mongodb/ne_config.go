package mongodb

import (
	"errors"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// neConfigFromMNe derives a CliNeConfig DTO from an mNe document.
func neConfigFromMNe(m *mNe) *db_models.CliNeConfig {
	return &db_models.CliNeConfig{
		ID:          m.ID,
		NeID:        m.ID,
		IPAddress:   m.ConfMasterIP,
		Port:        m.ConfPortMasterSSH,
		Username:    m.ConfUsername,
		Password:    m.ConfPassword,
		Protocol:    m.ConfMode,
		Description: m.Description,
	}
}

// CreateCliNeConfig writes connection config into the CliNe document's conf_* fields.
func (c *Client) CreateCliNeConfig(cfg *db_models.CliNeConfig) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	filter := bson.M{"id": cfg.NeID}
	update := bson.M{"$set": bson.M{
		"conf_master_ip":       cfg.IPAddress,
		"conf_port_master_ssh": cfg.Port,
		"conf_username":        cfg.Username,
		"conf_password":        cfg.Password,
		"conf_mode":            cfg.Protocol,
		"description":          cfg.Description,
	}}
	res, err := c.col(colNe).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("NE not found")
	}
	return nil
}

// GetCliNeConfigByNeId returns a single-element slice with config derived from CliNe.
func (c *Client) GetCliNeConfigByNeId(neId int64) ([]*db_models.CliNeConfig, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	var m mNe
	err := c.col(colNe).FindOne(ctx, bson.M{"id": neId}).Decode(&m)
	if err == mongo.ErrNoDocuments {
		return []*db_models.CliNeConfig{}, nil
	}
	if err != nil {
		return nil, err
	}
	return []*db_models.CliNeConfig{neConfigFromMNe(&m)}, nil
}

// GetCliNeConfigById treats id as the NE id (one config per NE).
func (c *Client) GetCliNeConfigById(id int64) (*db_models.CliNeConfig, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	var m mNe
	err := c.col(colNe).FindOne(ctx, bson.M{"id": id}).Decode(&m)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return neConfigFromMNe(&m), nil
}

// UpdateCliNeConfig updates CliNe's conf_* fields using NeID (falls back to ID).
func (c *Client) UpdateCliNeConfig(cfg *db_models.CliNeConfig) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	neId := cfg.NeID
	if neId == 0 {
		neId = cfg.ID
	}
	filter := bson.M{"id": neId}
	update := bson.M{"$set": bson.M{
		"conf_master_ip":       cfg.IPAddress,
		"conf_port_master_ssh": cfg.Port,
		"conf_username":        cfg.Username,
		"conf_password":        cfg.Password,
		"conf_mode":            cfg.Protocol,
		"description":          cfg.Description,
	}}
	_, err := c.col(colNe).UpdateOne(ctx, filter, update)
	return err
}

// DeleteCliNeConfigById clears the conf_* fields of the CliNe with the given id.
func (c *Client) DeleteCliNeConfigById(id int64) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	filter := bson.M{"id": id}
	update := bson.M{"$set": bson.M{
		"conf_master_ip":       "",
		"conf_port_master_ssh": 0,
		"conf_username":        "",
		"conf_password":        "",
		"conf_mode":            "",
	}}
	_, err := c.col(colNe).UpdateOne(ctx, filter, update)
	return err
}

// DeleteCliNeConfigByNeId clears the conf_* fields of the CliNe with the given NE id.
func (c *Client) DeleteCliNeConfigByNeId(neId int64) error {
	return c.DeleteCliNeConfigById(neId)
}

// cascade helpers

func (c *Client) DeleteAllUserNeMappingByNeId(neId int64) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colUserNeMapping).DeleteMany(ctx, bson.M{"tbl_ne_id": neId})
	return err
}

// DeleteNeMonitorByNeId is a no-op: monitor data is derived from CliNe, no collection to delete from.
func (c *Client) DeleteNeMonitorByNeId(neId int64) error {
	return nil
}

// DeleteCliNeSlaveByNeId is a no-op: cli_ne_slave collection no longer exists.
func (c *Client) DeleteCliNeSlaveByNeId(neId int64) error {
	return nil
}
