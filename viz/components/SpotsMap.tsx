'use client';

import { APIProvider, AdvancedMarker, InfoWindow, Map, useMap } from '@vis.gl/react-google-maps';
import { useEffect, useMemo, useState } from 'react';
import { buildPresenterColorMap } from '@/lib/color';
import { loadSpotsData } from '@/lib/data';
import { SpotWithPosition } from '@/lib/types';
import { AmsterdamXMarker } from './AmsterdamXMarker';
import { PresenterFilterPanel } from './PresenterFilterPanel';

const AMS_BOUNDS: google.maps.LatLngBoundsLiteral = {
  south: 52.274525,
  west: 4.711585,
  north: 52.461764,
  east: 5.073559,
};

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
  const [selectedPresenters, setSelectedPresenters] = useState<Set<string>>(new Set());
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
        const uniquePresenters = data.presenters.map((p) => p.presenterName);
        const resolved = data.spots.map((spot) => ({
          ...spot,
          position: { lat: spot.latitude, lng: spot.longitude },
        }));

        setPresenters(uniquePresenters);
        setSelectedPresenters(new Set(uniquePresenters));
        setSpots(resolved);
        setError(null);
      } catch (e) {
        setError(e instanceof Error ? e.message : 'Unknown map/data error occurred.');
      }
    };

    void geocode();
  }, [map]);

  const colors = useMemo(() => buildPresenterColorMap(presenters), [presenters]);
  const filteredSpots = useMemo(
    () => spots.filter((spot) => selectedPresenters.has(spot.presenterName)),
    [spots, selectedPresenters],
  );
  const activeSpot = useMemo(
    () => filteredSpots.find((spot) => `${spot.placeId}-${spot.presenterName}` === activeSpotKey) ?? null,
    [filteredSpots, activeSpotKey],
  );

  if (error) {
    return <div className="errorState" role="alert">{error}</div>;
  }

  return (
    <>
      <PresenterFilterPanel
        presenters={presenters}
        selectedPresenters={selectedPresenters}
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

      {filteredSpots.map((spot) => {
        const markerKey = `${spot.placeId}-${spot.presenterName}`;
        return (
          <AdvancedMarker
            key={markerKey}
            position={spot.position}
            clickable
            onClick={() => setActiveSpotKey((prev) => (prev === markerKey ? null : markerKey))}
          >
            <AmsterdamXMarker
              color={colors[spot.presenterName] ?? '#0B57D0'}
              label={spot.spotName}
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
          <div className="spotTooltip" role="dialog" aria-label={`Spot details: ${activeSpot.spotName}`}>
            <div className="spotTooltipTopRow">
              {activeSpot.articleUrl ? (
                <a
                  className="spotTooltipMetaLink"
                  href={activeSpot.articleUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  {activeSpot.presenterName}
                </a>
              ) : (
                <p className="spotTooltipMeta">{activeSpot.presenterName}</p>
              )}
              <button
                type="button"
                className="spotTooltipClose"
                aria-label="Close tooltip"
                onClick={() => setActiveSpotKey(null)}
              >
                ×
              </button>
            </div>
            <h3>{activeSpot.spotName}</h3>
            <button
              type="button"
              className="spotTooltipAction"
              onClick={() => window.open(activeSpot.youtubeLink, '_blank', 'noopener,noreferrer')}
            >
              ▶ Watch on YouTube
            </button>
            <button
              type="button"
              className="spotTooltipAction"
              onClick={() => window.open(`https://www.google.com/maps/place/?q=place_id:${activeSpot.placeId}`, '_blank', 'noopener,noreferrer')}
            >
              📍 Open in Google Maps
            </button>
          </div>
        </InfoWindow>
      )}
    </>
  );
}

export function SpotsMap() {
  const apiKey = process.env.NEXT_PUBLIC_DEMO_GOOGLE_MAPS_API_KEY;
  const mapId = process.env.NEXT_PUBLIC_DEMO_MAP_ID;

  if (!apiKey || !mapId) {
    return (
      <div className="errorState" role="alert">
        Missing Google Maps config. Set NEXT_PUBLIC_DEMO_GOOGLE_MAPS_API_KEY and NEXT_PUBLIC_DEMO_MAP_ID.
      </div>
    );
  }

  return (
    <APIProvider apiKey={apiKey}>
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
        </header>
        <Map
          mapId={mapId}
          renderingType="VECTOR"
          mapTypeControl={false}
          defaultCenter={{ lat: 52.3676, lng: 4.9041 }}
          defaultZoom={12}
          gestureHandling="greedy"
        >
          <BoundsFitter />
          <SpotsLayer />
        </Map>
      </main>
    </APIProvider>
  );
}
