package app

import (
	"context"
	"time"
)

// Entity ID for general sections
const (
	GeneralSectionEntityID   = 1
	GeneralSectionEntityName = "Eve Universe"
)

type SectionStatus struct {
	EntityID     int32
	EntityName   string
	CompletedAt  time.Time
	ContentHash  string
	ErrorMessage string
	SectionID    string
	SectionName  string
	StartedAt    time.Time
	Timeout      time.Duration
}

func (s SectionStatus) IsGeneralSection() bool {
	return s.EntityID == GeneralSectionEntityID
}

func (s SectionStatus) IsOK() bool {
	return s.ErrorMessage == ""
}

func (s SectionStatus) IsExpired() bool {
	if s.CompletedAt.IsZero() {
		return true
	}
	deadline := s.CompletedAt.Add(s.Timeout)
	return time.Now().After(deadline)
}

func (s SectionStatus) IsCurrent() bool {
	if s.CompletedAt.IsZero() {
		return false
	}
	return time.Now().Before(s.CompletedAt.Add(s.Timeout * 2))
}

func (s SectionStatus) IsMissing() bool {
	return s.CompletedAt.IsZero()
}

func (s SectionStatus) IsRunning() bool {
	return !s.StartedAt.IsZero()
}

type StatusCacheStorage interface {
	ListCharacterSectionStatus(context.Context, int32) ([]*CharacterSectionStatus, error)
	ListGeneralSectionStatus(context.Context) ([]*GeneralSectionStatus, error)
	ListCharactersShort(context.Context) ([]*CharacterShort, error)
}

type StatusCacheService interface {
	CharacterName(int32) string
	CharacterSectionExists(int32, CharacterSection) bool
	CharacterSectionSet(*CharacterSectionStatus)
	CharacterSectionSummary(int32) StatusSummary
	GeneralSectionExists(GeneralSection) bool
	GeneralSectionSet(*GeneralSectionStatus)
	GeneralSectionSummary() StatusSummary
	ListCharacters() []*CharacterShort
	SectionList(int32) []SectionStatus
	Summary() StatusSummary
	UpdateCharacters(ctx context.Context, r StatusCacheStorage) error
}
