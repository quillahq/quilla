package sql

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/quilla-hq/quilla/pkg/store"
	"github.com/quilla-hq/quilla/types"
)

func (s *SQLStore) CreateGate(gate *types.Gate) (*types.Gate, error) {
	if gate.ID == "" {
		gate.ID = uuid.New().String()
	}

	tx := s.db.Begin()
	if err := tx.Create(gate).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()

	return gate, nil
}

func (s *SQLStore) UpdateGate(gate *types.Gate) error {
	if gate.ID == "" {
		return fmt.Errorf("ID not specified")
	}
	return s.db.Save(gate).Error
}

func (s *SQLStore) GetGate(q *types.GetGateQuery) (*types.Gate, error) {
	var result types.Gate
	err := s.db.Where(&types.Gate{
		ID:         q.ID,
		Identifier: q.Identifier,
	}).First(&result).Error

	if err == gorm.ErrRecordNotFound {

		return nil, store.ErrRecordNotFound
	}

	return &result, err
}
