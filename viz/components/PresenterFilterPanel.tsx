type Props = {
  presenters: string[];
  selectedPresenters: Set<string>;
  setPresenter: (name: string, checked: boolean) => void;
  selectAll: () => void;
  deselectAll: () => void;
};

export function PresenterFilterPanel({ presenters, selectedPresenters, setPresenter, selectAll, deselectAll }: Props) {
  return (
    <aside className="panel" aria-label="Presenter filters">
      <details open>
        <summary className="panelHeader"><h2>Presenters</h2></summary>

        <div className="panelActions">
          <button type="button" onClick={selectAll}>Select all</button>
          <button type="button" onClick={deselectAll}>Deselect all</button>
        </div>

        <ul className="presenterList">
          {presenters.map((name) => {
            const checked = selectedPresenters.has(name);
            return (
              <li key={name}>
                <label>
                  <input
                    type="checkbox"
                    checked={checked}
                    onChange={(event) => setPresenter(name, event.target.checked)}
                  />
                  <span>{name}</span>
                </label>
              </li>
            );
          })}
        </ul>
      </details>
    </aside>
  );
}
