'use client';

import { ChevronUp } from 'lucide-react';
import { useEffect, useId, useMemo, useState } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { Checkbox } from '@/components/ui/checkbox';
import { cn } from '@/lib/utils';

export const PRESENTER_FILTER_DESKTOP_QUERY = '(min-width: 768px)';

type FilterCategory = 'presenters' | 'categories';

const FILTER_STATUS_TABS: Array<{ id: FilterCategory; primary: string }> = [
  { id: 'presenters', primary: 'SPOTS VAN' },
  { id: 'categories', primary: 'CATEGORIEËN' },
];

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

export function getCategoryFilterSummary(selectedCount: number, totalCount: number) {
  if (totalCount === 0) return 'Geen categorieën';
  if (selectedCount === totalCount) return `Alle ${totalCount} categorieën`;
  if (selectedCount === 0) return 'Geen categorieën geselecteerd';
  if (selectedCount === 1) return `1 van ${totalCount} categorieën`;
  return `${selectedCount} van ${totalCount} categorieën`;
}

type Props = {
  presenters: string[];
  selectedPresenters: Set<string>;
  colors: Record<string, string>;
  categories?: string[];
  selectedCategories?: Set<string>;
  setPresenter: (name: string, checked: boolean) => void;
  setCategory?: (name: string, checked: boolean) => void;
  selectAll: () => void;
  deselectAll: () => void;
  selectAllCategories?: () => void;
  deselectAllCategories?: () => void;
};

export function PresenterFilterPanel({
  presenters,
  selectedPresenters,
  colors,
  categories = [],
  selectedCategories = new Set<string>(),
  setPresenter,
  setCategory,
  selectAll,
  deselectAll,
  selectAllCategories,
  deselectAllCategories,
}: Props) {
  const panelId = useId();
  const [expanded, setExpanded] = useState(() => getInitialPresenterFilterExpanded());
  const [activeFilterCategory, setActiveFilterCategory] = useState<FilterCategory>('presenters');
  const selectedCount = useMemo(
    () => presenters.filter((name) => selectedPresenters.has(name)).length,
    [presenters, selectedPresenters],
  );
  const selectedCategoryCount = useMemo(
    () => categories.filter((name) => selectedCategories.has(name)).length,
    [categories, selectedCategories],
  );
  const presenterSummary = getPresenterFilterSummary(selectedCount, presenters.length);
  const categorySummary = getCategoryFilterSummary(selectedCategoryCount, categories.length);
  const statusTabs = FILTER_STATUS_TABS.map((tab) => ({
    ...tab,
    additional: tab.id === 'presenters' ? presenterSummary : categorySummary,
  }));

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
          <CardContent id={panelId} className="flex min-h-0 flex-1 flex-col space-y-3 border-b border-[#d8d8d8] p-4 pb-3">
            {activeFilterCategory === 'presenters' ? (
              <>
                <div className="flex-none">
                  <p className="mb-2 text-sm font-bold uppercase tracking-wide text-[#5c5c5c]">Amsterdammers</p>
                  <div className="flex gap-2">
                    <Button variant="outline" size="sm" onClick={selectAll}>Alles selecteren</Button>
                    <Button variant="outline" size="sm" onClick={deselectAll}>Alles deselecteren</Button>
                  </div>
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
              </>
            ) : (
              <>
                <div className="flex-none">
                  <p className="mb-2 text-sm font-bold uppercase tracking-wide text-[#5c5c5c]">Categorieën</p>
                  {setCategory && (selectAllCategories || deselectAllCategories) && (
                    <div className="flex gap-2">
                      {selectAllCategories && <Button variant="outline" size="sm" onClick={selectAllCategories}>Alles selecteren</Button>}
                      {deselectAllCategories && <Button variant="outline" size="sm" onClick={deselectAllCategories}>Alles deselecteren</Button>}
                    </div>
                  )}
                </div>
                <div className="flex-1 min-h-0 overflow-y-auto overscroll-contain pr-3 -mr-1 [-webkit-overflow-scrolling:touch]">
                {categories.length > 0 && setCategory ? (
                  <ul className="space-y-1">
                    {categories.map((name) => {
                      const checked = selectedCategories.has(name);
                      return (
                        <li key={name}>
                          <label className="flex min-h-10 cursor-pointer items-center gap-2 rounded-md px-1 text-base leading-[1.4] [font-family:'Noto_Sans',Inter,'Helvetica_Neue',helvetica,arial,sans-serif] hover:bg-[#f5f5f5]">
                            <Checkbox checked={checked} onCheckedChange={(value) => setCategory(name, value === true)} />
                            <span>{name}</span>
                          </label>
                        </li>
                      );
                    })}
                  </ul>
                ) : (
                  <p className="text-sm text-[#5c5c5c]">Geen categorieën beschikbaar</p>
                )}
                </div>
              </>
            )}
          </CardContent>
        ) : (
          <div id={panelId} hidden />
        )}
        <CardHeader className="flex-none p-0">
          <div
            className="flex min-h-12 w-full items-stretch border-0 bg-white text-left"
            onClick={() => setExpanded((current) => !current)}
          >
            <div className="relative min-w-0 flex-1 py-2 pl-3 pr-0">
              <div
                role="tablist"
                aria-label="Filterstatus"
                className="flex min-w-0 snap-x scroll-px-2 gap-2 overflow-x-auto overscroll-x-contain scroll-smooth pr-8 [-webkit-overflow-scrolling:touch]"
              >
                {statusTabs.map((tab) => {
                  const active = tab.id === activeFilterCategory;
                  return (
                    <button
                      key={tab.id}
                      type="button"
                      role="tab"
                      aria-selected={active}
                      aria-controls={panelId}
                      tabIndex={active ? 0 : -1}
                      onClick={(event) => {
                        event.stopPropagation();
                        setActiveFilterCategory(tab.id);
                        setExpanded(true);
                        event.currentTarget.scrollIntoView?.({
                          behavior: 'smooth',
                          block: 'nearest',
                          inline: tab.id === 'categories' ? 'end' : 'start',
                        });
                      }}
                      className={cn(
                        'shrink-0 snap-start !rounded-md !border !border-transparent !px-2 !py-1 text-left outline-none hover:!border-[#d0d0d0] focus:outline-none focus:ring-0 focus-visible:ring-2 focus-visible:ring-[#1a73e8] focus-visible:ring-offset-2',
                        active ? 'min-w-[11rem]' : 'w-auto',
                        'transition-colors duration-200 ease-out',
                      )}
                    >
                      <span
                        className={cn(
                          "block whitespace-nowrap text-lg font-bold uppercase leading-none [font-family:'Garage_Gothic',Inter,'Helvetica_Neue',helvetica,arial,sans-serif] transition-colors duration-200 ease-out",
                          active ? 'text-[#1f1f1f]' : 'text-[#b8b8b8]',
                        )}
                      >
                        {tab.primary}
                      </span>
                      {active && (
                        <span className="mt-1 block truncate text-sm normal-case text-[#5c5c5c] [font-family:'Noto_Sans',Inter,'Helvetica_Neue',helvetica,arial,sans-serif] transition-all duration-200 ease-out">
                          {tab.additional}
                        </span>
                      )}
                    </button>
                  );
                })}
              </div>
              <div
                aria-hidden="true"
                className={cn(
                  'pointer-events-none absolute inset-y-0 left-0 w-5 bg-gradient-to-r from-white/80 to-transparent transition-opacity duration-200 ease-out',
                  activeFilterCategory === 'categories' ? 'opacity-100' : 'opacity-0',
                )}
              />
              <div
                aria-hidden="true"
                className="pointer-events-none absolute inset-y-0 right-0 w-5 bg-gradient-to-l from-white/80 to-transparent"
              />
            </div>
            <button
              type="button"
              aria-label={expanded ? 'Filteropties inklappen' : 'Filteropties uitklappen'}
              aria-expanded={expanded}
              aria-controls={panelId}
              onClick={(event) => {
                event.stopPropagation();
                setExpanded((current) => !current);
              }}
              className="flex w-10 shrink-0 appearance-none items-center justify-center !rounded-none !border-0 !border-l !border-[#d8d8d8] !bg-transparent !p-0 outline-none hover:!bg-[#f5f5f5] focus:outline-none focus-visible:ring-2 focus-visible:ring-[#1a73e8] focus-visible:ring-inset"
            >
              <ChevronUp
                className={cn('h-5 w-5 transition-transform', expanded && 'rotate-180')}
                aria-hidden="true"
              />
            </button>
          </div>
        </CardHeader>
      </Card>
    </div>
  );
}
