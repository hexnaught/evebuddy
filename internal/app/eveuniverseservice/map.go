package eveuniverseservice

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/antihax/goesi/esi"
	esioptional "github.com/antihax/goesi/optional"
	"golang.org/x/sync/errgroup"

	"github.com/ErikKalkoken/evebuddy/internal/app"
	"github.com/ErikKalkoken/evebuddy/internal/app/storage"
	"github.com/ErikKalkoken/evebuddy/internal/xslices"
)

func (s *EveUniverseService) GetStargateSolarSystemsESI(ctx context.Context, stargateIDs []int32) ([]*app.EveSolarSystem, error) {
	g := new(errgroup.Group)
	systemIDs := make([]int32, len(stargateIDs))
	for i, id := range stargateIDs {
		g.Go(func() error {
			x, _, err := s.esiClient.ESI.UniverseApi.GetUniverseStargatesStargateId(ctx, id, nil)
			if err != nil {
				return err
			}
			systemIDs[i] = x.Destination.SystemId
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	g = new(errgroup.Group)
	systems := make([]*app.EveSolarSystem, len(systemIDs))
	for i, id := range systemIDs {
		g.Go(func() error {
			st, err := s.GetOrCreateSolarSystemESI(ctx, id)
			if err != nil {
				return err
			}
			systems[i] = st
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	slices.SortFunc(systems, func(a, b *app.EveSolarSystem) int {
		return cmp.Compare(a.Name, b.Name)
	})
	return systems, nil
}

func (s *EveUniverseService) GetSolarSystemPlanets(ctx context.Context, planets []app.EveSolarSystemPlanet) ([]*app.EvePlanet, error) {
	oo := make([]*app.EvePlanet, len(planets))
	g := new(errgroup.Group)
	for i, p := range planets {
		g.Go(func() error {
			st, err := s.GetOrCreatePlanetESI(ctx, p.PlanetID)
			if err != nil {
				return err
			}
			oo[i] = st
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	slices.SortFunc(oo, func(a, b *app.EvePlanet) int {
		return cmp.Compare(a.Name, b.Name)
	})
	return oo, nil
}

func (s *EveUniverseService) GetStarTypeID(ctx context.Context, id int32) (int32, error) {
	x2, _, err := s.esiClient.ESI.UniverseApi.GetUniverseStarsStarId(ctx, id, nil)
	if err != nil {
		return 0, err
	}
	return x2.TypeId, nil
}

func (s *EveUniverseService) GetSolarSystemInfoESI(ctx context.Context, solarSystemID int32) (int32, []app.EveSolarSystemPlanet, []int32, []*app.EveEntity, []*app.EveLocation, error) {
	x, _, err := s.esiClient.ESI.UniverseApi.GetUniverseSystemsSystemId(ctx, solarSystemID, nil)
	if err != nil {
		return 0, nil, nil, nil, nil, err
	}
	planets := xslices.Map(x.Planets, func(p esi.GetUniverseSystemsSystemIdPlanet) app.EveSolarSystemPlanet {
		return app.EveSolarSystemPlanet{
			AsteroidBeltIDs: p.AsteroidBelts,
			MoonIDs:         p.Moons,
			PlanetID:        p.PlanetId,
		}
	})
	_, err = s.AddMissingEntities(ctx, slices.Concat(
		[]int32{solarSystemID, x.ConstellationId},
		x.Stations,
	))
	if err != nil {
		return 0, nil, nil, nil, nil, err
	}
	stations := make([]*app.EveEntity, len(x.Stations))
	for i, id := range x.Stations {
		st, err := s.getValidEntity(ctx, id)
		if err != nil {
			return 0, nil, nil, nil, nil, err
		}
		stations[i] = st
	}
	slices.SortFunc(stations, func(a, b *app.EveEntity) int {
		return a.Compare(b)
	})
	xx, err := s.st.ListEveLocationInSolarSystem(ctx, solarSystemID)
	if err != nil {
		return 0, nil, nil, nil, nil, err
	}
	structures := xslices.Filter(xx, func(x *app.EveLocation) bool {
		return x.Variant() == app.EveLocationStructure
	})
	return x.StarId, planets, x.Stargates, stations, structures, nil
}

func (s *EveUniverseService) GetRegionConstellationsESI(ctx context.Context, id int32) ([]*app.EveEntity, error) {
	region, _, err := s.esiClient.ESI.UniverseApi.GetUniverseRegionsRegionId(ctx, id, nil)
	if err != nil {
		return nil, err
	}
	xx, err := s.ToEntities(ctx, region.Constellations)
	if err != nil {
		return nil, err
	}
	oo := slices.Collect(maps.Values(xx))
	slices.SortFunc(oo, func(a, b *app.EveEntity) int {
		return a.Compare(b)
	})
	return oo, nil
}

func (s *EveUniverseService) GetConstellationSolarSytemsESI(ctx context.Context, id int32) ([]*app.EveSolarSystem, error) {
	o, _, err := s.esiClient.ESI.UniverseApi.GetUniverseConstellationsConstellationId(ctx, id, nil)
	if err != nil {
		return nil, err
	}
	g := new(errgroup.Group)
	systems := make([]*app.EveSolarSystem, len(o.Systems))
	for i, id := range o.Systems {
		g.Go(func() error {
			st, err := s.GetOrCreateSolarSystemESI(ctx, id)
			if err != nil {
				return err
			}
			systems[i] = st
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	slices.SortFunc(systems, func(a, b *app.EveSolarSystem) int {
		return cmp.Compare(a.Name, b.Name)
	})
	return systems, nil
}

func (s *EveUniverseService) GetOrCreateRegionESI(ctx context.Context, id int32) (*app.EveRegion, error) {
	o, err := s.st.GetEveRegion(ctx, id)
	if errors.Is(err, app.ErrNotFound) {
		return s.createRegionFromESI(ctx, id)
	}
	return o, err
}

func (s *EveUniverseService) createRegionFromESI(ctx context.Context, id int32) (*app.EveRegion, error) {
	key := fmt.Sprintf("createRegionFromESI-%d", id)
	y, err, _ := s.sfg.Do(key, func() (any, error) {
		region, _, err := s.esiClient.ESI.UniverseApi.GetUniverseRegionsRegionId(ctx, id, nil)
		if err != nil {
			return nil, err
		}
		arg := storage.CreateEveRegionParams{
			ID:          region.RegionId,
			Description: region.Description,
			Name:        region.Name,
		}
		return s.st.CreateEveRegion(ctx, arg)
	})
	if err != nil {
		return nil, err
	}
	return y.(*app.EveRegion), nil
}

func (s *EveUniverseService) GetOrCreateConstellationESI(ctx context.Context, id int32) (*app.EveConstellation, error) {
	o, err := s.st.GetEveConstellation(ctx, id)
	if errors.Is(err, app.ErrNotFound) {
		return s.createConstellationFromESI(ctx, id)
	}
	return o, err
}

func (s *EveUniverseService) createConstellationFromESI(ctx context.Context, id int32) (*app.EveConstellation, error) {
	key := fmt.Sprintf("createConstellationFromESI-%d", id)
	y, err, _ := s.sfg.Do(key, func() (any, error) {
		constellation, _, err := s.esiClient.ESI.UniverseApi.GetUniverseConstellationsConstellationId(ctx, id, nil)
		if err != nil {
			return nil, err
		}
		_, err = s.GetOrCreateRegionESI(ctx, constellation.RegionId)
		if err != nil {
			return nil, err
		}
		arg := storage.CreateEveConstellationParams{
			ID:       constellation.ConstellationId,
			RegionID: constellation.RegionId,
			Name:     constellation.Name,
		}
		if err := s.st.CreateEveConstellation(ctx, arg); err != nil {
			return nil, err
		}
		return s.st.GetEveConstellation(ctx, id)
	})
	if err != nil {
		return nil, err
	}
	return y.(*app.EveConstellation), nil
}

func (s *EveUniverseService) GetOrCreateSolarSystemESI(ctx context.Context, id int32) (*app.EveSolarSystem, error) {
	o, err := s.st.GetEveSolarSystem(ctx, id)
	if errors.Is(err, app.ErrNotFound) {
		return s.createSolarSystemFromESI(ctx, id)
	}
	return o, err
}

func (s *EveUniverseService) createSolarSystemFromESI(ctx context.Context, id int32) (*app.EveSolarSystem, error) {
	key := fmt.Sprintf("createSolarSystemFromESI-%d", id)
	y, err, _ := s.sfg.Do(key, func() (any, error) {
		system, _, err := s.esiClient.ESI.UniverseApi.GetUniverseSystemsSystemId(ctx, id, nil)
		if err != nil {
			return nil, err
		}
		constellation, err := s.GetOrCreateConstellationESI(ctx, system.ConstellationId)
		if err != nil {
			return nil, err
		}
		arg := storage.CreateEveSolarSystemParams{
			ID:              system.SystemId,
			ConstellationID: constellation.ID,
			Name:            system.Name,
			SecurityStatus:  system.SecurityStatus,
		}
		if err := s.st.CreateEveSolarSystem(ctx, arg); err != nil {
			return nil, err
		}
		return s.st.GetEveSolarSystem(ctx, id)
	})
	if err != nil {
		return nil, err
	}
	return y.(*app.EveSolarSystem), nil
}

func (s *EveUniverseService) GetOrCreatePlanetESI(ctx context.Context, id int32) (*app.EvePlanet, error) {
	o, err := s.st.GetEvePlanet(ctx, id)
	if errors.Is(err, app.ErrNotFound) {
		return s.createPlanetFromESI(ctx, id)
	}
	return o, err
}

func (s *EveUniverseService) createPlanetFromESI(ctx context.Context, id int32) (*app.EvePlanet, error) {
	key := fmt.Sprintf("createPlanetFromESI-%d", id)
	y, err, _ := s.sfg.Do(key, func() (any, error) {
		planet, _, err := s.esiClient.ESI.UniverseApi.GetUniversePlanetsPlanetId(ctx, id, nil)
		if err != nil {
			return nil, err
		}
		system, err := s.GetOrCreateSolarSystemESI(ctx, planet.SystemId)
		if err != nil {
			return nil, err
		}
		type_, err := s.GetOrCreateTypeESI(ctx, planet.TypeId)
		if err != nil {
			return nil, err
		}
		arg := storage.CreateEvePlanetParams{
			ID:            planet.PlanetId,
			Name:          planet.Name,
			SolarSystemID: system.ID,
			TypeID:        type_.ID,
		}
		if err := s.st.CreateEvePlanet(ctx, arg); err != nil {
			return nil, err
		}
		return s.st.GetEvePlanet(ctx, id)
	})
	if err != nil {
		return nil, err
	}
	return y.(*app.EvePlanet), nil
}

func (s *EveUniverseService) GetOrCreateMoonESI(ctx context.Context, id int32) (*app.EveMoon, error) {
	o, err := s.st.GetEveMoon(ctx, id)
	if errors.Is(err, app.ErrNotFound) {
		return s.createMoonFromESI(ctx, id)
	}
	return o, err
}

func (s *EveUniverseService) createMoonFromESI(ctx context.Context, id int32) (*app.EveMoon, error) {
	key := fmt.Sprintf("createMoonFromESI-%d", id)
	y, err, _ := s.sfg.Do(key, func() (any, error) {
		moon, _, err := s.esiClient.ESI.UniverseApi.GetUniverseMoonsMoonId(ctx, id, nil)
		if err != nil {
			return nil, err
		}
		system, err := s.GetOrCreateSolarSystemESI(ctx, moon.SystemId)
		if err != nil {
			return nil, err
		}
		arg := storage.CreateEveMoonParams{
			ID:            moon.MoonId,
			Name:          moon.Name,
			SolarSystemID: system.ID,
		}
		if err := s.st.CreateEveMoon(ctx, arg); err != nil {
			return nil, err
		}
		return s.st.GetEveMoon(ctx, id)
	})
	if err != nil {
		return nil, err
	}
	return y.(*app.EveMoon), nil
}

// GetRouteESI returns a route between two solar systems.
// When no route can be found it returns an empty slice.
type RoutePreference string

func (s *EveUniverseService) GetRouteESI(ctx context.Context, destination, origin *app.EveSolarSystem, flag app.RoutePreference) ([]*app.EveSolarSystem, error) {
	if slices.Index(app.RoutePreferences(), flag) == -1 {
		return nil, fmt.Errorf("invalid flag: %s", flag)
	}
	if destination.ID == origin.ID {
		return []*app.EveSolarSystem{origin}, nil
	}
	if destination.IsWormholeSpace() || origin.IsWormholeSpace() {
		return []*app.EveSolarSystem{}, nil // no route possible
	}
	ids, r, err := s.esiClient.ESI.RoutesApi.GetRouteOriginDestination(ctx, destination.ID, origin.ID, &esi.GetRouteOriginDestinationOpts{
		Flag: esioptional.NewString(flag.String()),
	})
	if err != nil {
		if r.StatusCode == 404 {
			return []*app.EveSolarSystem{}, nil // no route found
		}
		return nil, err
	}
	systems := make([]*app.EveSolarSystem, len(ids))
	g := new(errgroup.Group)
	for i, id := range ids {
		g.Go(func() error {
			system, err := s.GetOrCreateSolarSystemESI(ctx, id)
			if err != nil {
				return err
			}
			systems[i] = system
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return systems, nil
}
