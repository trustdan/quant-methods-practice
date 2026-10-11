import {
  ArcadeInput,
  ArcadeMode,
  ArcadeState,
  FIRE_COOLDOWN,
  HAZARD_LABELS,
  Hazard,
  LOGICAL_HEIGHT,
  LOGICAL_WIDTH,
  MAX_HAZARDS,
  MAX_PARTICLES,
  MAX_PROJECTILES,
  Particle,
  PROJECTILE_SPEED,
  Projectile,
  SHIP_SPEED,
  Ship,
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
function createInitialStars(rng: () => number, count = 75): Star[] {
  const stars: Star[] = [];
  for (let i = 0; i < count; i++) {
    const layer = rng();
    let speed = 40;
    let size = 1;
    let brightness = 0.4 + rng() * 0.3;

    if (layer > 0.75) {
      // Near / fast stars
      speed = 220;
      size = 2.2;
      brightness = 0.9 + rng() * 0.1;
    } else if (layer > 0.4) {
      // Mid stars
      speed = 110;
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
  };

  return {
    mode,
    paused: false,
    ship,
    projectiles: [],
    hazards: [],
    particles: [],
    stars: createInitialStars(rng),
    score: 0,
    elapsedTime: 0,
    targetsDestroyed: 0,
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
  if (state.mode === 'manual') return state;

  return {
    ...state,
    mode: 'manual',
    isDemo: false,
    score: 0,
    elapsedTime: 0,
    targetsDestroyed: 0,
    // Clear ongoing demo projectiles and hazards for a clean player start
    projectiles: [],
    hazards: [],
    spawnTimer: 0.5,
  };
}

/**
 * Computes autopilot artificial input during attract demo mode.
 */
function computeAttractInput(state: ArcadeState, elapsedTime: number): ArcadeInput {
  const { ship, hazards } = state;
  let up = false;
  let down = false;
  let fire = false;

  // Gentle sinusoidal cruising base
  const targetY = LOGICAL_HEIGHT / 2 + Math.sin(elapsedTime * 1.8) * 110;

  // Look for oncoming threats in range
  const oncoming = hazards
    .filter((h) => h.active && h.x > ship.x)
    .sort((a, b) => a.x - b.x)[0];

  if (oncoming) {
    const dy = oncoming.y - ship.y;
    if (Math.abs(dy) < 40) {
      fire = true;
    }
    if (dy > 12) down = true;
    else if (dy < -12) up = true;
  } else {
    if (ship.y < targetY - 15) down = true;
    else if (ship.y > targetY + 15) up = true;
    // Occasional demo fire
    if (Math.sin(elapsedTime * 4) > 0.5) fire = true;
  }

  return {
    up,
    down,
    left: false,
    right: false,
    fire,
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

  // Check if player took over during attract mode
  const playerAction =
    rawInput.up ||
    rawInput.down ||
    rawInput.left ||
    rawInput.right ||
    rawInput.fire;

  let current = state;
  if (state.mode === 'attract' && playerAction) {
    current = takeoverManual(state);
  }

  const effectiveInput =
    current.mode === 'attract'
      ? computeAttractInput(current, current.elapsedTime + dt)
      : rawInput;

  let nextEntityId = current.nextEntityId;

  // 1. Advance ship
  const ship = { ...current.ship };
  let targetVx = 0;
  let targetVy = 0;

  if (effectiveInput.left) targetVx -= SHIP_SPEED;
  if (effectiveInput.right) targetVx += SHIP_SPEED;
  if (effectiveInput.up) targetVy -= SHIP_SPEED;
  if (effectiveInput.down) targetVy += SHIP_SPEED;

  // Diagonal velocity normalization
  if (targetVx !== 0 && targetVy !== 0) {
    const invSqrt2 = 0.70710678;
    targetVx *= invSqrt2;
    targetVy *= invSqrt2;
  }

  // Smooth movement acceleration
  ship.vx = ship.vx + (targetVx - ship.vx) * Math.min(1, dt * 14);
  ship.vy = ship.vy + (targetVy - ship.vy) * Math.min(1, dt * 14);

  ship.x = Math.max(20, Math.min(LOGICAL_WIDTH - 60, ship.x + ship.vx * dt));
  ship.y = Math.max(25, Math.min(LOGICAL_HEIGHT - 25, ship.y + ship.vy * dt));

  // Banking roll tilt (-0.35 up, +0.35 down)
  const targetBanking = (ship.vy / SHIP_SPEED) * 0.35;
  ship.banking += (targetBanking - ship.banking) * Math.min(1, dt * 10);

  // Weapon cooldown
  ship.cooldown = Math.max(0, ship.cooldown - dt);

  // Projectiles
  const newProjectiles: Projectile[] = [];
  if (effectiveInput.fire && ship.cooldown <= 0) {
    ship.cooldown = FIRE_COOLDOWN;
    newProjectiles.push({
      id: nextEntityId++,
      x: ship.x + ship.width / 2 + 6,
      y: ship.y,
      vx: PROJECTILE_SPEED,
      vy: ship.vy * 0.1,
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

  // 2. Starfield parallax movement
  const stars: Star[] = current.stars.map((star) => {
    let nx = star.x - star.speed * dt;
    let ny = star.y;
    if (nx < 0) {
      nx = LOGICAL_WIDTH + rng() * 10;
      ny = rng() * LOGICAL_HEIGHT;
    }
    return { ...star, x: nx, y: ny };
  });

  // 3. Hazard spawning & movement
  let spawnTimer = current.spawnTimer - dt;
  const updatedHazards: Hazard[] = [];

  for (const h of current.hazards) {
    if (!h.active) continue;
    const nx = h.x + h.vx * dt;
    const ny = h.y + h.vy * dt;
    if (nx > -50 && ny >= 10 && ny <= LOGICAL_HEIGHT - 10) {
      updatedHazards.push({ ...h, x: nx, y: ny });
    }
  }

  if (spawnTimer <= 0 && updatedHazards.length < MAX_HAZARDS) {
    spawnTimer = 1.2 + rng() * 1.1;
    const labelIdx = Math.floor(rng() * HAZARD_LABELS.length);
    const label = HAZARD_LABELS[labelIdx] || 'Confounding';
    const speed = 90 + rng() * 70;

    updatedHazards.push({
      id: nextEntityId++,
      x: LOGICAL_WIDTH + 35,
      y: 40 + rng() * (LOGICAL_HEIGHT - 80),
      vx: -speed,
      vy: (rng() - 0.5) * 30,
      radius: 18,
      hp: 1,
      maxHp: 1,
      label,
      points: 100,
      active: true,
    });
  }

  // 4. Collision detection and resolutions
  let score = current.score;
  let targetsDestroyed = current.targetsDestroyed;
  let screenShake = Math.max(0, current.screenShake - dt * 3);
  const newParticles: Particle[] = [];

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
          targetsDestroyed++;
          if (!current.isDemo) {
            score += haz.points;
          }

          if (!reducedMotion) {
            screenShake = Math.min(6, screenShake + 2.5);
          }

          // Explosion particles
          const particleCount = reducedMotion ? 4 : 12;
          for (let p = 0; p < particleCount; p++) {
            const angle = rng() * Math.PI * 2;
            const speed = 40 + rng() * 120;
            newParticles.push({
              id: nextEntityId++,
              x: haz.x,
              y: haz.y,
              vx: Math.cos(angle) * speed + haz.vx * 0.2,
              vy: Math.sin(angle) * speed,
              life: 0.3 + rng() * 0.3,
              maxLife: 0.6,
              color: p % 2 === 0 ? '#38bdf8' : '#f59e0b',
              size: 2 + rng() * 2.5,
            });
          }
        }
        break;
      }
    }
  }

  // Ship vs Hazard
  for (const haz of updatedHazards) {
    if (!haz.active) continue;
    const dx = ship.x - haz.x;
    const dy = ship.y - haz.y;
    const distSq = dx * dx + dy * dy;
    const minDist = 18 + haz.radius;

    if (distSq <= minDist * minDist) {
      haz.active = false;
      ship.shields = Math.max(0, ship.shields - 15);
      if (!reducedMotion) {
        screenShake = Math.min(8, screenShake + 4);
      }

      // Spark particles
      const particleCount = reducedMotion ? 3 : 8;
      for (let p = 0; p < particleCount; p++) {
        const angle = rng() * Math.PI * 2;
        const speed = 50 + rng() * 90;
        newParticles.push({
          id: nextEntityId++,
          x: ship.x,
          y: ship.y,
          vx: Math.cos(angle) * speed,
          vy: Math.sin(angle) * speed,
          life: 0.2 + rng() * 0.25,
          maxLife: 0.45,
          color: '#ef4444',
          size: 2 + rng() * 2,
        });
      }
    }
  }

  // Engine exhaust trail particles
  if (!reducedMotion && current.particles.length < MAX_PARTICLES) {
    newParticles.push({
      id: nextEntityId++,
      x: ship.x - ship.width / 2,
      y: ship.y + (rng() - 0.5) * 4,
      vx: -120 - rng() * 60,
      vy: (rng() - 0.5) * 20,
      life: 0.18 + rng() * 0.12,
      maxLife: 0.3,
      color: '#38bdf8',
      size: 1.8 + rng() * 1.5,
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
    particles: updatedParticles,
    stars,
    score,
    elapsedTime: current.elapsedTime + dt,
    targetsDestroyed,
    nextEntityId,
    spawnTimer,
    screenShake: reducedMotion ? 0 : screenShake,
  };
}
