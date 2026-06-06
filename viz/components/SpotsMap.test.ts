import { describe, expect, it } from 'vitest';
import { deriveCategoriesFromMetadata, derivePresentersFromMetadata, getDefaultSelectedPresenters } from './SpotsMap';

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

describe('export metadata ordering', () => {
  it('uses top-level presenter metadata order and appends missing spot presenters deterministically', () => {
    const spots = [{ presenterName: 'B' }, { presenterName: 'A' }, { presenterName: 'C' }];
    expect(derivePresentersFromMetadata([{ presenterName: 'C' }, { presenterName: 'A' }], spots)).toEqual(['C', 'A', 'B']);
  });

  it('uses top-level category metadata order and appends missing spot categories deterministically', () => {
    const spots = [{ categoryName: 'Restaurants' }, { categoryName: 'Overig' }, { categoryName: 'Winkelen' }];
    expect(deriveCategoriesFromMetadata([{ categoryName: 'Winkelen' }, { categoryName: 'Restaurants' }], spots)).toEqual(['Winkelen', 'Restaurants', 'Overig']);
  });
});
