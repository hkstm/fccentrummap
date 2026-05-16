import { describe, expect, it } from 'vitest';
import { PRESENTER_PALETTE, buildPresenterColorMap } from './color';

describe('buildPresenterColorMap', () => {
  it('assigns colors from the given presenter/filter order instead of alphabetical order', () => {
    const presenters = ['Zulu', 'Alpha', 'Mike'];

    expect(buildPresenterColorMap(presenters)).toEqual({
      Zulu: PRESENTER_PALETTE[0],
      Alpha: PRESENTER_PALETTE[1],
      Mike: PRESENTER_PALETTE[2],
    });
  });

  it('is stable for repeated builds with the same presenter list', () => {
    const presenters = ['Ray Fuego', 'Sef', 'Akwasi', 'Winne'];

    expect(buildPresenterColorMap(presenters)).toEqual(buildPresenterColorMap(presenters));
  });

  it('wraps the fixed palette in a deterministic sequence', () => {
    const presenters = Array.from({ length: PRESENTER_PALETTE.length + 2 }, (_, idx) => `Presenter ${idx + 1}`);
    const colors = buildPresenterColorMap(presenters);

    expect(new Set(PRESENTER_PALETTE).size).toBe(PRESENTER_PALETTE.length);
    expect(colors[presenters[0]]).toBe(PRESENTER_PALETTE[0]);
    expect(colors[presenters[PRESENTER_PALETTE.length - 1]]).toBe(PRESENTER_PALETTE[PRESENTER_PALETTE.length - 1]);
    expect(colors[presenters[PRESENTER_PALETTE.length]]).toBe(PRESENTER_PALETTE[0]);
    expect(colors[presenters[PRESENTER_PALETTE.length + 1]]).toBe(PRESENTER_PALETTE[1]);
  });
});
