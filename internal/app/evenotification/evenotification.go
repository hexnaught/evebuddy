// Package evenotification contains the business logic for dealing with Eve Online notifications.
// It defines the notification types and related categories
// and provides a service for rendering notifications titles and bodies.
package evenotification

import (
	"context"
	"time"

	"github.com/ErikKalkoken/evebuddy/internal/app"
	"github.com/ErikKalkoken/evebuddy/internal/optional"
)

type EveUniverseService interface {
	GetOrCreateEntityESI(ctx context.Context, id int32) (*app.EveEntity, error)
	GetOrCreateMoonESI(ctx context.Context, id int32) (*app.EveMoon, error)
	GetOrCreateLocationESI(ctx context.Context, id int64) (*app.EveLocation, error)
	GetOrCreatePlanetESI(ctx context.Context, id int32) (*app.EvePlanet, error)
	GetOrCreateSolarSystemESI(ctx context.Context, id int32) (*app.EveSolarSystem, error)
	GetOrCreateTypeESI(ctx context.Context, id int32) (*app.EveType, error)
	ToEntities(ctx context.Context, ids []int32) (map[int32]*app.EveEntity, error)
}

// EveNotificationService is a service for rendering notifications.
type EveNotificationService struct {
	eus EveUniverseService
}

func New(eus EveUniverseService) *EveNotificationService {
	s := &EveNotificationService{eus: eus}
	return s
}

// RenderESI renders title and body for all supported notification types and returns them.
// Returns empty title and body for unsupported notification types.
func (s *EveNotificationService) RenderESI(ctx context.Context, type_, text string, timestamp time.Time) (optional.Optional[string], optional.Optional[string], error) {
	switch t := Type(type_); t {
	case BillOutOfMoneyMsg,
		BillPaidCorpAllMsg,
		CorpAllBillMsg,
		InfrastructureHubBillAboutToExpire,
		IHubDestroyedByBillFailure:
		return s.renderBilling(ctx, t, text)

	case CharAppAcceptMsg,
		CharAppRejectMsg,
		CharAppWithdrawMsg,
		CharLeftCorpMsg,
		CorpAppInvitedMsg,
		CorpAppNewMsg,
		CorpAppRejectCustomMsg:
		return s.renderCorporate(ctx, t, text)

	case OrbitalAttacked,
		OrbitalReinforced:
		return s.renderOrbital(ctx, t, text)

	case MoonminingExtractionStarted,
		MoonminingExtractionFinished,
		MoonminingAutomaticFracture,
		MoonminingExtractionCancelled,
		MoonminingLaserFired:
		return s.renderMoonMining(ctx, t, text)

	case OwnershipTransferred,
		StructureAnchoring,
		StructureDestroyed,
		StructureFuelAlert,
		StructureImpendingAbandonmentAssetsAtRisk,
		StructureItemsDelivered,
		StructureItemsMovedToSafety,
		StructureLostArmor,
		StructureLostShields,
		StructureOnline,
		StructureServicesOffline,
		StructuresReinforcementChanged,
		StructureUnanchoring,
		StructureUnderAttack,
		StructureWentHighPower,
		StructureWentLowPower:
		return s.renderStructure(ctx, t, text, timestamp)

	case TowerAlertMsg,
		TowerResourceAlertMsg:
		return s.renderTower(ctx, t, text)
	case AllWarSurrenderMsg,
		CorpWarSurrenderMsg,
		DeclareWar,
		WarAdopted,
		WarDeclared,
		WarHQRemovedFromSpace,
		WarInherited,
		WarInvalid,
		WarRetractedByConcord:
		return s.renderWar(ctx, t, text)
	case EntosisCaptureStarted,
		SovAllClaimAcquiredMsg,
		SovAllClaimLostMsg,
		SovCommandNodeEventStarted,
		SovStructureDestroyed,
		SovStructureReinforced:
		return s.renderSov(ctx, t, text)
	}
	return optional.Optional[string]{}, optional.Optional[string]{}, nil
}
