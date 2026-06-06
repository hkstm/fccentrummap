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

function isPresenterMetadata(value: unknown): boolean {
  const v = value as { presenterName?: unknown };
  return !!v && typeof v.presenterName === 'string' && v.presenterName.trim().length > 0;
}

function isCategoryMetadata(value: unknown): boolean {
  const v = value as { categoryName?: unknown };
  return !!v && typeof v.categoryName === 'string' && v.categoryName.trim().length > 0;
}

function validateSpotsData(value: unknown): value is SpotsData {
  const v = value as SpotsData;
  return !!v
    && Array.isArray(v.spots)
    && (v.presenters === undefined || Array.isArray(v.presenters))
    && (v.categories === undefined || Array.isArray(v.categories))
    && (v.presenters === undefined || v.presenters.every(isPresenterMetadata))
    && (v.categories === undefined || v.categories.every(isCategoryMetadata))
    && v.spots.every((spot) => typeof spot.spotId === 'string'
      && spot.spotId.trim().length > 0
      && typeof spot.placeId === 'string'
      && typeof spot.spotName === 'string'
      && typeof spot.presenterName === 'string'
      && typeof spot.categoryName === 'string'
      && spot.categoryName.trim().length > 0
      && typeof spot.latitude === 'number'
      && typeof spot.longitude === 'number'
      && typeof spot.youtubeLink === 'string'
      && spot.youtubeLink.trim().length > 0
      && isValidUrl(spot.youtubeLink)
      && (spot.articleUrl === undefined || typeof spot.articleUrl === 'string'))
    && new Set(v.spots.map((spot) => spot.spotId.trim())).size === v.spots.length;
}

export async function loadSpotsData(): Promise<SpotsData> {
  const basePath = process.env.NEXT_PUBLIC_BASE_PATH ?? '';
  const dataUrl = `${basePath}/data/spots.json`;

  let response: Response;
  try {
    response = await fetch(dataUrl, { cache: 'no-store' });
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    throw new Error(`Network or fetch error loading ${dataUrl}: ${message}`);
  }

  if (!response.ok) {
    throw new Error(`HTTP error loading ${dataUrl}: ${response.status} ${response.statusText}`);
  }

  let parsed: unknown;
  try {
    parsed = await response.json();
  } catch {
    throw new Error(`The ${dataUrl} file is not valid JSON.`);
  }

  if (!validateSpotsData(parsed)) {
    throw new Error(`The ${dataUrl} file does not match the required spots schema.`);
  }

  return parsed;
}
