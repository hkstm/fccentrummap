type Props = {
  color: string;
  label: string;
  selected?: boolean;
};

export function AmsterdamXMarker({ color, label, selected = false }: Props) {
  return (
    <div
      className={selected ? 'markerButton markerButtonSelected' : 'markerButton'}
      role="img"
      aria-label={`Open spot details: ${label}`}
      title={label}
    >
      <span className="markerHitbox" aria-hidden="true" />
      <svg viewBox="0 0 40 40" aria-hidden="true">
        <g transform="translate(20,20)">
          {selected && (
            <>
              <rect className="markerHalo" x={-4} y={-22} width={8} height={44} rx={2} fill="none" stroke="white" strokeWidth={8} transform="rotate(45)" />
              <rect className="markerHalo" x={-4} y={-22} width={8} height={44} rx={2} fill="none" stroke="white" strokeWidth={8} transform="rotate(-45)" />
            </>
          )}
          <rect x={-4} y={-22} width={8} height={44} rx={2} fill={color} transform="rotate(45)" />
          <rect x={-4} y={-22} width={8} height={44} rx={2} fill={color} transform="rotate(-45)" />
        </g>
      </svg>
    </div>
  );
}
