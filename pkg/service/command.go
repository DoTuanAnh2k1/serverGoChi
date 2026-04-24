package service

import (
	"errors"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
)

var (
	ErrCommandNotFound = errors.New("command: not found")
	ErrCommandExists   = errors.New("command: already registered for this NE/service")
	ErrCommandService  = errors.New("command: service must be ne-config or ne-command")
)

func validateService(s string) error {
	if s != db_models.CommandServiceNeConfig && s != db_models.CommandServiceNeCommand {
		return ErrCommandService
	}
	return nil
}

func CreateCommand(c *db_models.Command) error {
	if err := validateService(c.Service); err != nil {
		return err
	}
	existing, err := store.GetSingleton().GetCommandByTriple(c.NeID, c.Service, c.CmdText)
	if err != nil {
		return err
	}
	if existing != nil {
		return ErrCommandExists
	}
	return store.GetSingleton().CreateCommand(c)
}

func GetCommand(id int64) (*db_models.Command, error) {
	c, err := store.GetSingleton().GetCommandByID(id)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, ErrCommandNotFound
	}
	return c, nil
}

// ListCommands filters by NE and/or service. Pass 0 / "" to disable a filter.
func ListCommands(neID int64, service string) ([]*db_models.Command, error) {
	return store.GetSingleton().ListCommands(neID, service)
}

func UpdateCommand(c *db_models.Command) error {
	if err := validateService(c.Service); err != nil {
		return err
	}
	return store.GetSingleton().UpdateCommand(c)
}

func DeleteCommand(id int64) error {
	return store.GetSingleton().DeleteCommandByID(id)
}
