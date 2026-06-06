'use client';

import { APIProvider, AdvancedMarker, InfoWindow, Map, useMap } from '@vis.gl/react-google-maps';
import { useEffect, useMemo, useState } from 'react';
import { buildPresenterColorMap } from '@/lib/color';
import { loadSpotsData } from '@/lib/data';
import { buildMapShareSearch, getInitialMapShareState, getSpotKey } from '@/lib/share-state';
import { CategoryMetadata, PresenterMetadata, SpotWithPosition } from '@/lib/types';
import { AmsterdamXMarker } from './AmsterdamXMarker';
import { PresenterFilterPanel } from './PresenterFilterPanel';
import { SpotTooltipCard } from './SpotTooltipCard';

declare global {
  interface Window {
    gm_authFailure?: () => void;
  }
}

const AMS_BOUNDS: google.maps.LatLngBoundsLiteral = {
  south: 52.274525,
  west: 4.711585,
  north: 52.461764,
  east: 5.073559,
};

const DEFAULT_SELECTED_PRESENTER_COUNT = 3;

export function getDefaultSelectedPresenters(presenters: string[]) {
  return new Set(presenters.slice(0, DEFAULT_SELECTED_PRESENTER_COUNT));
}

export function derivePresentersFromSpots(spots: { presenterName: string }[]) {
  return [...new Set(spots.map((spot) => spot.presenterName.trim()).filter(Boolean))].sort((a, b) => a.localeCompare(b));
}

export function derivePresentersFromMetadata(metadata: PresenterMetadata[] | undefined, spots: { presenterName: string }[]) {
  const presentInSpots = new Set(spots.map((spot) => spot.presenterName.trim()).filter(Boolean));
  const ordered = (metadata ?? []).map((entry) => entry.presenterName.trim()).filter((name) => presentInSpots.has(name));
  const seen = new Set<string>();
  const deduped = ordered.filter((name) => {
    if (seen.has(name)) return false;
    seen.add(name);
    return true;
  });
  const missing = derivePresentersFromSpots(spots).filter((name) => !seen.has(name));
  return [...deduped, ...missing];
}

export function deriveCategoriesFromMetadata(metadata: CategoryMetadata[] | undefined, spots: { categoryName: string }[]) {
  const presentInSpots = new Set(spots.map((spot) => spot.categoryName.trim()).filter(Boolean));
  const ordered = (metadata ?? []).map((entry) => entry.categoryName.trim()).filter((name) => presentInSpots.has(name));
  const seen = new Set<string>();
  const deduped = ordered.filter((name) => {
    if (seen.has(name)) return false;
    seen.add(name);
    return true;
  });
  const missing = [...presentInSpots].filter((name) => !seen.has(name)).sort((a, b) => a.localeCompare(b));
  return [...deduped, ...missing];
}

function BoundsFitter() {
  const map = useMap();
  useEffect(() => {
    if (!map) return;
    map.fitBounds(AMS_BOUNDS);
  }, [map]);
  return null;
}

function SpotsLayer() {
  const map = useMap();
  const [spots, setSpots] = useState<SpotWithPosition[]>([]);
  const [presenters, setPresenters] = useState<string[]>([]);
  const [categories, setCategories] = useState<string[]>([]);
  const [selectedPresenters, setSelectedPresenters] = useState<Set<string>>(new Set());
  const [selectedCategories, setSelectedCategories] = useState<Set<string>>(new Set());
  const [activeSpotKey, setActiveSpotKey] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!map) return;
    const listener = map.addListener('click', () => setActiveSpotKey(null));
    return () => listener.remove();
  }, [map]);

  useEffect(() => {
    const geocode = async () => {
      if (!map) return;

      try {
        const data = await loadSpotsData();
        const uniquePresenters = derivePresentersFromMetadata(data.presenters, data.spots);
        const uniqueCategories = deriveCategoriesFromMetadata(data.categories, data.spots);
        const resolved = data.spots.map((spot) => ({
          ...spot,
          position: { lat: spot.latitude, lng: spot.longitude },
        }));

        const defaultSelectedPresenters = getDefaultSelectedPresenters(uniquePresenters);
        const initialShareState = typeof window === 'undefined'
          ? { selectedPresenters: defaultSelectedPresenters, activeSpotKey: null }
          : getInitialMapShareState(window.location.search, uniquePresenters, resolved, defaultSelectedPresenters);

        setPresenters(uniquePresenters);
        setCategories(uniqueCategories);
        setSelectedPresenters(initialShareState.selectedPresenters);
        setSelectedCategories(new Set(uniqueCategories));
        setActiveSpotKey(initialShareState.activeSpotKey);
        setSpots(resolved);
        setError(null);
      } catch (e) {
        setError(e instanceof Error ? e.message : 'Er is een onbekende kaart/datafout opgetreden.');
      }
    };

    void geocode();
  }, [map]);

  const colors = useMemo(() => buildPresenterColorMap(presenters), [presenters]);
  const filteredSpots = useMemo(
    () => spots.filter((spot) => selectedPresenters.has(spot.presenterName) && selectedCategories.has(spot.categoryName)),
    [spots, selectedPresenters, selectedCategories],
  );
  useEffect(() => {
    if (!activeSpotKey) return;
    const stillVisible = filteredSpots.some((spot) => getSpotKey(spot) === activeSpotKey);
    if (!stillVisible) setActiveSpotKey(null);
  }, [activeSpotKey, filteredSpots]);

  const activeSpot = useMemo(
    () => filteredSpots.find((spot) => getSpotKey(spot) === activeSpotKey) ?? null,
    [filteredSpots, activeSpotKey],
  );

  useEffect(() => {
    if (typeof window === 'undefined' || presenters.length === 0) return;

    const nextSearch = buildMapShareSearch(
      window.location.search,
      presenters,
      selectedPresenters,
      activeSpotKey,
      getDefaultSelectedPresenters(presenters),
    );
    const currentSearch = window.location.search;
    if (nextSearch === currentSearch) return;

    window.history.replaceState({}, '', `${window.location.pathname}${nextSearch}${window.location.hash}`);
  }, [activeSpotKey, presenters, selectedPresenters]);

  if (error) {
    return <div className="errorState" role="alert">{error}</div>;
  }

  return (
    <>
      <PresenterFilterPanel
        presenters={presenters}
        selectedPresenters={selectedPresenters}
        colors={colors}
        categories={categories}
        selectedCategories={selectedCategories}
        setCategory={(name, checked) => {
          setSelectedCategories((prev) => {
            const next = new Set(prev);
            if (checked) next.add(name); else next.delete(name);
            return next;
          });
        }}
        selectAllCategories={() => setSelectedCategories(new Set(categories))}
        deselectAllCategories={() => setSelectedCategories(new Set())}
        setPresenter={(name, checked) => {
          setSelectedPresenters((prev) => {
            const next = new Set(prev);
            if (checked) next.add(name); else next.delete(name);
            return next;
          });
        }}
        selectAll={() => setSelectedPresenters(new Set(presenters))}
        deselectAll={() => setSelectedPresenters(new Set())}
      />

      {filteredSpots.map((spot, index) => {
        const markerKey = getSpotKey(spot);
        return (
          <AdvancedMarker
            key={markerKey}
            position={spot.position}
            clickable
            zIndex={index + 1}
            onClick={() => setActiveSpotKey((prev) => (prev === markerKey ? null : markerKey))}
          >
            <AmsterdamXMarker
              color={colors[spot.presenterName] ?? '#0B57D0'}
              label={spot.spotName}
              selected={activeSpotKey === markerKey}
            />
          </AdvancedMarker>
        );
      })}

      {activeSpot && (
        <InfoWindow
          position={activeSpot.position}
          pixelOffset={[0, -44]}
          headerDisabled
          onCloseClick={() => setActiveSpotKey(null)}
        >
          <SpotTooltipCard spot={activeSpot} onClose={() => setActiveSpotKey(null)} />
        </InfoWindow>
      )}
    </>
  );
}

export function SpotsMap() {
  const apiKey = process.env.NEXT_PUBLIC_DEMO_GOOGLE_MAPS_API_KEY;
  const mapId = 'c14f6dcc70143a8c9d9b26b0';
  const [mapsLoadError, setMapsLoadError] = useState<string | null>(null);

  useEffect(() => {
    window.gm_authFailure = () => {
      setMapsLoadError('Google Maps API-sleutel is ongeldig of niet toegestaan voor deze origin. Controleer de Maps JavaScript API key/referrer-instellingen.');
    };
    return () => {
      window.gm_authFailure = undefined;
    };
  }, []);

  if (!apiKey) {
    return (
      <div className="errorState" role="alert">
        Google Maps configuratie ontbreekt. Stel NEXT_PUBLIC_DEMO_GOOGLE_MAPS_API_KEY in.
      </div>
    );
  }

  if (mapsLoadError) {
    return <div className="errorState" role="alert">{mapsLoadError}</div>;
  }

  return (
    <APIProvider
      apiKey={apiKey}
      onError={() => setMapsLoadError('Google Maps kon niet worden geladen. Controleer de Maps JavaScript API key/referrer-instellingen.')}
    >
      <main className="layout">
        <header className="brandBanner" aria-label="FC Centrum">
          <a className="brandLogoLink" href="https://fccentrum.nl" target="_blank" rel="noopener noreferrer" aria-label="FC Centrum home">
            <img
              className="brandLogoImage"
              src="https://fccentrum.nl/wp-content/uploads/2023/12/fanclubcentrum-logo-wit.svg"
              alt="Fanclub Centrum"
              width={287}
              height={67}
            />
          </a>
          <div className="hidden justify-self-end self-start mt-1 sm:block">
            <a href="https://hkstm.dev" target="_blank" rel="noopener noreferrer" className="createdByLink">
              created by hkstm.dev
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 12 12" aria-hidden="true">
                <path fill="currentColor" d="M11 8H9.5V3.56L5.071 7.99l-1.06-1.061L8.44 2.5H4V1h7v7Z" />
              </svg>
            </a>
          </div>
        </header>
        <section className="mapTitleBar relative z-10" aria-label="Kaart sectie titel">
          <svg className="mapTitleIcon" xmlns="http://www.w3.org/2000/svg" width="17" height="24" viewBox="0 0 17 24" fill="none" aria-hidden="true">
            <path d="M14.6691 6.65563V4.39429H12.3491V2.26804H9.9469V0H7.02838V2.26804H4.6279V4.39429H2.30791V6.65563H0V13.9683C0.0195488 17.0067 1.23618 19.4226 3.61423 21.1503C4.66869 21.9086 5.77519 22.5963 6.92603 23.2084L8.50144 24L10.0872 23.2017L10.097 23.1973C11.2408 22.5889 12.3404 21.9051 13.3881 21.1509C15.7673 19.4204 16.9828 16.9966 17 13.9398V6.65563H14.6691Z" fill="#ED1C24" />
          </svg>
          <h1 className="mapTitleText">Kaart</h1>
          <div className="mapTitleDivider" aria-hidden="true" />
        </section>
        <div className="relative flex-1">
          <Map
            mapId={mapId}
            renderingType="VECTOR"
            mapTypeControl={false}
            streetViewControl={false}
            cameraControl={false}
            rotateControl={false}
            defaultCenter={{ lat: 52.3676, lng: 4.9041 }}
            defaultZoom={12}
            gestureHandling="greedy"
            style={{ width: '100%', height: '100%', position: 'absolute', inset: 0 }}
          >
            <BoundsFitter />
            <SpotsLayer />
          </Map>
        </div>
      </main>
    </APIProvider>
  );
}
