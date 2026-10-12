import {
  ArcadeState,
  LOGICAL_HEIGHT,
  LOGICAL_WIDTH,
  SECTORS,
} from './types';

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

  const sectorConfig = SECTORS[state.sector - 1] || SECTORS[0];

  // 1. Deep space background
  ctx.fillStyle = '#050814';
  ctx.fillRect(0, 0, LOGICAL_WIDTH, LOGICAL_HEIGHT);

  // Sector-themed background nebula gradient
  const nebulaGrad = ctx.createRadialGradient(
    LOGICAL_WIDTH * 0.65,
    LOGICAL_HEIGHT * 0.45,
    50,
    LOGICAL_WIDTH * 0.65,
    LOGICAL_HEIGHT * 0.45,
    380
  );
  nebulaGrad.addColorStop(0, sectorConfig.nebulaColor);
  nebulaGrad.addColorStop(0.5, 'rgba(15, 23, 42, 0.15)');
  nebulaGrad.addColorStop(1, 'rgba(5, 8, 20, 0)');
  ctx.fillStyle = nebulaGrad;
  ctx.fillRect(0, 0, LOGICAL_WIDTH, LOGICAL_HEIGHT);

  // 2. Parallax Starfield (Streaks during Warp)
  const isWarping = state.warpTimer > 0;
  for (const star of state.stars) {
    ctx.fillStyle = `rgba(255, 255, 255, ${star.brightness})`;
    if (isWarping && !reducedMotion) {
      ctx.strokeStyle = `rgba(224, 231, 255, ${Math.min(1, star.brightness * 1.5)})`;
      ctx.lineWidth = star.size;
      ctx.beginPath();
      ctx.moveTo(star.x, star.y);
      ctx.lineTo(star.x + star.speed * 0.18, star.y);
      ctx.stroke();
    } else {
      ctx.beginPath();
      ctx.arc(star.x, star.y, star.size, 0, Math.PI * 2);
      ctx.fill();
    }
  }

  // 3. Shockwaves (Smart Bomb Explosions)
  for (const sw of state.shockwaves) {
    if (!sw.active) continue;
    const progress = 1 - sw.duration / sw.maxDuration;
    const alpha = Math.max(0, (1 - progress) * 0.85);

    ctx.save();
    ctx.strokeStyle = `rgba(56, 189, 248, ${alpha})`;
    ctx.lineWidth = 4 * (1 - progress * 0.5);
    ctx.beginPath();
    ctx.arc(sw.x, sw.y, sw.radius, 0, Math.PI * 2);
    ctx.stroke();

    // Inner secondary shockwave ring
    if (sw.radius > 30) {
      ctx.strokeStyle = `rgba(244, 63, 94, ${alpha * 0.6})`;
      ctx.lineWidth = 2;
      ctx.beginPath();
      ctx.arc(sw.x, sw.y, sw.radius * 0.75, 0, Math.PI * 2);
      ctx.stroke();
    }
    ctx.restore();
  }

  // 4. Particles
  for (const p of state.particles) {
    const alpha = Math.max(0, p.life / p.maxLife);
    ctx.fillStyle = p.color;
    ctx.globalAlpha = alpha;
    ctx.beginPath();
    ctx.arc(p.x, p.y, p.size, 0, Math.PI * 2);
    ctx.fill();
  }
  ctx.globalAlpha = 1.0;

  // 5. Projectiles
  ctx.fillStyle = '#38bdf8';
  ctx.shadowColor = '#0284c7';
  ctx.shadowBlur = 8;
  for (const proj of state.projectiles) {
    ctx.beginPath();
    ctx.roundRect(proj.x - 8, proj.y - 2, 16, 4, 2);
    ctx.fill();
  }
  ctx.shadowBlur = 0;

  // 6. Hazards (Standard, Heavy, Fragment)
  for (const haz of state.hazards) {
    ctx.save();
    ctx.translate(haz.x, haz.y);

    if (haz.kind === 'heavy') {
      // Heavy Octagonal Armored Hazard
      ctx.fillStyle = '#261122';
      ctx.strokeStyle = '#f43f5e';
      ctx.lineWidth = 2.5;
      ctx.beginPath();
      const sides = 8;
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
      ctx.fillStyle = '#fb7185';
      ctx.beginPath();
      ctx.arc(0, 0, haz.radius * 0.4, 0, Math.PI * 2);
      ctx.fill();

      // HP Pips bar
      const pipWidth = 6;
      const pipGap = 3;
      const totalWidth = haz.maxHp * pipWidth + (haz.maxHp - 1) * pipGap;
      const startX = -totalWidth / 2;
      const pipY = -haz.radius - 12;

      for (let i = 0; i < haz.maxHp; i++) {
        ctx.fillStyle = i < haz.hp ? '#f43f5e' : '#475569';
        ctx.fillRect(startX + i * (pipWidth + pipGap), pipY, pipWidth, 3);
      }
    } else if (haz.kind === 'fragment') {
      // Sharp Fragment Shard
      ctx.fillStyle = '#1e1b4b';
      ctx.strokeStyle = '#fbbf24';
      ctx.lineWidth = 1.8;
      ctx.beginPath();
      ctx.moveTo(haz.radius, 0);
      ctx.lineTo(-haz.radius * 0.8, -haz.radius * 0.7);
      ctx.lineTo(-haz.radius * 0.4, 0);
      ctx.lineTo(-haz.radius * 0.8, haz.radius * 0.7);
      ctx.closePath();
      ctx.fill();
      ctx.stroke();

      ctx.fillStyle = '#f59e0b';
      ctx.beginPath();
      ctx.arc(0, 0, 3, 0, Math.PI * 2);
      ctx.fill();
    } else {
      // Standard Hexagonal Hazard
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
    }

    // Label tag
    ctx.fillStyle = '#cbd5e1';
    ctx.font = '10px ui-sans-serif, system-ui, -apple-system, sans-serif';
    ctx.textAlign = 'center';
    ctx.fillText(haz.label, 0, -haz.radius - (haz.kind === 'heavy' ? 16 : 6));

    ctx.restore();
  }

  // 7. Player / Autopilot Interceptor Ship
  const { ship } = state;
  if (!state.gameOver) {
    ctx.save();
    ctx.translate(ship.x, ship.y);
    ctx.rotate(ship.banking);

    // Invulnerability flicker
    if (ship.invulnerableTimer > 0 && Math.floor(ship.invulnerableTimer * 12) % 2 === 0) {
      ctx.globalAlpha = 0.4;
    }

    // Engine exhaust glow
    const glowGrad = ctx.createRadialGradient(-18, 0, 2, -26, 0, 16);
    glowGrad.addColorStop(0, isWarping ? 'rgba(168, 85, 247, 0.95)' : 'rgba(56, 189, 248, 0.9)');
    glowGrad.addColorStop(1, 'rgba(56, 189, 248, 0)');
    ctx.fillStyle = glowGrad;
    ctx.beginPath();
    ctx.arc(-20, 0, isWarping ? 20 : 14, 0, Math.PI * 2);
    ctx.fill();

    // Shield Aura bubble when shields active
    if (ship.shields > 10) {
      const shieldRatio = ship.shields / ship.maxShields;
      ctx.strokeStyle = `rgba(56, 189, 248, ${0.15 + shieldRatio * 0.25})`;
      ctx.lineWidth = 1.5;
      ctx.beginPath();
      ctx.ellipse(2, 0, 28, 20, 0, 0, Math.PI * 2);
      ctx.stroke();
    }

    // Ship hull (original vector interceptor)
    ctx.fillStyle = '#0f172a';
    ctx.strokeStyle = isWarping ? '#c084fc' : '#38bdf8';
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
    ctx.fillStyle = isWarping ? '#e879f9' : '#06b6d4';
    ctx.beginPath();
    ctx.moveTo(8, 0);
    ctx.lineTo(-4, -3.5);
    ctx.lineTo(-7, 0);
    ctx.lineTo(-4, 3.5);
    ctx.closePath();
    ctx.fill();

    ctx.restore();
  }

  // 8. Heads-Up Display (HUD)
  renderHud(ctx, state);

  // 9. Overlays (Warp, Attract, Paused, Game Over)
  if (state.warpTimer > 0) {
    renderWarpBanner(ctx, state.sector);
  } else if (state.gameOver) {
    renderGameOverOverlay(ctx, state);
  } else if (state.mode === 'attract') {
    renderAttractBanner(ctx, state.elapsedTime);
  } else if (state.paused) {
    renderPausedOverlay(ctx);
  }

  ctx.restore();
}

/**
 * Draws the top HUD with mode, score, sector progress, dual health meters, and bomb counts.
 */
function renderHud(ctx: CanvasRenderingContext2D, state: ArcadeState): void {
  // Top HUD Bar background
  ctx.fillStyle = 'rgba(15, 23, 42, 0.85)';
  ctx.fillRect(0, 0, LOGICAL_WIDTH, 42);
  ctx.strokeStyle = 'rgba(56, 189, 248, 0.25)';
  ctx.beginPath();
  ctx.moveTo(0, 42);
  ctx.lineTo(LOGICAL_WIDTH, 42);
  ctx.stroke();

  ctx.font = 'bold 11px ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace';

  // Mode badge
  if (state.isDemo) {
    ctx.fillStyle = '#eab308';
    ctx.fillText('⚡ ATTRACT', 14, 25);
  } else {
    ctx.fillStyle = '#10b981';
    ctx.fillText('🚀 PILOT', 14, 25);
  }

  // Score
  ctx.fillStyle = '#94a3b8';
  ctx.fillText('SCORE: ', 95, 25);
  ctx.fillStyle = '#f8fafc';
  ctx.fillText(state.score.toString().padStart(6, '0'), 145, 25);

  // Sector & Progress
  const sectorCfg = SECTORS[state.sector - 1] || SECTORS[0];
  ctx.fillStyle = sectorCfg.accentColor;
  ctx.fillText(`S${state.sector}: ${sectorCfg.name.toUpperCase()}`, 215, 25);

  // Sector Progress Mini Bar
  const progWidth = 50;
  const progHeight = 6;
  const progX = 390;
  const progY = 19;
  ctx.fillStyle = '#334155';
  ctx.fillRect(progX, progY, progWidth, progHeight);
  const progRatio = Math.min(1, state.sectorProgress / state.sectorTarget);
  ctx.fillStyle = sectorCfg.accentColor;
  ctx.fillRect(progX, progY, progWidth * progRatio, progHeight);
  ctx.strokeStyle = '#475569';
  ctx.strokeRect(progX, progY, progWidth, progHeight);

  // Dual Health Meters: Shields & Hull
  // Shields
  ctx.fillStyle = '#38bdf8';
  ctx.fillText('SHD:', 455, 25);
  const meterWidth = 55;
  const meterHeight = 8;
  const shdX = 485;
  const shdY = 17;
  ctx.fillStyle = '#1e293b';
  ctx.fillRect(shdX, shdY, meterWidth, meterHeight);
  const shdRatio = Math.max(0, state.ship.shields / state.ship.maxShields);
  ctx.fillStyle = '#38bdf8';
  ctx.fillRect(shdX, shdY, meterWidth * shdRatio, meterHeight);
  ctx.strokeStyle = '#475569';
  ctx.strokeRect(shdX, shdY, meterWidth, meterHeight);

  // Hull
  ctx.fillStyle = '#f43f5e';
  ctx.fillText('HUL:', 550, 25);
  const hulX = 580;
  const hulY = 17;
  ctx.fillStyle = '#1e293b';
  ctx.fillRect(hulX, hulY, meterWidth, meterHeight);
  const hulRatio = Math.max(0, state.ship.hull / state.ship.maxHull);
  ctx.fillStyle = hulRatio > 0.5 ? '#10b981' : hulRatio > 0.25 ? '#f59e0b' : '#ef4444';
  ctx.fillRect(hulX, hulY, meterWidth * hulRatio, meterHeight);
  ctx.strokeStyle = '#475569';
  ctx.strokeRect(hulX, hulY, meterWidth, meterHeight);

  // Smart Bombs Stock
  ctx.fillStyle = '#f59e0b';
  ctx.fillText('BOMB:', 648, 25);
  for (let b = 0; b < state.ship.maxBombs; b++) {
    const bx = 692 + b * 13;
    const by = 21;
    ctx.fillStyle = b < state.ship.bombs ? '#f59e0b' : '#334155';
    ctx.beginPath();
    // Diamond vector bomb icon
    ctx.moveTo(bx, by - 5);
    ctx.lineTo(bx + 4, by);
    ctx.lineTo(bx, by + 5);
    ctx.lineTo(bx - 4, by);
    ctx.closePath();
    ctx.fill();
  }

  // Key shortcuts prompt
  ctx.fillStyle = '#64748b';
  ctx.font = '10px ui-sans-serif, system-ui, sans-serif';
  ctx.textAlign = 'right';
  ctx.fillText('P: Pause | Esc: Skip', LOGICAL_WIDTH - 12, 25);
  ctx.textAlign = 'left';
}

/**
 * Warp Drive engaged banner.
 */
function renderWarpBanner(ctx: CanvasRenderingContext2D, sector: number): void {
  const sectorCfg = SECTORS[sector - 1] || SECTORS[0];
  ctx.save();
  ctx.fillStyle = 'rgba(15, 23, 42, 0.75)';
  ctx.roundRect(LOGICAL_WIDTH / 2 - 240, LOGICAL_HEIGHT / 2 - 40, 480, 80, 10);
  ctx.fill();
  ctx.strokeStyle = sectorCfg.accentColor;
  ctx.lineWidth = 2;
  ctx.stroke();

  ctx.fillStyle = '#f8fafc';
  ctx.font = 'bold 20px ui-sans-serif, system-ui, -apple-system, sans-serif';
  ctx.textAlign = 'center';
  ctx.fillText('⚡ HYPER-DRIVE ENGAGED ⚡', LOGICAL_WIDTH / 2, LOGICAL_HEIGHT / 2 - 10);

  ctx.fillStyle = sectorCfg.accentColor;
  ctx.font = 'bold 14px ui-monospace, monospace';
  ctx.fillText(
    `WARPING TO SECTOR ${sector}: ${sectorCfg.name.toUpperCase()}`,
    LOGICAL_WIDTH / 2,
    LOGICAL_HEIGHT / 2 + 18
  );
  ctx.restore();
}

/**
 * Attract call-to-action overlay.
 */
function renderAttractBanner(ctx: CanvasRenderingContext2D, time: number): void {
  const pulse = (Math.sin(time * 3.5) + 1) / 2;
  const alpha = 0.85 + pulse * 0.15;

  ctx.save();
  ctx.fillStyle = `rgba(15, 23, 42, 0.85)`;
  ctx.roundRect(LOGICAL_WIDTH / 2 - 250, LOGICAL_HEIGHT - 65, 500, 44, 8);
  ctx.fill();
  ctx.strokeStyle = `rgba(56, 189, 248, ${alpha})`;
  ctx.lineWidth = 2;
  ctx.stroke();

  ctx.fillStyle = `rgba(255, 255, 255, ${alpha})`;
  ctx.font = 'bold 12px ui-sans-serif, system-ui, -apple-system, sans-serif';
  ctx.textAlign = 'center';
  ctx.fillText(
    'ATTRACT DEMO — PRESS WASD / ARROWS / SPACE TO TAKE FLIGHT',
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
  ctx.fillStyle = 'rgba(5, 8, 20, 0.7)';
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

/**
 * Game Over overlay.
 */
function renderGameOverOverlay(ctx: CanvasRenderingContext2D, state: ArcadeState): void {
  ctx.save();
  ctx.fillStyle = 'rgba(15, 23, 42, 0.88)';
  ctx.fillRect(0, 0, LOGICAL_WIDTH, LOGICAL_HEIGHT);

  ctx.fillStyle = '#ef4444';
  ctx.font = 'bold 26px ui-sans-serif, system-ui, -apple-system, sans-serif';
  ctx.textAlign = 'center';
  ctx.fillText('💥 CRITICAL HULL COLLAPSE 💥', LOGICAL_WIDTH / 2, LOGICAL_HEIGHT / 2 - 50);

  ctx.fillStyle = '#f8fafc';
  ctx.font = 'bold 16px ui-monospace, monospace';
  ctx.fillText(
    `FINAL SCORE: ${state.score}   |   SECTOR REACHED: ${state.sector}`,
    LOGICAL_WIDTH / 2,
    LOGICAL_HEIGHT / 2 - 15
  );

  ctx.fillStyle = '#38bdf8';
  ctx.font = '14px ui-sans-serif, system-ui, -apple-system, sans-serif';
  ctx.fillText(
    `Total Hazards Cleared: ${state.targetsDestroyed}`,
    LOGICAL_WIDTH / 2,
    LOGICAL_HEIGHT / 2 + 15
  );

  ctx.fillStyle = '#94a3b8';
  ctx.font = '12px ui-sans-serif, system-ui, -apple-system, sans-serif';
  ctx.fillText(
    'Enter your initials below to record your flight, or press Esc to return to practice',
    LOGICAL_WIDTH / 2,
    LOGICAL_HEIGHT / 2 + 48
  );

  ctx.restore();
}
