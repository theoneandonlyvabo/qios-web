"use client";

import { useEffect, useRef, useState } from "react";
import { Icon } from "@/components/icons";

const HOLD_MS = 800;

export function SlideToConfirm({
  disabled = false,
  onConfirm
}: {
  disabled?: boolean;
  onConfirm: () => void;
}) {
  const trackRef = useRef<HTMLDivElement | null>(null);
  const [isPressed, setIsPressed] = useState(false);
  const [progress, setProgress] = useState(0);
  const [holdProgress, setHoldProgress] = useState(0);
  const [completed, setCompleted] = useState(false);
  const [maxOffset, setMaxOffset] = useState(0);

  useEffect(() => {
    if (!isPressed || progress < 0.92 || completed || disabled) {
      if (progress < 0.92) setHoldProgress(0);
      return;
    }

    const startedAt = Date.now();
    const interval = window.setInterval(() => {
      const next = Math.min(1, (Date.now() - startedAt) / HOLD_MS);
      setHoldProgress(next);
      if (next >= 1) {
        window.clearInterval(interval);
        setCompleted(true);
        window.setTimeout(onConfirm, 180);
      }
    }, 16);

    return () => window.clearInterval(interval);
  }, [completed, disabled, isPressed, onConfirm, progress]);

  function updateFromPointer(clientX: number) {
    const track = trackRef.current;
    if (!track || disabled) return;
    const rect = track.getBoundingClientRect();
    const handleWidth = 56;
    const nextMaxOffset = Math.max(0, rect.width - handleWidth - 16);
    setMaxOffset(nextMaxOffset);
    const raw = (clientX - rect.left - handleWidth / 2) / Math.max(1, nextMaxOffset);
    setProgress(Math.max(0, Math.min(1, raw)));
  }

  function reset() {
    if (completed) return;
    setIsPressed(false);
    setProgress(0);
    setHoldProgress(0);
  }

  return (
    <div
      ref={trackRef}
      className={`relative h-16 overflow-hidden rounded-full border bg-card-highest p-2 shadow-inner ${
        disabled ? "border-border opacity-60" : "border-border-warm/50"
      }`}
      onPointerMove={(event) => {
        if (isPressed) updateFromPointer(event.clientX);
      }}
      onPointerUp={reset}
      onPointerCancel={reset}
      onPointerLeave={() => {
        if (isPressed && progress < 0.92) reset();
      }}
    >
      <div
        className="absolute inset-y-2 left-2 rounded-full bg-primary/20 transition-[width]"
        style={{ width: `${progress * 100}%` }}
      />
      {holdProgress > 0 && (
        <div
          className="absolute inset-y-2 left-2 rounded-full bg-success/20 transition-[width]"
          style={{ width: `${holdProgress * 100}%` }}
        />
      )}

      <div className="pointer-events-none absolute inset-0 flex items-center justify-center">
        <span className="text-[11px] font-extrabold uppercase tracking-[0.18em] text-muted-foreground/55">
          {completed ? "Pembayaran dikonfirmasi" : holdProgress > 0 ? "Tahan sebentar..." : "Geser untuk Konfirmasi"}
        </span>
      </div>

      <button
        aria-label="Geser untuk konfirmasi pembayaran"
        disabled={disabled || completed}
        className={`absolute left-2 top-2 flex h-12 w-12 touch-none items-center justify-center rounded-full shadow-lg transition-colors ${
          completed ? "bg-success text-white" : "bg-primary text-primary-foreground"
        }`}
        style={{ transform: `translateX(${progress * maxOffset}px)` }}
        onPointerDown={(event) => {
          if (disabled || completed) return;
          event.currentTarget.setPointerCapture(event.pointerId);
          setIsPressed(true);
          updateFromPointer(event.clientX);
        }}
        type="button"
      >
        <Icon name={completed ? "check" : "chevron-right"} className="h-5 w-5" />
      </button>
    </div>
  );
}
