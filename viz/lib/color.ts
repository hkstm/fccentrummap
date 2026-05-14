const PRESENTER_PALETTE = [
  '#0B57D0', '#D93025', '#137333', '#A142F4', '#E37400', '#00897B', '#8E24AA', '#C2185B',
  '#1565C0', '#2E7D32', '#EF6C00', '#6D4C41', '#455A64', '#7B1FA2', '#1E88E5', '#43A047'
] as const;

export function buildPresenterColorMap(presenters: string[]): Record<string, string> {
  return [...presenters].sort((a, b) => a.localeCompare(b)).reduce<Record<string, string>>((acc, name, idx) => {
    acc[name] = PRESENTER_PALETTE[idx % PRESENTER_PALETTE.length];
    return acc;
  }, {});
}
