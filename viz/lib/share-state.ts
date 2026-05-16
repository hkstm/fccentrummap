import type { Spot } from './types';

export const SPOT_QUERY_PARAM = 'spot';
export const PRESENTERS_QUERY_PARAM = 'presenters';

export function getSpotKey(spot: Pick<Spot, 'placeId' | 'presenterName' | 'spotName' | 'latitude' | 'longitude'>) {
  return [spot.placeId || 'no-place-id', spot.presenterName, spot.spotName, spot.latitude, spot.longitude].join('::');
}

export function getInitialMapShareState<TSpot extends Pick<Spot, 'placeId' | 'presenterName' | 'spotName' | 'latitude' | 'longitude'>>(
  search: string,
  presenters: string[],
  spots: TSpot[],
  defaultSelectedPresenters: Set<string> = new Set(presenters),
) {
  const presenterSet = new Set(presenters);
  const params = new URLSearchParams(search);
  const presenterParam = params.get(PRESENTERS_QUERY_PARAM);
  const selectedPresenters = new Set<string>();

  if (presenterParam === null) {
    presenters
      .filter((name) => defaultSelectedPresenters.has(name))
      .forEach((name) => selectedPresenters.add(name));
  } else {
    presenterParam
      .split(',')
      .map((name) => name.trim())
      .filter((name) => presenterSet.has(name))
      .forEach((name) => selectedPresenters.add(name));
  }

  let activeSpotKey: string | null = null;
  const requestedSpotKey = params.get(SPOT_QUERY_PARAM);
  if (requestedSpotKey) {
    const matchedSpot = spots.find((spot) => getSpotKey(spot) === requestedSpotKey);
    if (matchedSpot) {
      activeSpotKey = requestedSpotKey;
      selectedPresenters.add(matchedSpot.presenterName);
    }
  }

  return { selectedPresenters, activeSpotKey };
}

export function buildMapShareSearch(
  currentSearch: string,
  presenters: string[],
  selectedPresenters: Set<string>,
  activeSpotKey: string | null,
  defaultSelectedPresenters: Set<string> = new Set(presenters),
) {
  const currentParams = new URLSearchParams(currentSearch);
  const nextParams = new URLSearchParams();

  if (activeSpotKey) {
    nextParams.append(SPOT_QUERY_PARAM, activeSpotKey);
  }

  const selectedInPresenterOrder = presenters.filter((name) => selectedPresenters.has(name));
  const defaultSelectedInPresenterOrder = presenters.filter((name) => defaultSelectedPresenters.has(name));
  if (selectedInPresenterOrder.join('\0') !== defaultSelectedInPresenterOrder.join('\0')) {
    nextParams.append(PRESENTERS_QUERY_PARAM, selectedInPresenterOrder.join(','));
  }

  for (const [key, value] of currentParams.entries()) {
    if (key === SPOT_QUERY_PARAM || key === PRESENTERS_QUERY_PARAM) continue;
    nextParams.append(key, value);
  }

  const nextSearch = nextParams.toString();
  return nextSearch ? `?${nextSearch}` : '';
}
