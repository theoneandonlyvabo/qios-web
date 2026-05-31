export function QiosLogo({ compact = false }: { compact?: boolean }) {
  return (
    <div className="flex items-center gap-3">
      <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-br from-brand to-primary text-lg font-black text-white shadow-[0_0_24px_rgba(202,64,10,0.28)]">
        Q
      </div>
      {!compact && (
        <div className="leading-tight">
          <p className="text-lg font-extrabold tracking-tight text-foreground">QIOS</p>
          <p className="text-[10px] font-bold uppercase tracking-[0.22em] text-muted-foreground">Skalar Solutions</p>
        </div>
      )}
    </div>
  );
}
