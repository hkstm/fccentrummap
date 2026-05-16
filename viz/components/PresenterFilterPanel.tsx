'use client';

import { ChevronUp } from 'lucide-react';
import { useEffect, useId, useMemo, useState } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { Checkbox } from '@/components/ui/checkbox';
import { cn } from '@/lib/utils';

export const PRESENTER_FILTER_DESKTOP_QUERY = '(min-width: 768px)';

export function getInitialPresenterFilterExpanded() {
  return true;
}

export function getPresenterFilterSummary(selectedCount: number, totalCount: number) {
  if (totalCount === 0) return 'Geen Amsterdammers';
  if (selectedCount === totalCount) return `Alle ${totalCount} Amsterdammers`;
  if (selectedCount === 0) return 'Geen Amsterdammers geselecteerd';
  if (selectedCount === 1) return `1 van ${totalCount} Amsterdammers`;
  return `${selectedCount} van ${totalCount} Amsterdammers`;
}

type Props = {
  presenters: string[];
  selectedPresenters: Set<string>;
  colors: Record<string, string>;
  setPresenter: (name: string, checked: boolean) => void;
  selectAll: () => void;
  deselectAll: () => void;
};

export function PresenterFilterPanel({
  presenters,
  selectedPresenters,
  colors,
  setPresenter,
  selectAll,
  deselectAll,
}: Props) {
  const panelId = useId();
  const [expanded, setExpanded] = useState(() => getInitialPresenterFilterExpanded());
  const selectedCount = useMemo(
    () => presenters.filter((name) => selectedPresenters.has(name)).length,
    [presenters, selectedPresenters],
  );
  const summary = getPresenterFilterSummary(selectedCount, presenters.length);

  useEffect(() => {
    if (!window.matchMedia) return;
    setExpanded(window.matchMedia(PRESENTER_FILTER_DESKTOP_QUERY).matches);
  }, []);

  return (
    <div
      className={cn(
        'absolute bottom-3 left-3 right-3 z-[2] pointer-events-none flex flex-col items-stretch sm:bottom-4 sm:left-4 sm:right-auto sm:w-[min(340px,calc(100%-32px))]',
      )}
    >
      <Card
        className={cn(
          'pointer-events-auto flex max-h-[70dvh] min-h-0 flex-col overflow-hidden',
        )}
      >
        {expanded ? (
          <CardContent id={panelId} className="space-y-3 flex flex-col flex-1 min-h-0 p-4 pb-3">
            <div className="flex gap-2 flex-none">
              <Button variant="outline" size="sm" onClick={selectAll}>Alles selecteren</Button>
              <Button variant="outline" size="sm" onClick={deselectAll}>Alles deselecteren</Button>
            </div>
            <div className="flex-1 min-h-0 overflow-y-auto overscroll-contain pr-3 -mr-1 [-webkit-overflow-scrolling:touch]">
              <ul className="space-y-1">
                {presenters.map((name) => {
                  const checked = selectedPresenters.has(name);
                  return (
                    <li key={name}>
                      <label className="flex min-h-11 cursor-pointer items-center gap-2 rounded-md px-1 text-base leading-[1.4] [font-family:'Noto_Sans',Inter,'Helvetica_Neue',helvetica,arial,sans-serif] hover:bg-[#f5f5f5]">
                        <Checkbox checked={checked} onCheckedChange={(value) => setPresenter(name, value === true)} />
                        <span>{name}</span>
                        {checked && (
                          <svg className="ml-1 h-6 w-6" viewBox="0 0 60 60" aria-hidden="true">
                            <g transform="translate(30,30)">
                              <rect x={-3} y={-15} width={6} height={30} rx={1.5} fill={colors[name] ?? '#0B57D0'} transform="rotate(45)" />
                              <rect x={-3} y={-15} width={6} height={30} rx={1.5} fill={colors[name] ?? '#0B57D0'} transform="rotate(-45)" />
                            </g>
                          </svg>
                        )}
                      </label>
                    </li>
                  );
                })}
              </ul>
            </div>
          </CardContent>
        ) : (
          <div id={panelId} hidden />
        )}
        <CardHeader className="flex-none p-0">
          <button
            type="button"
            aria-expanded={expanded}
            aria-controls={panelId}
            onClick={() => setExpanded((current) => !current)}
            className="flex min-h-12 w-full items-center justify-between gap-3 border-0 bg-white px-4 py-3 text-left hover:bg-[#f5f5f5] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#1a73e8] focus-visible:ring-offset-2"
          >
            <span className="min-w-0">
              <span className="block text-lg font-bold uppercase leading-none [font-family:'Garage_Gothic',Inter,'Helvetica_Neue',helvetica,arial,sans-serif]">
                Spots van
              </span>
              <span className="mt-1 block truncate text-sm normal-case text-[#5c5c5c] [font-family:'Noto_Sans',Inter,'Helvetica_Neue',helvetica,arial,sans-serif]">
                {summary}
              </span>
            </span>
            <ChevronUp
              className={cn('h-5 w-5 shrink-0 transition-transform', expanded && 'rotate-180')}
              aria-hidden="true"
            />
          </button>
        </CardHeader>
      </Card>
    </div>
  );
}
