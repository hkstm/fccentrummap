'use client';

import { APIProvider, AdvancedMarker, InfoWindow, Map, useMap } from '@vis.gl/react-google-maps';
import { useEffect, useMemo, useState } from 'react';
import { buildPresenterColorMap } from '@/lib/color';
import { loadSpotsData } from '@/lib/data';
import { SpotWithPosition } from '@/lib/types';
import { AmsterdamXMarker } from './AmsterdamXMarker';
import { PresenterFilterPanel } from './PresenterFilterPanel';
import { SpotTooltipCard } from './SpotTooltipCard';

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
        setSelectedPresenters(new Set(uniquePresenters.length > 0 ? [uniquePresenters[0]] : []));
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
        colors={colors}
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
          <SpotTooltipCard spot={activeSpot} onClose={() => setActiveSpotKey(null)} />
        </InfoWindow>
      )}
    </>
  );
}

export function SpotsMap() {
  const apiKey = process.env.NEXT_PUBLIC_DEMO_GOOGLE_MAPS_API_KEY;
  const mapId = 'c14f6dcc70143a8c9d9b26b0';

  if (!apiKey) {
    return (
      <div className="errorState" role="alert">
        Missing Google Maps config. Set NEXT_PUBLIC_DEMO_GOOGLE_MAPS_API_KEY.
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
        <section className="mapTitleBar relative z-10" aria-label="Map section title">
          <svg className="mapTitleIcon" xmlns="http://www.w3.org/2000/svg" width="17" height="24" viewBox="0 0 17 24" fill="none" aria-hidden="true">
            <path d="M14.6691 6.65563V4.39429H12.3491V2.26804H9.9469V0H7.02838V2.26804H4.6279V4.39429H2.30791V6.65563H0V13.9683C0.0195488 17.0067 1.23618 19.4226 3.61423 21.1503C4.66869 21.9086 5.77519 22.5963 6.92603 23.2084L8.50144 24L10.0872 23.2017L10.097 23.1973C11.2408 22.5889 12.3404 21.9051 13.3881 21.1509C15.7673 19.4204 16.9828 16.9966 17 13.9398V6.65563H14.6691Z" fill="#ED1C24" />
          </svg>
          <h1 className="mapTitleText">Map</h1>
          <div className="mapTitleDivider" aria-hidden="true" />
        </section>
        <div className="relative flex-1">
          <Map
            mapId={mapId}
            renderingType="VECTOR"
            mapTypeControl={false}
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
