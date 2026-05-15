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
    <Card className="absolute left-4 top-[88px] z-[2] w-[min(340px,calc(100%-32px))]">
      <CardHeader className="pb-3">
        <CardTitle className="text-lg font-bold uppercase [font-family:'Garage_Gothic',Inter,'Helvetica_Neue',helvetica,arial,sans-serif]">
          Presenters
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        <div className="flex gap-2">
          <Button variant="outline" size="sm" onClick={selectAll}>Select all</Button>
          <Button variant="outline" size="sm" onClick={deselectAll}>Deselect all</Button>
        </div>
        <ScrollArea className="max-h-[40dvh] pr-2">
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
  );
}
