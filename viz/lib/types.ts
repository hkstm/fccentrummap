export type Spot = {
  spotId: string;
  placeId: string;
  spotName: string;
  presenterName: string;
  categoryName: string;
  latitude: number;
  longitude: number;
  youtubeLink: string;
  articleUrl?: string;
};

export type PresenterMetadata = {
  presenterName: string;
};

export type CategoryMetadata = {
  categoryName: string;
};

export type SpotsData = {
  presenters?: PresenterMetadata[];
  categories?: CategoryMetadata[];
  spots: Spot[];
};

export type SpotWithPosition = Spot & {
  position: google.maps.LatLngLiteral;
};
