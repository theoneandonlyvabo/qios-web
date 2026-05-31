function isFinder(row: number, col: number) {
  const inTopLeft = row < 7 && col < 7;
  const inTopRight = row < 7 && col > 13;
  const inBottomLeft = row > 13 && col < 7;
  if (!(inTopLeft || inTopRight || inBottomLeft)) return false;

  const localRow = row < 7 ? row : row - 14;
  const localCol = col < 7 ? col : col - 14;
  const edge = localRow === 0 || localRow === 6 || localCol === 0 || localCol === 6;
  const center = localRow >= 2 && localRow <= 4 && localCol >= 2 && localCol <= 4;
  return edge || center;
}

function isFinderWhite(row: number, col: number) {
  const inTopLeft = row < 7 && col < 7;
  const inTopRight = row < 7 && col > 13;
  const inBottomLeft = row > 13 && col < 7;
  if (!(inTopLeft || inTopRight || inBottomLeft)) return false;

  const localRow = row < 7 ? row : row - 14;
  const localCol = col < 7 ? col : col - 14;
  return localRow >= 1 && localRow <= 5 && localCol >= 1 && localCol <= 5 && !(localRow >= 2 && localRow <= 4 && localCol >= 2 && localCol <= 4);
}

export function FakeQrCode() {
  const cells = Array.from({ length: 21 * 21 }, (_, index) => {
    const row = Math.floor(index / 21);
    const col = index % 21;
    if (isFinderWhite(row, col)) return false;
    if (isFinder(row, col)) return true;
    return ((row * 7 + col * 11 + row * col) % 5 === 0 || (row + col) % 7 === 0) && row !== 10;
  });

  return (
    <div className="rounded-2xl bg-white p-4 shadow-2xl">
      <div className="qr-grid h-full w-full">
        {cells.map((active, index) => (
          <span key={index} className={`qr-cell ${active ? "bg-black" : "bg-white"}`} />
        ))}
      </div>
    </div>
  );
}
