/**
 * Pan-zoom utility for SVG/HTML viewports.
 * Mouse drag to pan, scroll-wheel zoom toward cursor (clamped 0.3x-3x), touch pinch.
 * Returns a cleanup function to remove all listeners.
 */

export interface PanZoomOptions {
  minScale?: number;
  maxScale?: number;
  /** Initial scale */
  scale?: number;
  /** Show hint text on first interaction */
  hint?: boolean;
}

export interface PanZoomState {
  x: number;
  y: number;
  scale: number;
}

export function initPanZoom(
  viewport: HTMLElement,
  world: HTMLElement,
  opts: PanZoomOptions = {},
): { cleanup: () => void; getState: () => PanZoomState; reset: () => void; setState: (newState: Partial<PanZoomState>, animate?: boolean) => void } {
  const minScale = opts.minScale ?? 0.3;
  const maxScale = opts.maxScale ?? 3;
  const state: PanZoomState = { x: 0, y: 0, scale: opts.scale ?? 1 };

  let isDragging = false;
  let startX = 0;
  let startY = 0;
  let startPanX = 0;
  let startPanY = 0;

  // Touch pinch state
  let lastPinchDist = 0;

  function applyTransform() {
    world.style.transform = `translate(${state.x}px, ${state.y}px) scale(${state.scale})`;
  }

  function clampScale(s: number): number {
    return Math.min(maxScale, Math.max(minScale, s));
  }

  // Hint overlay
  let hintEl: HTMLElement | null = null;
  if (opts.hint !== false) {
    hintEl = document.createElement('div');
    hintEl.textContent = 'Scroll to zoom \u00b7 Drag to pan';
    hintEl.style.cssText =
      'position:absolute;bottom:12px;left:50%;transform:translateX(-50%);' +
      'font-size:0.75rem;color:var(--text-muted);opacity:0.7;pointer-events:none;' +
      'font-family:var(--font-mono);transition:opacity 0.5s;z-index:5;';
    viewport.style.position = viewport.style.position || 'relative';
    viewport.appendChild(hintEl);
  }

  function dismissHint() {
    if (hintEl) {
      hintEl.style.opacity = '0';
      setTimeout(() => hintEl?.remove(), 500);
      hintEl = null;
    }
  }

  // --- Mouse drag ---
  function onPointerDown(e: PointerEvent) {
    if (e.button !== 0) return;
    isDragging = true;
    startX = e.clientX;
    startY = e.clientY;
    startPanX = state.x;
    startPanY = state.y;
    viewport.style.cursor = 'grabbing';
    viewport.setPointerCapture(e.pointerId);
    dismissHint();
  }

  function onPointerMove(e: PointerEvent) {
    if (!isDragging) return;
    state.x = startPanX + (e.clientX - startX);
    state.y = startPanY + (e.clientY - startY);
    applyTransform();
  }

  function onPointerUp(e: PointerEvent) {
    if (!isDragging) return;
    isDragging = false;
    viewport.style.cursor = 'grab';
    viewport.releasePointerCapture(e.pointerId);
  }

  // --- Scroll-wheel zoom toward cursor ---
  function onWheel(e: WheelEvent) {
    e.preventDefault();
    dismissHint();

    const rect = viewport.getBoundingClientRect();
    const cx = e.clientX - rect.left;
    const cy = e.clientY - rect.top;

    const factor = e.deltaY < 0 ? 1.1 : 1 / 1.1;
    const newScale = clampScale(state.scale * factor);
    const ratio = newScale / state.scale;

    // Zoom toward cursor
    state.x = cx - ratio * (cx - state.x);
    state.y = cy - ratio * (cy - state.y);
    state.scale = newScale;

    applyTransform();
  }

  // --- Touch pinch ---
  function pinchDist(e: TouchEvent): number {
    const [a, b] = [e.touches[0], e.touches[1]];
    return Math.hypot(b.clientX - a.clientX, b.clientY - a.clientY);
  }

  function onTouchStart(e: TouchEvent) {
    if (e.touches.length === 2) {
      lastPinchDist = pinchDist(e);
      dismissHint();
    }
  }

  function onTouchMove(e: TouchEvent) {
    if (e.touches.length === 2) {
      e.preventDefault();
      const dist = pinchDist(e);
      const factor = dist / lastPinchDist;
      const newScale = clampScale(state.scale * factor);

      const rect = viewport.getBoundingClientRect();
      const cx =
        (e.touches[0].clientX + e.touches[1].clientX) / 2 - rect.left;
      const cy =
        (e.touches[0].clientY + e.touches[1].clientY) / 2 - rect.top;
      const ratio = newScale / state.scale;

      state.x = cx - ratio * (cx - state.x);
      state.y = cy - ratio * (cy - state.y);
      state.scale = newScale;

      lastPinchDist = dist;
      applyTransform();
    }
  }

  // Set initial cursor and transform
  viewport.style.cursor = 'grab';
  viewport.style.overflow = 'hidden';
  world.style.transformOrigin = '0 0';
  applyTransform();

  // Bind events
  viewport.addEventListener('pointerdown', onPointerDown);
  viewport.addEventListener('pointermove', onPointerMove);
  viewport.addEventListener('pointerup', onPointerUp);
  viewport.addEventListener('pointercancel', onPointerUp);
  viewport.addEventListener('wheel', onWheel, { passive: false });
  viewport.addEventListener('touchstart', onTouchStart, { passive: true });
  viewport.addEventListener('touchmove', onTouchMove, { passive: false });

  function cleanup() {
    viewport.removeEventListener('pointerdown', onPointerDown);
    viewport.removeEventListener('pointermove', onPointerMove);
    viewport.removeEventListener('pointerup', onPointerUp);
    viewport.removeEventListener('pointercancel', onPointerUp);
    viewport.removeEventListener('wheel', onWheel);
    viewport.removeEventListener('touchstart', onTouchStart);
    viewport.removeEventListener('touchmove', onTouchMove);
    hintEl?.remove();
  }

  function reset() {
    state.x = 0;
    state.y = 0;
    state.scale = opts.scale ?? 1;
    applyTransform();
  }

  function setState(newState: Partial<PanZoomState>, animate = false) {
    if (newState.x !== undefined) state.x = newState.x;
    if (newState.y !== undefined) state.y = newState.y;
    if (newState.scale !== undefined) state.scale = clampScale(newState.scale);
    if (animate) {
      world.style.transition = 'transform 0.5s cubic-bezier(0.4, 0, 0.2, 1)';
      applyTransform();
      const onEnd = () => { world.style.transition = ''; world.removeEventListener('transitionend', onEnd); };
      world.addEventListener('transitionend', onEnd);
    } else {
      applyTransform();
    }
  }

  return { cleanup, getState: () => ({ ...state }), reset, setState };
}
