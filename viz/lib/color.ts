export const PRESENTER_PALETTE = [
  '#0B57D0', // blue
  '#D93025', // red
  '#137333', // green
  '#A142F4', // violet
  '#E37400', // orange
  '#00897B', // teal
  '#C2185B', // magenta
  '#455A64', // blue gray
  '#1565C0', // deep blue
  '#EF6C00', // deep orange
  '#2E7D32', // forest green
  '#7B1FA2', // purple
  '#6D4C41', // brown
  '#1E88E5', // bright blue
  '#8E24AA', // royal purple
  '#43A047', // bright green
] as const;

export function buildPresenterColorMap(presenters: string[]): Record<string, string> {
  return presenters.reduce<Record<string, string>>((acc, name, idx) => {
    acc[name] = PRESENTER_PALETTE[idx % PRESENTER_PALETTE.length];
    return acc;
  }, {});
}
