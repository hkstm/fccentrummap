import { describe, expect, it } from 'vitest';
import { getDefaultSelectedPresenters } from './SpotsMap';

describe('getDefaultSelectedPresenters', () => {
  it('selects the first three presenters from the export order by default', () => {
    expect([...getDefaultSelectedPresenters(['Latest', 'Second', 'Third', 'Fourth'])]).toEqual([
      'Latest',
      'Second',
      'Third',
    ]);
  });

  it('selects all presenters when fewer than three are available', () => {
    expect([...getDefaultSelectedPresenters(['One', 'Two'])]).toEqual(['One', 'Two']);
  });
});
