import { ArcadeState, LOGICAL_HEIGHT, LOGICAL_WIDTH } from './types';

/**
 * Renders the entire Quant Flight arcade scene onto a 2D canvas context.
 */
export function renderArcade(
  ctx: CanvasRenderingContext2D,
  state: ArcadeState,
  reducedMotion = false
): void {
  ctx.save();

  // Screen shake translation
  if (!reducedMotion && state.screenShake > 0.1) {
    const shakeAngle = Math.random() * Math.PI * 2;
    const shakeDist = state.screenShake * 1.5;
    ctx.translate(Math.cos(shakeAngle) * shakeDist, Math.sin(shakeAngle) * shakeDist);
  }

  // 1. Deep space background
  ctx.fillStyle = '#050814';
  ctx.fillRect(0, 0, LOGICAL_WIDTH, LOGICAL_HEIGHT);

  // Subtle background gradient nebula
  const nebulaGrad = ctx.createRadialGradient(
    LOGICAL_WIDTH * 0.7,
    LOGICAL_HEIGHT * 0.4,
    40,
    LOGICAL_WIDTH * 0.7,
    LOGICAL_HEIGHT * 0.4,
    340
  );
  nebulaGrad.addColorStop(0, 'rgba(30, 58, 138, 0.18)');
  nebulaGrad.addColorStop(0.6, 'rgba(15, 23, 42, 0.08)');
  nebulaGrad.addColorStop(1, 'rgba(5, 8, 20, 0)');
  ctx.fillStyle = nebulaGrad;
  ctx.fillRect(0, 0, LOGICAL_WIDTH, LOGICAL_HEIGHT);

  // 2. Parallax Starfield
  for (const star of state.stars) {
    ctx.fillStyle = `rgba(255, 255, 255, ${star.brightness})`;
    ctx.beginPath();
    ctx.arc(star.x, star.y, star.size, 0, Math.PI * 2);
    ctx.fill();
  }

  // 3. Particles
  for (const p of state.particles) {
    const alpha = Math.max(0, p.life / p.maxLife);
    ctx.fillStyle = p.color;
    ctx.globalAlpha = alpha;
    ctx.beginPath();
    ctx.arc(p.x, p.y, p.size, 0, Math.PI * 2);
    ctx.fill();
  }
  ctx.globalAlpha = 1.0;

  // 4. Projectiles
  ctx.fillStyle = '#38bdf8';
  ctx.shadowColor = '#0284c7';
  ctx.shadowBlur = 8;
  for (const proj of state.projectiles) {
    ctx.beginPath();
    // Capsule energy bolt
    ctx.roundRect(proj.x - 8, proj.y - 2, 16, 4, 2);
    ctx.fill();
  }
  ctx.shadowBlur = 0;

  // 5. Hazards
  for (const haz of state.hazards) {
    ctx.save();
    ctx.translate(haz.x, haz.y);

    // Hazard body: faceted diamond / hexagon
    ctx.fillStyle = '#1e293b';
    ctx.strokeStyle = '#f59e0b';
    ctx.lineWidth = 2;
    ctx.beginPath();
    const sides = 6;
    for (let i = 0; i < sides; i++) {
      const angle = (i * Math.PI * 2) / sides;
      const px = Math.cos(angle) * haz.radius;
      const py = Math.sin(angle) * haz.radius;
      if (i === 0) ctx.moveTo(px, py);
      else ctx.lineTo(px, py);
    }
    ctx.closePath();
    ctx.fill();
    ctx.stroke();

    // Inner danger core
    ctx.fillStyle = '#ef4444';
    ctx.beginPath();
    ctx.arc(0, 0, haz.radius * 0.35, 0, Math.PI * 2);
    ctx.fill();

    // Label tag
    ctx.fillStyle = '#cbd5e1';
    ctx.font = '10px ui-sans-serif, system-ui, -apple-system, sans-serif';
    ctx.textAlign = 'center';
    ctx.fillText(haz.label, 0, -haz.radius - 6);

    ctx.restore();
  }

  // 6. Player / Autopilot Interceptor Ship
  const { ship } = state;
  ctx.save();
  ctx.translate(ship.x, ship.y);
  ctx.rotate(ship.banking);

  // Engine exhaust glow
  const glowGrad = ctx.createRadialGradient(-18, 0, 2, -26, 0, 16);
  glowGrad.addColorStop(0, 'rgba(56, 189, 248, 0.9)');
  glowGrad.addColorStop(1, 'rgba(56, 189, 248, 0)');
  ctx.fillStyle = glowGrad;
  ctx.beginPath();
  ctx.arc(-20, 0, 14, 0, Math.PI * 2);
  ctx.fill();

  // Ship hull (original geometric vector fighter)
  ctx.fillStyle = '#0f172a';
  ctx.strokeStyle = '#38bdf8';
  ctx.lineWidth = 2;

  ctx.beginPath();
  ctx.moveTo(22, 0); // Nose tip
  ctx.lineTo(-14, -13); // Upper wingtip
  ctx.lineTo(-9, -4); // Inner wing joint
  ctx.lineTo(-18, -3); // Upper engine block
  ctx.lineTo(-18, 3); // Lower engine block
  ctx.lineTo(-9, 4); // Inner wing joint
  ctx.lineTo(-14, 13); // Lower wingtip
  ctx.closePath();
  ctx.fill();
  ctx.stroke();

  // Cockpit canopy
  ctx.fillStyle = '#06b6d4';
  ctx.beginPath();
  ctx.moveTo(8, 0);
  ctx.lineTo(-4, -3.5);
  ctx.lineTo(-7, 0);
  ctx.lineTo(-4, 3.5);
  ctx.closePath();
  ctx.fill();

  ctx.restore();

  // 7. Heads-Up Display (HUD)
  renderHud(ctx, state);

  // 8. Attract / Paused Overlays
  if (state.mode === 'attract') {
    renderAttractBanner(ctx, state.elapsedTime);
  } else if (state.paused) {
    renderPausedOverlay(ctx);
  }

  ctx.restore();
}

/**
 * Draws the top HUD with mode, score, shields, and controls.
 */
function renderHud(ctx: CanvasRenderingContext2D, state: ArcadeState): void {
  // Top HUD Bar background
  ctx.fillStyle = 'rgba(15, 23, 42, 0.75)';
  ctx.fillRect(0, 0, LOGICAL_WIDTH, 36);
  ctx.strokeStyle = 'rgba(56, 189, 248, 0.25)';
  ctx.beginPath();
  ctx.moveTo(0, 36);
  ctx.lineTo(LOGICAL_WIDTH, 36);
  ctx.stroke();

  ctx.font = 'bold 12px ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace';

  // Mode badge
  if (state.isDemo) {
    ctx.fillStyle = '#eab308';
    ctx.fillText('⚡ DEMO / ATTRACT FLIGHT', 16, 23);
  } else {
    ctx.fillStyle = '#10b981';
    ctx.fillText('🚀 ACTIVE PILOT', 16, 23);
  }

  // Score
  ctx.fillStyle = '#94a3b8';
  ctx.fillText('SCORE: ', 220, 23);
  ctx.fillStyle = '#f8fafc';
  ctx.fillText(state.score.toString().padStart(6, '0'), 275, 23);

  // Targets
  ctx.fillStyle = '#94a3b8';
  ctx.fillText('HAZARDS CLEARED: ', 360, 23);
  ctx.fillStyle = '#38bdf8';
  ctx.fillText(state.targetsDestroyed.toString(), 495, 23);

  // Shields meter
  ctx.fillStyle = '#94a3b8';
  ctx.fillText('SHIELDS: ', 550, 23);

  const shieldWidth = 100;
  const shieldHeight = 10;
  const shieldX = 620;
  const shieldY = 14;

  ctx.fillStyle = '#334155';
  ctx.fillRect(shieldX, shieldY, shieldWidth, shieldHeight);

  const shieldPct = Math.max(0, state.ship.shields / state.ship.maxShields);
  ctx.fillStyle = shieldPct > 0.5 ? '#10b981' : shieldPct > 0.25 ? '#f59e0b' : '#ef4444';
  ctx.fillRect(shieldX, shieldY, shieldWidth * shieldPct, shieldHeight);

  ctx.strokeStyle = '#64748b';
  ctx.lineWidth = 1;
  ctx.strokeRect(shieldX, shieldY, shieldWidth, shieldHeight);

  // Small shortcut reminders
  ctx.fillStyle = '#64748b';
  ctx.font = '10px ui-sans-serif, system-ui, sans-serif';
  ctx.textAlign = 'right';
  ctx.fillText('P: Pause | Esc/Enter: Skip', LOGICAL_WIDTH - 16, 23);
  ctx.textAlign = 'left';
}

/**
 * Attract call-to-action overlay.
 */
function renderAttractBanner(ctx: CanvasRenderingContext2D, time: number): void {
  const pulse = (Math.sin(time * 3.5) + 1) / 2;
  const alpha = 0.85 + pulse * 0.15;

  ctx.save();
  ctx.fillStyle = `rgba(15, 23, 42, 0.8)`;
  ctx.roundRect(LOGICAL_WIDTH / 2 - 220, LOGICAL_HEIGHT - 65, 440, 44, 8);
  ctx.fill();
  ctx.strokeStyle = `rgba(56, 189, 248, ${alpha})`;
  ctx.lineWidth = 2;
  ctx.stroke();

  ctx.fillStyle = `rgba(255, 255, 255, ${alpha})`;
  ctx.font = 'bold 13px ui-sans-serif, system-ui, -apple-system, sans-serif';
  ctx.textAlign = 'center';
  ctx.fillText(
    'ATTRACT MODE — PRESS ARROWS / WASD / SPACE TO TAKE OVER',
    LOGICAL_WIDTH / 2,
    LOGICAL_HEIGHT - 38
  );
  ctx.restore();
}

/**
 * Paused state overlay.
 */
function renderPausedOverlay(ctx: CanvasRenderingContext2D): void {
  ctx.save();
  ctx.fillStyle = 'rgba(5, 8, 20, 0.65)';
  ctx.fillRect(0, 0, LOGICAL_WIDTH, LOGICAL_HEIGHT);

  ctx.fillStyle = '#f8fafc';
  ctx.font = 'bold 22px ui-sans-serif, system-ui, -apple-system, sans-serif';
  ctx.textAlign = 'center';
  ctx.fillText('⏸ SIMULATION PAUSED', LOGICAL_WIDTH / 2, LOGICAL_HEIGHT / 2 - 10);

  ctx.fillStyle = '#94a3b8';
  ctx.font = '13px ui-sans-serif, system-ui, -apple-system, sans-serif';
  ctx.fillText('Press P to resume, or Esc/Enter to return to practice', LOGICAL_WIDTH / 2, LOGICAL_HEIGHT / 2 + 20);
  ctx.restore();
}
