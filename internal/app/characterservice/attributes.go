package characterservice

import (
	"context"

	"github.com/ErikKalkoken/evebuddy/internal/app"
	"github.com/ErikKalkoken/evebuddy/internal/app/storage"
	"github.com/antihax/goesi/esi"
)

func (s *CharacterService) GetAttributes(ctx context.Context, characterID int32) (*app.CharacterAttributes, error) {
	return s.st.GetCharacterAttributes(ctx, characterID)
}

func (s *CharacterService) updateAttributesESI(ctx context.Context, arg app.CharacterUpdateSectionParams) (bool, error) {
	if arg.Section != app.SectionAttributes {
		panic("called with wrong section")
	}
	return s.updateSectionIfChanged(
		ctx, arg,
		func(ctx context.Context, characterID int32) (any, error) {
			attributes, _, err := s.esiClient.ESI.SkillsApi.GetCharactersCharacterIdAttributes(ctx, characterID, nil)
			if err != nil {
				return false, err
			}
			return attributes, nil
		},
		func(ctx context.Context, characterID int32, data any) error {
			attributes := data.(esi.GetCharactersCharacterIdAttributesOk)
			arg := storage.UpdateOrCreateCharacterAttributesParams{
				CharacterID:   characterID,
				BonusRemaps:   int(attributes.BonusRemaps),
				Charisma:      int(attributes.Charisma),
				Intelligence:  int(attributes.Intelligence),
				LastRemapDate: attributes.LastRemapDate,
				Memory:        int(attributes.Memory),
				Perception:    int(attributes.Perception),
				Willpower:     int(attributes.Willpower),
			}
			if err := s.st.UpdateOrCreateCharacterAttributes(ctx, arg); err != nil {
				return err
			}
			return nil
		})
}
