import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Separator } from '@/components/ui/separator';
import { SpotWithPosition } from '@/lib/types';

type Props = {
  spot: SpotWithPosition;
  onClose: () => void;
};

export function SpotTooltipCard({ spot, onClose }: Props) {
  return (
    <Card className="spotTooltip" role="dialog" aria-label={`Spot details: ${spot.spotName}`}>
      <CardHeader className="pb-2">
        <div className="spotTooltipTopRow">
          {spot.articleUrl ? (
            <a
              className="spotTooltipMetaLink"
              href={spot.articleUrl}
              target="_blank"
              rel="noopener noreferrer"
            >
              {spot.presenterName}
            </a>
          ) : (
            <p className="spotTooltipMeta">{spot.presenterName}</p>
          )}
          <Button
            type="button"
            variant="ghost"
            size="sm"
            className="h-6 w-6 p-0 text-lg leading-none"
            aria-label="Close tooltip"
            onClick={onClose}
          >
            ×
          </Button>
        </div>
        <CardTitle className="spotTooltipTitle">{spot.spotName}</CardTitle>
      </CardHeader>
      <CardContent className="pt-0">
        <Separator className="mb-2" />
        <Button
          type="button"
          className="spotTooltipAction"
          onClick={() => window.open(spot.youtubeLink, '_blank', 'noopener,noreferrer')}
        >
          <svg className="spotTooltipActionIcon" viewBox="0 0 576 512" aria-hidden="true">
            <path d="M549.655 124.083c-6.281-23.65-24.787-42.276-48.284-48.597C458.781 64 288 64 288 64S117.22 64 74.629 75.486c-23.497 6.322-42.003 24.947-48.284 48.597-11.412 42.867-11.412 132.305-11.412 132.305s0 89.438 11.412 132.305c6.281 23.65 24.787 41.5 48.284 47.821C117.22 448 288 448 288 448s170.78 0 213.371-11.486c23.497-6.321 42.003-24.171 48.284-47.821 11.412-42.867 11.412-132.305 11.412-132.305s0-89.438-11.412-132.305zm-317.51 213.508V175.185l142.739 81.205-142.739 81.201z" />
          </svg>
          Watch on YouTube
        </Button>
        <Button
          type="button"
          className="spotTooltipAction mt-2"
          onClick={() => window.open(`https://www.google.com/maps/place/?q=place_id:${spot.placeId}`, '_blank', 'noopener,noreferrer')}
        >
          <svg className="spotTooltipActionIcon" viewBox="0 0 384 512" aria-hidden="true">
            <path d="M172.3 501.7C26.97 291 0 269.4 0 192 0 85.96 85.96 0 192 0s192 85.96 192 192c0 77.4-26.97 99-172.3 309.7-9.535 13.77-29.93 13.77-39.46 0zM192 272c44.11 0 80-35.89 80-80s-35.89-80-80-80-80 35.89-80 80 35.9 80 80 80z" />
          </svg>
          Open in Google Maps
        </Button>
      </CardContent>
    </Card>
  );
}
