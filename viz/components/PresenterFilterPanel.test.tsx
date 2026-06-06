import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { useState } from 'react';
import { describe, expect, it, vi } from 'vitest';
import { PresenterFilterPanel, getCategoryFilterSummary, getPresenterFilterSummary } from './PresenterFilterPanel';

const presenters = ['Ray Fuego', 'Sef', 'Akwasi'];
const categories = ['Muziek', 'Kunst'];
const colors = {
  'Ray Fuego': '#e30613',
  Sef: '#111111',
  Akwasi: '#0B57D0',
};

function mockViewport(isDesktop: boolean) {
  Object.defineProperty(window, 'matchMedia', {
    writable: true,
    value: vi.fn().mockImplementation((query: string) => ({
      matches: isDesktop && query === '(min-width: 768px)',
      media: query,
      onchange: null,
      addListener: vi.fn(),
      removeListener: vi.fn(),
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      dispatchEvent: vi.fn(),
    })),
  });
}

function StatefulFilter({ withCategories = false }: { withCategories?: boolean }) {
  const [selected, setSelected] = useState(() => new Set(presenters));
  const [selectedCategories, setSelectedCategories] = useState(() => new Set(categories));

  return (
    <PresenterFilterPanel
      presenters={presenters}
      selectedPresenters={selected}
      colors={colors}
      categories={withCategories ? categories : []}
      selectedCategories={withCategories ? selectedCategories : new Set<string>()}
      setPresenter={(name, checked) => {
        setSelected((current) => {
          const next = new Set(current);
          if (checked) next.add(name);
          else next.delete(name);
          return next;
        });
      }}
      setCategory={withCategories ? (name, checked) => {
        setSelectedCategories((current) => {
          const next = new Set(current);
          if (checked) next.add(name);
          else next.delete(name);
          return next;
        });
      } : undefined}
      selectAll={() => setSelected(new Set(presenters))}
      deselectAll={() => setSelected(new Set())}
      selectAllCategories={() => setSelectedCategories(new Set(categories))}
      deselectAllCategories={() => setSelectedCategories(new Set())}
    />
  );
}

describe('PresenterFilterPanel', () => {
  it('renders collapsed by default on mobile and keeps only the compact tabbed panel visible', () => {
    mockViewport(false);

    render(<StatefulFilter />);

    const toggle = screen.getByRole('button', { name: /filteropties uitklappen/i });
    expect(toggle).toHaveAttribute('aria-expanded', 'false');
    expect(screen.getByRole('tab', { name: /spots van/i })).toHaveAttribute('aria-selected', 'true');
    expect(screen.getByRole('tab', { name: /categorieën/i })).toHaveAttribute('aria-selected', 'false');
    expect(screen.getByText('Alle 3 Amsterdammers')).toBeInTheDocument();
    expect(screen.queryByText('Geen categorieën')).not.toBeInTheDocument();
    expect(screen.queryByLabelText('Ray Fuego')).not.toBeInTheDocument();
  });

  it('renders expanded by default on desktop', () => {
    mockViewport(true);

    render(<StatefulFilter />);

    expect(screen.getByRole('button', { name: /filteropties inklappen/i })).toHaveAttribute('aria-expanded', 'true');
    expect(screen.getByLabelText('Ray Fuego')).toBeVisible();
    expect(screen.getByLabelText('Sef')).toBeVisible();
    expect(screen.getByLabelText('Akwasi')).toBeVisible();
  });

  it('opens the options automatically when either status tab is clicked', async () => {
    mockViewport(false);
    const user = userEvent.setup();

    render(<StatefulFilter withCategories />);

    const toggle = screen.getByRole('button', { name: /filteropties uitklappen/i });
    expect(toggle).toHaveAttribute('aria-expanded', 'false');

    await user.click(screen.getByRole('tab', { name: /categorieën/i }));
    expect(screen.getByRole('button', { name: /filteropties inklappen/i })).toHaveAttribute('aria-expanded', 'true');
    expect(screen.getByLabelText('Muziek')).toBeVisible();

    await user.click(screen.getByRole('button', { name: /filteropties inklappen/i }));
    expect(screen.getByRole('button', { name: /filteropties uitklappen/i })).toHaveAttribute('aria-expanded', 'false');

    await user.click(screen.getByRole('tab', { name: /spots van/i }));
    expect(screen.getByRole('button', { name: /filteropties inklappen/i })).toHaveAttribute('aria-expanded', 'true');
    expect(screen.getByLabelText('Ray Fuego')).toBeVisible();
  });

  it('toggles from a semantic button while preserving presenter selection state', async () => {
    mockViewport(false);
    const user = userEvent.setup();

    render(<StatefulFilter />);

    const toggle = screen.getByRole('button', { name: /filteropties uitklappen/i });
    await user.click(toggle);
    expect(toggle).toHaveAttribute('aria-expanded', 'true');

    await user.click(screen.getByLabelText('Ray Fuego'));
    expect(screen.getByRole('tab', { name: /spots van/i })).toHaveTextContent('2 van 3 Amsterdammers');

    await user.click(toggle);
    expect(toggle).toHaveAttribute('aria-expanded', 'false');
    expect(screen.getByRole('tab', { name: /spots van/i })).toHaveTextContent('2 van 3 Amsterdammers');

    await user.click(toggle);
    expect(screen.getByLabelText('Ray Fuego')).not.toBeChecked();
    expect(screen.getByLabelText('Sef')).toBeChecked();
    expect(screen.getByLabelText('Akwasi')).toBeChecked();
  });

  it('keeps multi-select and select-all/deselect-all behavior available in the expanded presenter panel', async () => {
    mockViewport(true);
    const user = userEvent.setup();

    render(<StatefulFilter />);

    const panel = screen.getByRole('button', { name: /filteropties inklappen/i }).getAttribute('aria-controls');
    expect(panel).toBeTruthy();
    const controls = within(document.getElementById(panel!)!);

    expect(controls.getByLabelText('Ray Fuego')).toBeChecked();
    expect(controls.getByLabelText('Sef')).toBeChecked();
    expect(controls.getByLabelText('Akwasi')).toBeChecked();

    await user.click(controls.getByRole('button', { name: 'Alles deselecteren' }));
    expect(controls.getByLabelText('Ray Fuego')).not.toBeChecked();
    expect(controls.getByLabelText('Sef')).not.toBeChecked();
    expect(controls.getByLabelText('Akwasi')).not.toBeChecked();
    expect(screen.getByRole('tab', { name: /spots van/i })).toHaveTextContent('Geen Amsterdammers geselecteerd');

    await user.click(controls.getByRole('button', { name: 'Alles selecteren' }));
    expect(controls.getByLabelText('Ray Fuego')).toBeChecked();
    expect(controls.getByLabelText('Sef')).toBeChecked();
    expect(controls.getByLabelText('Akwasi')).toBeChecked();
    expect(screen.getByRole('tab', { name: /spots van/i })).toHaveTextContent('Alle 3 Amsterdammers');
  });

  it('switches status tabs to show category options without changing existing selections', async () => {
    mockViewport(true);
    const user = userEvent.setup();

    render(<StatefulFilter withCategories />);

    expect(screen.getByLabelText('Ray Fuego')).toBeChecked();
    expect(screen.queryByLabelText('Muziek')).not.toBeInTheDocument();

    await user.click(screen.getByRole('tab', { name: /categorieën/i }));
    expect(screen.getByRole('tab', { name: /categorieën/i })).toHaveAttribute('aria-selected', 'true');
    expect(screen.getByLabelText('Muziek')).toBeChecked();
    expect(screen.getByLabelText('Kunst')).toBeChecked();
    expect(screen.queryByLabelText('Ray Fuego')).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Alles selecteren' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Alles deselecteren' })).toBeInTheDocument();

    await user.click(screen.getByLabelText('Muziek'));
    expect(screen.getByRole('tab', { name: /categorieën/i })).toHaveTextContent('1 van 2 categorieën');

    await user.click(screen.getByRole('tab', { name: /spots van/i }));
    expect(screen.getByLabelText('Ray Fuego')).toBeChecked();
    expect(screen.queryByLabelText('Muziek')).not.toBeInTheDocument();

    await user.click(screen.getByRole('tab', { name: /categorieën/i }));
    expect(screen.getByLabelText('Muziek')).not.toBeChecked();
    expect(screen.getByLabelText('Kunst')).toBeChecked();
  });

  it('formats collapsed selection summaries', () => {
    expect(getPresenterFilterSummary(0, 0)).toBe('Geen Amsterdammers');
    expect(getPresenterFilterSummary(3, 3)).toBe('Alle 3 Amsterdammers');
    expect(getPresenterFilterSummary(0, 3)).toBe('Geen Amsterdammers geselecteerd');
    expect(getPresenterFilterSummary(1, 3)).toBe('1 van 3 Amsterdammers');
    expect(getPresenterFilterSummary(2, 3)).toBe('2 van 3 Amsterdammers');
    expect(getCategoryFilterSummary(0, 0)).toBe('Geen categorieën');
    expect(getCategoryFilterSummary(2, 2)).toBe('Alle 2 categorieën');
    expect(getCategoryFilterSummary(0, 2)).toBe('Geen categorieën geselecteerd');
    expect(getCategoryFilterSummary(1, 2)).toBe('1 van 2 categorieën');
  });
});
