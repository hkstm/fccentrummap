import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Checkbox } from '@/components/ui/checkbox';
import { ScrollArea } from '@/components/ui/scroll-area';

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
  return (
    <div className="absolute left-4 top-4 bottom-4 z-[2] w-[min(340px,calc(100%-32px))] pointer-events-none flex flex-col justify-start">
      <Card className="flex flex-col min-h-0 pointer-events-auto max-h-full">
        <CardHeader className="pb-3 flex-none">
        <CardTitle className="text-lg font-bold uppercase [font-family:'Garage_Gothic',Inter,'Helvetica_Neue',helvetica,arial,sans-serif]">
          Presenters
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-3 flex flex-col flex-1 min-h-0">
        <div className="flex gap-2 flex-none">
          <Button variant="outline" size="sm" onClick={selectAll}>Select all</Button>
          <Button variant="outline" size="sm" onClick={deselectAll}>Deselect all</Button>
        </div>
        <ScrollArea className="flex flex-col flex-1 min-h-0 pr-3 -mr-1">
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
        </ScrollArea>
        </CardContent>
      </Card>
    </div>
  );
}
