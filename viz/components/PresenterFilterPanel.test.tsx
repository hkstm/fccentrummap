import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { useState } from 'react';
import { describe, expect, it, vi } from 'vitest';
import { PresenterFilterPanel, getPresenterFilterSummary } from './PresenterFilterPanel';

const presenters = ['Ray Fuego', 'Sef', 'Akwasi'];
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

function StatefulFilter() {
  const [selected, setSelected] = useState(() => new Set(presenters));

  return (
    <PresenterFilterPanel
      presenters={presenters}
      selectedPresenters={selected}
      colors={colors}
      setPresenter={(name, checked) => {
        setSelected((current) => {
          const next = new Set(current);
          if (checked) next.add(name);
          else next.delete(name);
          return next;
        });
      }}
      selectAll={() => setSelected(new Set(presenters))}
      deselectAll={() => setSelected(new Set())}
    />
  );
}

describe('PresenterFilterPanel', () => {
  it('renders collapsed by default on mobile and keeps only the compact panel visible', () => {
    mockViewport(false);

    render(<StatefulFilter />);

    const toggle = screen.getByRole('button', { name: /spots van/i });
    expect(toggle).toHaveAttribute('aria-expanded', 'false');
    expect(toggle).toHaveTextContent('Alle 3 Amsterdammers');
    expect(screen.queryByLabelText('Ray Fuego')).not.toBeInTheDocument();
  });

  it('renders expanded by default on desktop', () => {
    mockViewport(true);

    render(<StatefulFilter />);

    const toggle = screen.getByRole('button', { name: /spots van/i });
    expect(toggle).toHaveAttribute('aria-expanded', 'true');
    expect(screen.getByLabelText('Ray Fuego')).toBeVisible();
    expect(screen.getByLabelText('Sef')).toBeVisible();
    expect(screen.getByLabelText('Akwasi')).toBeVisible();
  });

  it('toggles from a semantic button while preserving presenter selection state', async () => {
    mockViewport(false);
    const user = userEvent.setup();

    render(<StatefulFilter />);

    const toggle = screen.getByRole('button', { name: /spots van/i });
    await user.click(toggle);
    expect(toggle).toHaveAttribute('aria-expanded', 'true');

    await user.click(screen.getByLabelText('Ray Fuego'));
    expect(toggle).toHaveTextContent('2 van 3 Amsterdammers');

    await user.click(toggle);
    expect(toggle).toHaveAttribute('aria-expanded', 'false');
    expect(toggle).toHaveTextContent('2 van 3 Amsterdammers');

    await user.click(toggle);
    expect(screen.getByLabelText('Ray Fuego')).not.toBeChecked();
    expect(screen.getByLabelText('Sef')).toBeChecked();
    expect(screen.getByLabelText('Akwasi')).toBeChecked();
  });

  it('keeps multi-select and select-all/deselect-all behavior available in the expanded panel', async () => {
    mockViewport(true);
    const user = userEvent.setup();

    render(<StatefulFilter />);

    const panel = screen.getByRole('button', { name: /spots van/i }).getAttribute('aria-controls');
    expect(panel).toBeTruthy();
    const controls = within(document.getElementById(panel!)!);

    expect(controls.getByLabelText('Ray Fuego')).toBeChecked();
    expect(controls.getByLabelText('Sef')).toBeChecked();
    expect(controls.getByLabelText('Akwasi')).toBeChecked();

    await user.click(controls.getByRole('button', { name: 'Alles deselecteren' }));
    expect(controls.getByLabelText('Ray Fuego')).not.toBeChecked();
    expect(controls.getByLabelText('Sef')).not.toBeChecked();
    expect(controls.getByLabelText('Akwasi')).not.toBeChecked();
    expect(screen.getByRole('button', { name: /spots van/i })).toHaveTextContent('Geen Amsterdammers geselecteerd');

    await user.click(controls.getByRole('button', { name: 'Alles selecteren' }));
    expect(controls.getByLabelText('Ray Fuego')).toBeChecked();
    expect(controls.getByLabelText('Sef')).toBeChecked();
    expect(controls.getByLabelText('Akwasi')).toBeChecked();
    expect(screen.getByRole('button', { name: /spots van/i })).toHaveTextContent('Alle 3 Amsterdammers');
  });

  it('formats collapsed selection summaries', () => {
    expect(getPresenterFilterSummary(0, 0)).toBe('Geen Amsterdammers');
    expect(getPresenterFilterSummary(3, 3)).toBe('Alle 3 Amsterdammers');
    expect(getPresenterFilterSummary(0, 3)).toBe('Geen Amsterdammers geselecteerd');
    expect(getPresenterFilterSummary(1, 3)).toBe('1 van 3 Amsterdammers');
    expect(getPresenterFilterSummary(2, 3)).toBe('2 van 3 Amsterdammers');
  });
});
