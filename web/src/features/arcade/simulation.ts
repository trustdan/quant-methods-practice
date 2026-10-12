import {
  ArcadeInput,
  ArcadeMode,
  ArcadeState,
  BOMB_COOLDOWN,
  BOMB_DURATION,
  BOMB_MAX_RADIUS,
  FIRE_COOLDOWN,
  HAZARD_LABELS,
  HEAVY_HAZARD_LABELS,
  Hazard,
  LOGICAL_HEIGHT,
  LOGICAL_WIDTH,
  MAX_HAZARDS,
  MAX_PARTICLES,
  MAX_PROJECTILES,
  Particle,
  PROJECTILE_SPEED,
  Projectile,
  SECTORS,
  SHIP_SPEED,
  Ship,
  Shockwave,
  Star,
} from './types';

/**
 * Deterministic pseudo-random number generator (Mulberry32).
 */
export function createRng(seed: number): () => number {
  let s = Math.floor(seed) >>> 0;
  return () => {
    s = (s + 0x6d2b79f5) >>> 0;
    let t = Math.imul(s ^ (s >>> 15), 1 | s);
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

/**
 * Generates initial starfield distribution.
 */
function createInitialStars(rng: () => number, count = 80): Star[] {
  const stars: Star[] = [];
  for (let i = 0; i < count; i++) {
    const layer = rng();
    let speed = 40;
    let size = 1;
    let brightness = 0.4 + rng() * 0.3;

    if (layer > 0.75) {
      // Near / fast stars
      speed = 230;
      size = 2.2;
      brightness = 0.9 + rng() * 0.1;
    } else if (layer > 0.4) {
      // Mid stars
      speed = 115;
      size = 1.5;
      brightness = 0.6 + rng() * 0.3;
    }

    stars.push({
      x: rng() * LOGICAL_WIDTH,
      y: rng() * LOGICAL_HEIGHT,
      speed,
      size,
      brightness,
    });
  }
  return stars;
}

/**
 * Creates a fresh initial game state.
 */
export function createInitialState(seed = 1337, mode: ArcadeMode = 'attract'): ArcadeState {
  const rng = createRng(seed);
  const ship: Ship = {
    x: 80,
    y: LOGICAL_HEIGHT / 2,
    vx: 0,
    vy: 0,
    width: 38,
    height: 22,
    banking: 0,
    cooldown: 0,
    shields: 100,
    maxShields: 100,
    hull: 100,
    maxHull: 100,
    bombs: 3,
    maxBombs: 3,
    bombCooldown: 0,
    shieldRechargeTimer: 0,
    invulnerableTimer: 0,
  };

  const initialSector = 1;
  const sectorConfig = SECTORS[initialSector - 1] || SECTORS[0];

  return {
    mode,
    paused: false,
    ship,
    projectiles: [],
    hazards: [],
    shockwaves: [],
    particles: [],
    stars: createInitialStars(rng),
    score: 0,
    elapsedTime: 0,
    targetsDestroyed: 0,
    sector: initialSector,
    sectorProgress: 0,
    sectorTarget: sectorConfig.targetHazards,
    warpTimer: 0,
    gameOver: false,
    highScore: 0,
    isDemo: mode === 'attract',
    nextEntityId: 1,
    spawnTimer: 0.8,
    screenShake: 0,
  };
}

/**
 * Transitions immediately from attract demo mode to player manual mode.
 * Resets score, elapsed timer, and target count so demo play never inflates human records.
 */
export function takeoverManual(state: ArcadeState): ArcadeState {
  if (state.mode === 'manual' && !state.gameOver) return state;

  const initialSector = 1;
  const sectorConfig = SECTORS[initialSector - 1] || SECTORS[0];

  return {
    ...state,
    mode: 'manual',
    isDemo: false,
    score: 0,
    elapsedTime: 0,
    targetsDestroyed: 0,
    sector: initialSector,
    sectorProgress: 0,
    sectorTarget: sectorConfig.targetHazards,
    warpTimer: 0,
    gameOver: false,
    projectiles: [],
    hazards: [],
    shockwaves: [],
    particles: [],
    ship: {
      ...state.ship,
      x: 80,
      y: LOGICAL_HEIGHT / 2,
      vx: 0,
      vy: 0,
      banking: 0,
      cooldown: 0,
      shields: 100,
      maxShields: 100,
      hull: 100,
      maxHull: 100,
      bombs: 3,
      maxBombs: 3,
      bombCooldown: 0,
      shieldRechargeTimer: 0,
      invulnerableTimer: 1.0, // 1s spawn protection
    },
    spawnTimer: 0.5,
    screenShake: 0,
  };
}

/**
 * Computes autopilot artificial input during attract demo mode.
 */
function computeAttractInput(state: ArcadeState, elapsedTime: number): ArcadeInput {
  const { ship, hazards } = state;
  let up = false;
  let down = false;
  let left = false;
  let right = false;
  let fire = false;
  let bomb = false;

  // Gentle sinusoidal cruising base
  const targetY = LOGICAL_HEIGHT / 2 + Math.sin(elapsedTime * 1.8) * 110;

  // Look for oncoming threats in range
  const oncoming = hazards
    .filter((h) => h.active && h.x > ship.x)
    .sort((a, b) => a.x - b.x);

  const closest = oncoming[0];

  if (closest) {
    const dy = closest.y - ship.y;
    if (Math.abs(dy) < 45) {
      fire = true;
    }
    // Evasive maneuvers
    if (dy > 12) down = true;
    else if (dy < -12) up = true;

    // Use smart bomb if multiple threats are crowding nearby
    if (oncoming.length >= 3 && closest.x - ship.x < 150 && ship.bombs > 0 && ship.bombCooldown <= 0) {
      bomb = true;
    }
  } else {
    if (ship.y < targetY - 15) down = true;
    else if (ship.y > targetY + 15) up = true;
    // Periodic autofire demonstration
    if (Math.sin(elapsedTime * 4.5) > 0.2) fire = true;
  }

  // Position maintenance
  if (ship.x > 180) left = true;
  else if (ship.x < 70) right = true;

  return {
    up,
    down,
    left,
    right,
    fire,
    bomb,
  };
}

/**
 * Pure simulation step advancing the game by fixed time delta `dt`.
 */
export function step(
  state: ArcadeState,
  rawInput: ArcadeInput,
  dt: number,
  rng: () => number,
  reducedMotion = false
): ArcadeState {
  if (state.paused) {
    return state;
  }

  // Handle game over in manual mode vs attract demo
  if (state.gameOver) {
    // In attract mode, automatically restart attract flight after brief pause
    if (state.isDemo) {
      return createInitialState(Math.floor(rng() * 100000), 'attract');
    }
    return state;
  }

  // Check if player took over during attract mode
  const playerAction =
    rawInput.up ||
    rawInput.down ||
    rawInput.left ||
    rawInput.right ||
    rawInput.fire ||
    rawInput.bomb;

  let current = state;
  if (state.mode === 'attract' && playerAction) {
    current = takeoverManual(state);
  }

  const effectiveInput =
    current.mode === 'attract'
      ? computeAttractInput(current, current.elapsedTime + dt)
      : rawInput;

  let nextEntityId = current.nextEntityId;
  const currentSectorConfig = SECTORS[current.sector - 1] || SECTORS[0];

  // 1. Advance Warp Transition if active
  let warpTimer = Math.max(0, current.warpTimer - dt);
  const isWarping = warpTimer > 0;

  // 2. Advance Ship state
  const ship = { ...current.ship };
  ship.cooldown = Math.max(0, ship.cooldown - dt);
  ship.bombCooldown = Math.max(0, ship.bombCooldown - dt);
  ship.invulnerableTimer = Math.max(0, ship.invulnerableTimer - dt);

  // Passive Shield Recharge: if unhit for delay period, recharge gradually
  if (ship.shieldRechargeTimer > 0) {
    ship.shieldRechargeTimer = Math.max(0, ship.shieldRechargeTimer - dt);
  } else if (ship.shields < ship.maxShields) {
    ship.shields = Math.min(ship.maxShields, ship.shields + 12 * dt);
  }

  // Velocity targets with inertia damping
  let targetVx = 0;
  let targetVy = 0;

  if (effectiveInput.left) targetVx -= SHIP_SPEED;
  if (effectiveInput.right) targetVx += SHIP_SPEED;
  if (effectiveInput.up) targetVy -= SHIP_SPEED;
  if (effectiveInput.down) targetVy += SHIP_SPEED;

  // 8-way diagonal velocity normalization
  if (targetVx !== 0 && targetVy !== 0) {
    const invSqrt2 = 0.70710678;
    targetVx *= invSqrt2;
    targetVy *= invSqrt2;
  }

  // Responsive acceleration & deceleration curves
  const accelRate = targetVx !== 0 || targetVy !== 0 ? 15 : 10;
  ship.vx = ship.vx + (targetVx - ship.vx) * Math.min(1, dt * accelRate);
  ship.vy = ship.vy + (targetVy - ship.vy) * Math.min(1, dt * accelRate);

  ship.x = Math.max(20, Math.min(LOGICAL_WIDTH - 60, ship.x + ship.vx * dt));
  ship.y = Math.max(25, Math.min(LOGICAL_HEIGHT - 25, ship.y + ship.vy * dt));

  // Banking roll tilt (-0.35 up, +0.35 down)
  const targetBanking = (ship.vy / SHIP_SPEED) * 0.35;
  ship.banking += (targetBanking - ship.banking) * Math.min(1, dt * 10);

  // 3. Projectile firing (Autofire held fire key)
  const newProjectiles: Projectile[] = [];
  if (effectiveInput.fire && ship.cooldown <= 0 && !isWarping) {
    ship.cooldown = FIRE_COOLDOWN;
    newProjectiles.push({
      id: nextEntityId++,
      x: ship.x + ship.width / 2 + 6,
      y: ship.y,
      vx: PROJECTILE_SPEED,
      vy: ship.vy * 0.08,
      radius: 4,
      active: true,
    });
  }

  // Move existing projectiles
  const updatedProjectiles: Projectile[] = [];
  for (const p of current.projectiles) {
    if (!p.active) continue;
    const nx = p.x + p.vx * dt;
    const ny = p.y + p.vy * dt;
    if (nx < LOGICAL_WIDTH + 20 && ny >= 0 && ny <= LOGICAL_HEIGHT) {
      updatedProjectiles.push({ ...p, x: nx, y: ny });
    }
  }
  for (const np of newProjectiles) {
    if (updatedProjectiles.length < MAX_PROJECTILES) {
      updatedProjectiles.push(np);
    }
  }

  // 4. Smart Bomb Detonation
  let screenShake = Math.max(0, current.screenShake - dt * 3.5);
  const newShockwaves: Shockwave[] = [];
  const newParticles: Particle[] = [];

  if (
    effectiveInput.bomb &&
    ship.bombs > 0 &&
    ship.bombCooldown <= 0 &&
    !isWarping
  ) {
    ship.bombs -= 1;
    ship.bombCooldown = BOMB_COOLDOWN;
    newShockwaves.push({
      id: nextEntityId++,
      x: ship.x,
      y: ship.y,
      radius: 10,
      maxRadius: BOMB_MAX_RADIUS,
      duration: BOMB_DURATION,
      maxDuration: BOMB_DURATION,
      damage: 4,
      active: true,
    });

    if (!reducedMotion) {
      screenShake = Math.min(10, screenShake + 6);
    }

    // Detonation shockwave particle ring
    const ringParticles = reducedMotion ? 8 : 24;
    for (let i = 0; i < ringParticles; i++) {
      const angle = (i * Math.PI * 2) / ringParticles;
      const speed = 120 + rng() * 100;
      newParticles.push({
        id: nextEntityId++,
        x: ship.x,
        y: ship.y,
        vx: Math.cos(angle) * speed,
        vy: Math.sin(angle) * speed,
        life: 0.35 + rng() * 0.2,
        maxLife: 0.55,
        color: '#38bdf8',
        size: 3 + rng() * 2,
      });
    }
  }

  // Update existing shockwaves
  const updatedShockwaves: Shockwave[] = [];
  for (const sw of current.shockwaves) {
    if (!sw.active) continue;
    const nextDur = sw.duration - dt;
    if (nextDur > 0) {
      const progress = 1 - nextDur / sw.maxDuration;
      const nextRadius = 10 + progress * (sw.maxRadius - 10);
      updatedShockwaves.push({
        ...sw,
        duration: nextDur,
        radius: nextRadius,
      });
    }
  }
  for (const nsw of newShockwaves) {
    updatedShockwaves.push(nsw);
  }

  // 5. Parallax Starfield Movement (accelerates 3.5x during warp)
  const starSpeedMult = isWarping ? 3.5 : 1.0;
  const stars: Star[] = current.stars.map((star) => {
    let nx = star.x - star.speed * dt * starSpeedMult;
    let ny = star.y;
    if (nx < 0) {
      nx = LOGICAL_WIDTH + rng() * 10;
      ny = rng() * LOGICAL_HEIGHT;
    }
    return { ...star, x: nx, y: ny };
  });

  // 6. Hazard Spawning & Movement
  let spawnTimer = current.spawnTimer - dt;
  const updatedHazards: Hazard[] = [];

  for (const h of current.hazards) {
    if (!h.active) continue;
    const nx = h.x + h.vx * dt * (isWarping ? 2.0 : 1.0);
    const ny = h.y + h.vy * dt;
    if (nx > -60 && ny >= 10 && ny <= LOGICAL_HEIGHT - 10) {
      updatedHazards.push({ ...h, x: nx, y: ny });
    }
  }

  // Spawn new hazard if not currently warping
  if (!isWarping && spawnTimer <= 0 && updatedHazards.length < MAX_HAZARDS) {
    const baseInterval = Math.max(0.65, 1.3 - (current.sector - 1) * 0.18);
    spawnTimer = baseInterval + rng() * 0.9;

    const isHeavy = rng() < currentSectorConfig.heavySpawnChance;
    const speed = (95 + rng() * 65) * currentSectorConfig.speedMultiplier;

    if (isHeavy) {
      const labelIdx = Math.floor(rng() * HEAVY_HAZARD_LABELS.length);
      const label = HEAVY_HAZARD_LABELS[labelIdx] || 'Heavy Anomaly';
      updatedHazards.push({
        id: nextEntityId++,
        kind: 'heavy',
        x: LOGICAL_WIDTH + 45,
        y: 45 + rng() * (LOGICAL_HEIGHT - 90),
        vx: -speed * 0.85,
        vy: (rng() - 0.5) * 25,
        radius: 26,
        hp: 3,
        maxHp: 3,
        label,
        points: 250,
        active: true,
      });
    } else {
      const labelIdx = Math.floor(rng() * HAZARD_LABELS.length);
      const label = HAZARD_LABELS[labelIdx] || 'Bias Hazard';
      updatedHazards.push({
        id: nextEntityId++,
        kind: 'standard',
        x: LOGICAL_WIDTH + 35,
        y: 40 + rng() * (LOGICAL_HEIGHT - 80),
        vx: -speed,
        vy: (rng() - 0.5) * 35,
        radius: 18,
        hp: 1,
        maxHp: 1,
        label,
        points: 100,
        active: true,
      });
    }
  }

  // 7. Collisions & Resolutions
  let score = current.score;
  let targetsDestroyed = current.targetsDestroyed;
  let sectorProgress = current.sectorProgress;
  let sector = current.sector;
  let sectorTarget = current.sectorTarget;
  let gameOver = false;

  const spawnFragmentsAndExplosions = (haz: Hazard) => {
    targetsDestroyed++;
    sectorProgress++;
    if (!current.isDemo) {
      score += haz.points;
    }

    if (!reducedMotion) {
      screenShake = Math.min(8, screenShake + (haz.kind === 'heavy' ? 4 : 2));
    }

    // Heavy hazards split into two divergent fragments
    if (haz.kind === 'heavy') {
      const fragLabels = ['Bias Fragment α', 'Bias Fragment β'];
      const fragVys = [-65, 65];
      for (let i = 0; i < 2; i++) {
        if (updatedHazards.length < MAX_HAZARDS) {
          updatedHazards.push({
            id: nextEntityId++,
            kind: 'fragment',
            x: haz.x,
            y: Math.max(20, Math.min(LOGICAL_HEIGHT - 20, haz.y + (i === 0 ? -15 : 15))),
            vx: haz.vx * 1.25,
            vy: fragVys[i] + (rng() - 0.5) * 20,
            radius: 13,
            hp: 1,
            maxHp: 1,
            label: fragLabels[i] || 'Fragment',
            points: 75,
            active: true,
          });
        }
      }
    }

    // Explosion particles
    const particleCount = reducedMotion ? 4 : haz.kind === 'heavy' ? 18 : 10;
    for (let p = 0; p < particleCount; p++) {
      const angle = rng() * Math.PI * 2;
      const speed = 40 + rng() * 130;
      newParticles.push({
        id: nextEntityId++,
        x: haz.x,
        y: haz.y,
        vx: Math.cos(angle) * speed + haz.vx * 0.2,
        vy: Math.sin(angle) * speed,
        life: 0.3 + rng() * 0.35,
        maxLife: 0.65,
        color: haz.kind === 'heavy' ? '#f43f5e' : p % 2 === 0 ? '#38bdf8' : '#f59e0b',
        size: haz.kind === 'heavy' ? 3 + rng() * 3 : 2 + rng() * 2.5,
      });
    }
  };

  // Shockwave vs Hazards
  for (const sw of updatedShockwaves) {
    if (!sw.active) continue;
    for (const haz of updatedHazards) {
      if (!haz.active) continue;
      const dx = sw.x - haz.x;
      const dy = sw.y - haz.y;
      const dist = Math.sqrt(dx * dx + dy * dy);

      if (dist <= sw.radius + haz.radius) {
        haz.hp -= sw.damage;
        if (haz.hp <= 0) {
          haz.active = false;
          spawnFragmentsAndExplosions(haz);
        }
      }
    }
  }

  // Projectile vs Hazard
  for (const proj of updatedProjectiles) {
    if (!proj.active) continue;
    for (const haz of updatedHazards) {
      if (!haz.active) continue;
      const dx = proj.x - haz.x;
      const dy = proj.y - haz.y;
      const distSq = dx * dx + dy * dy;
      const minDist = proj.radius + haz.radius;

      if (distSq <= minDist * minDist) {
        proj.active = false;
        haz.hp -= 1;

        if (haz.hp <= 0) {
          haz.active = false;
          spawnFragmentsAndExplosions(haz);
        } else {
          // Impact spark on heavy armor
          for (let sp = 0; sp < 3; sp++) {
            newParticles.push({
              id: nextEntityId++,
              x: proj.x,
              y: proj.y,
              vx: -80 - rng() * 60,
              vy: (rng() - 0.5) * 80,
              life: 0.15 + rng() * 0.1,
              maxLife: 0.25,
              color: '#38bdf8',
              size: 2,
            });
          }
        }
        break;
      }
    }
  }

  // Sector Progression check: advance sector upon reaching target
  if (sectorProgress >= sectorTarget && warpTimer <= 0) {
    warpTimer = 2.4; // 2.4s warp transition
    sectorProgress = 0;
    const oldSector = sector;
    sector = (sector % 4) + 1; // 1 -> 2 -> 3 -> 4 -> 1 loop
    const nextConfig = SECTORS[sector - 1] || SECTORS[0];
    sectorTarget = nextConfig.targetHazards;

    // Sector Completion Bonus & Replenishments
    if (!current.isDemo) {
      score += 1000 * oldSector;
    }
    ship.bombs = Math.min(ship.maxBombs, ship.bombs + 1);
    ship.shields = ship.maxShields;
    ship.invulnerableTimer = 2.8; // Invulnerable during warp

    if (!reducedMotion) {
      screenShake = Math.min(10, screenShake + 5);
    }
  }

  // Ship vs Hazard (Dual Health: Shields absorb first, then Hull)
  for (const haz of updatedHazards) {
    if (!haz.active) continue;
    const dx = ship.x - haz.x;
    const dy = ship.y - haz.y;
    const distSq = dx * dx + dy * dy;
    const minDist = 18 + haz.radius;

    if (distSq <= minDist * minDist) {
      haz.active = false;

      // Handle ship damage if not currently invulnerable
      if (ship.invulnerableTimer <= 0 && !isWarping) {
        const damage = haz.kind === 'heavy' ? 40 : haz.kind === 'fragment' ? 14 : 24;

        if (ship.shields > 0) {
          if (ship.shields >= damage) {
            ship.shields -= damage;
          } else {
            const bleedthrough = damage - ship.shields;
            ship.shields = 0;
            ship.hull = Math.max(0, ship.hull - bleedthrough);
          }
        } else {
          ship.hull = Math.max(0, ship.hull - damage);
        }

        // Damage resets shield recharge timer and grants brief hit grace
        ship.shieldRechargeTimer = 4.0;
        ship.invulnerableTimer = 0.4;

        if (!reducedMotion) {
          screenShake = Math.min(9, screenShake + 5);
        }

        // Spark particles
        const sparkCount = reducedMotion ? 3 : 9;
        for (let p = 0; p < sparkCount; p++) {
          const angle = rng() * Math.PI * 2;
          const speed = 60 + rng() * 100;
          newParticles.push({
            id: nextEntityId++,
            x: ship.x,
            y: ship.y,
            vx: Math.cos(angle) * speed,
            vy: Math.sin(angle) * speed,
            life: 0.2 + rng() * 0.25,
            maxLife: 0.45,
            color: ship.shields > 0 ? '#38bdf8' : '#ef4444',
            size: 2.2 + rng() * 2,
          });
        }

        // Check Hull Destruction / Game Over
        if (ship.hull <= 0) {
          gameOver = true;
          // Hull destruction explosion
          for (let p = 0; p < 30; p++) {
            const angle = rng() * Math.PI * 2;
            const speed = 50 + rng() * 180;
            newParticles.push({
              id: nextEntityId++,
              x: ship.x,
              y: ship.y,
              vx: Math.cos(angle) * speed,
              vy: Math.sin(angle) * speed,
              life: 0.4 + rng() * 0.4,
              maxLife: 0.8,
              color: p % 3 === 0 ? '#ef4444' : p % 3 === 1 ? '#f59e0b' : '#38bdf8',
              size: 3 + rng() * 3,
            });
          }
        }
      }
    }
  }

  // Engine exhaust trail particles
  if (!reducedMotion && !gameOver && current.particles.length < MAX_PARTICLES) {
    newParticles.push({
      id: nextEntityId++,
      x: ship.x - ship.width / 2,
      y: ship.y + (rng() - 0.5) * 4,
      vx: isWarping ? -350 - rng() * 120 : -130 - rng() * 60,
      vy: (rng() - 0.5) * 20,
      life: isWarping ? 0.3 + rng() * 0.15 : 0.18 + rng() * 0.12,
      maxLife: isWarping ? 0.45 : 0.3,
      color: isWarping ? '#a855f7' : '#38bdf8',
      size: isWarping ? 2.5 + rng() * 2 : 1.8 + rng() * 1.5,
    });
  }

  // Update existing particles
  const updatedParticles: Particle[] = [];
  for (const pt of current.particles) {
    const nextLife = pt.life - dt;
    if (nextLife > 0) {
      updatedParticles.push({
        ...pt,
        x: pt.x + pt.vx * dt,
        y: pt.y + pt.vy * dt,
        life: nextLife,
      });
    }
  }
  for (const np of newParticles) {
    if (updatedParticles.length < MAX_PARTICLES) {
      updatedParticles.push(np);
    }
  }

  return {
    ...current,
    ship,
    projectiles: updatedProjectiles.filter((p) => p.active),
    hazards: updatedHazards.filter((h) => h.active),
    shockwaves: updatedShockwaves,
    particles: updatedParticles,
    stars,
    score,
    elapsedTime: current.elapsedTime + dt,
    targetsDestroyed,
    sector,
    sectorProgress,
    sectorTarget,
    warpTimer,
    gameOver,
    nextEntityId,
    spawnTimer,
    screenShake: reducedMotion ? 0 : screenShake,
  };
}
