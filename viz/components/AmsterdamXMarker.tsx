type Props = {
  color: string;
  label: string;
};

export function AmsterdamXMarker({ color, label }: Props) {
  return (
    <div className="markerButton" role="img" aria-label={`Open spot details: ${label}`} title={label}>
      <svg viewBox="0 0 60 90" aria-hidden="true">
        <g transform="translate(30,42)">
          <rect x={-4} y={-22} width={8} height={44} rx={2} fill={color} transform="rotate(45)" />
          <rect x={-4} y={-22} width={8} height={44} rx={2} fill={color} transform="rotate(-45)" />
        </g>
      </svg>
    </div>
  );
}
