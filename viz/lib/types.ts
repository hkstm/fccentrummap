export type Spot = {
  placeId: string;
  spotName: string;
  presenterName: string;
  latitude: number;
  longitude: number;
  youtubeLink: string;
  articleUrl?: string;
};

export type Presenter = {
  presenterName: string;
};

export type SpotsData = {
  spots: Spot[];
  presenters: Presenter[];
};

export type SpotWithPosition = Spot & {
  position: google.maps.LatLngLiteral;
};
