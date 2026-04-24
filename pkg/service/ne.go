package service

import (
	"errors"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
)

var (
	ErrNeNotFound = errors.New("ne: not found")
	ErrNeExists   = errors.New("ne: namespace already taken")
)

func CreateNE(n *db_models.NE) error {
	existing, err := store.GetSingleton().GetNEByNamespace(n.Namespace)
	if err != nil {
		return err
	}
	if existing != nil {
		return ErrNeExists
	}
	return store.GetSingleton().CreateNE(n)
}

func GetNE(id int64) (*db_models.NE, error) {
	n, err := store.GetSingleton().GetNEByID(id)
	if err != nil {
		return nil, err
	}
	if n == nil {
		return nil, ErrNeNotFound
	}
	return n, nil
}

func GetNEByNamespace(ns string) (*db_models.NE, error) {
	n, err := store.GetSingleton().GetNEByNamespace(ns)
	if err != nil {
		return nil, err
	}
	if n == nil {
		return nil, ErrNeNotFound
	}
	return n, nil
}

func ListNEs() ([]*db_models.NE, error) {
	return store.GetSingleton().ListNEs()
}

func UpdateNE(n *db_models.NE) error {
	return store.GetSingleton().UpdateNE(n)
}

func DeleteNE(id int64) error {
	return store.GetSingleton().DeleteNEByID(id)
}
