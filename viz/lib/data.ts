import { SpotsData } from './types';

function isValidUrl(value: string): boolean {
  try {
    // eslint-disable-next-line no-new
    new URL(value);
    return true;
  } catch {
    return false;
  }
}

function validateSpotsData(value: unknown): value is SpotsData {
  const v = value as SpotsData;
  return !!v
    && Array.isArray(v.spots)
    && Array.isArray(v.presenters)
    && v.spots.every((spot) => typeof spot.placeId === 'string'
      && typeof spot.spotName === 'string'
      && typeof spot.presenterName === 'string'
      && typeof spot.latitude === 'number'
      && typeof spot.longitude === 'number'
      && typeof spot.youtubeLink === 'string'
      && spot.youtubeLink.trim().length > 0
      && isValidUrl(spot.youtubeLink)
      && (spot.articleUrl === undefined || typeof spot.articleUrl === 'string'))
    && v.presenters.every((p) => typeof p.presenterName === 'string');
}

export async function loadSpotsData(): Promise<SpotsData> {
  let response: Response;
  try {
    response = await fetch('/data/spots.json', { cache: 'no-store' });
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    throw new Error(`Network or fetch error loading /data/spots.json: ${message}`);
  }

  if (!response.ok) {
    throw new Error(`HTTP error loading /data/spots.json: ${response.status} ${response.statusText}`);
  }

  let parsed: unknown;
  try {
    parsed = await response.json();
  } catch {
    throw new Error('The /data/spots.json file is not valid JSON.');
  }

  if (!validateSpotsData(parsed)) {
    throw new Error('The /data/spots.json file does not match the required spots/presenters schema.');
  }

  return parsed;
}
