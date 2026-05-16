import { describe, expect, it } from 'vitest';
import { buildMapShareSearch, getInitialMapShareState, getSpotKey } from './share-state';

const presenters = ['Ray Fuego', 'Sef', 'Akwasi'];
const spots = [
  { placeId: 'place-ray', presenterName: 'Ray Fuego', spotName: 'Ray Spot', latitude: 52.1, longitude: 4.1 },
  { placeId: 'place-sef', presenterName: 'Sef', spotName: 'Sef Spot', latitude: 52.2, longitude: 4.2 },
  { placeId: 'place-akwasi', presenterName: 'Akwasi', spotName: 'Akwasi Spot', latitude: 52.3, longitude: 4.3 },
];

describe('map share state', () => {
  it('builds marker-first share URLs with presenters after the open marker', () => {
    const search = buildMapShareSearch(
      '?foo=bar&spot=old&presenters=old',
      presenters,
      new Set(['Sef', 'Akwasi']),
      getSpotKey(spots[1]),
    );

    expect(search).toBe('?spot=place-sef%3A%3ASef%3A%3ASef+Spot%3A%3A52.2%3A%3A4.2&presenters=Sef%2CAkwasi&foo=bar');
  });

  it('omits presenter filters from share URLs when all presenters are selected', () => {
    const search = buildMapShareSearch('?foo=bar', presenters, new Set(presenters), getSpotKey(spots[1]));

    expect(search).toBe('?spot=place-sef%3A%3ASef%3A%3ASef+Spot%3A%3A52.2%3A%3A4.2&foo=bar');
  });

  it('selects all presenters by default when no filter query is present', () => {
    const state = getInitialMapShareState('', presenters, spots);

    expect([...state.selectedPresenters]).toEqual(presenters);
    expect(state.activeSpotKey).toBeNull();
  });

  it('uses provided default presenters when no filter query is present', () => {
    const state = getInitialMapShareState('', presenters, spots, new Set(['Ray Fuego', 'Sef']));

    expect([...state.selectedPresenters]).toEqual(['Ray Fuego', 'Sef']);
    expect(state.activeSpotKey).toBeNull();
  });

  it('omits presenter filters from share URLs when the default presenters are selected', () => {
    const search = buildMapShareSearch('?foo=bar', presenters, new Set(['Ray Fuego', 'Sef']), null, new Set(['Ray Fuego', 'Sef']));

    expect(search).toBe('?foo=bar');
  });

  it('keeps all selected presenters shareable when default presenters are a subset', () => {
    const search = buildMapShareSearch('?foo=bar', presenters, new Set(presenters), null, new Set(['Ray Fuego', 'Sef']));

    expect(search).toBe('?presenters=Ray+Fuego%2CSef%2CAkwasi&foo=bar');
  });

  it('restores preselected filters from the URL', () => {
    const state = getInitialMapShareState('?presenters=Sef%2CAkwasi', presenters, spots);

    expect([...state.selectedPresenters]).toEqual(['Sef', 'Akwasi']);
    expect(state.activeSpotKey).toBeNull();
  });

  it('opens a requested marker and ensures its presenter is selected', () => {
    const activeKey = getSpotKey(spots[0]);
    const state = getInitialMapShareState(`?spot=${encodeURIComponent(activeKey)}&presenters=Sef`, presenters, spots);

    expect(state.activeSpotKey).toBe(activeKey);
    expect([...state.selectedPresenters]).toEqual(['Sef', 'Ray Fuego']);
  });
});
