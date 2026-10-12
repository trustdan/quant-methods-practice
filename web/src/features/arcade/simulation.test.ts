import { describe, it, expect } from 'vitest';
import {
  createInitialState,
  createRng,
  step,
  takeoverManual,
} from './simulation';
import {
  FIXED_DT,
  LOGICAL_HEIGHT,
  LOGICAL_WIDTH,
  PROJECTILE_SPEED,
  SHIP_SPEED,
} from './types';

describe('Quant Flight Arcade simulation', () => {
  it('initializes deterministically with a given seed', () => {
    const s1 = createInitialState(42, 'attract');
    const s2 = createInitialState(42, 'attract');

    expect(s1.ship.x).toBe(80);
    expect(s1.ship.y).toBe(LOGICAL_HEIGHT / 2);
    expect(s1.stars.length).toBe(s2.stars.length);
    expect(s1.stars[0].x).toBe(s2.stars[0].x);
    expect(s1.stars[0].y).toBe(s2.stars[0].y);
    expect(s1.isDemo).toBe(true);
    expect(s1.mode).toBe('attract');
  });

  it('runs deterministically across multiple fixed simulation steps', () => {
    const rng1 = createRng(100);
    const rng2 = createRng(100);

    let state1 = createInitialState(100, 'manual');
    let state2 = createInitialState(100, 'manual');

    const input = { up: false, down: false, left: false, right: true, fire: false };

    for (let i = 0; i < 30; i++) {
      state1 = step(state1, input, FIXED_DT, rng1);
      state2 = step(state2, input, FIXED_DT, rng2);
    }

    expect(state1.ship.x).toBeCloseTo(state2.ship.x, 5);
    expect(state1.ship.vx).toBeCloseTo(state2.ship.vx, 5);
    expect(state1.ship.x).toBeGreaterThan(80);
  });

  it('resets score and demo flag when takeoverManual is invoked', () => {
    const attract = createInitialState(99, 'attract');
    attract.score = 500;
    attract.targetsDestroyed = 5;
    expect(attract.isDemo).toBe(true);

    const manual = takeoverManual(attract);
    expect(manual.mode).toBe('manual');
    expect(manual.isDemo).toBe(false);
    expect(manual.score).toBe(0);
    expect(manual.targetsDestroyed).toBe(0);
    expect(manual.ship.x).toBeLessThan(LOGICAL_WIDTH);
  });

  it('automatically takes over from attract mode on any player input', () => {
    const rng = createRng(123);
    const attractState = createInitialState(123, 'attract');
    expect(attractState.mode).toBe('attract');
    expect(attractState.isDemo).toBe(true);

    // Player presses fire
    const nextState = step(
      attractState,
      { up: false, down: false, left: false, right: false, fire: true },
      FIXED_DT,
      rng
    );

    expect(nextState.mode).toBe('manual');
    expect(nextState.isDemo).toBe(false);
    expect(nextState.score).toBe(0);
    expect(nextState.targetsDestroyed).toBe(0);
  });

  it('clamps ship movement to logical canvas bounds', () => {
    const rng = createRng(1);
    let state = createInitialState(1, 'manual');

    // Steer left persistently
    for (let i = 0; i < 60; i++) {
      state = step(
        state,
        { up: false, down: false, left: true, right: false, fire: false },
        FIXED_DT,
        rng
      );
    }
    expect(state.ship.x).toBeGreaterThanOrEqual(20);

    // Steer up persistently
    for (let i = 0; i < 120; i++) {
      state = step(
        state,
        { up: true, down: false, left: false, right: false, fire: false },
        FIXED_DT,
        rng
      );
    }
    expect(state.ship.y).toBeGreaterThanOrEqual(25);
  });

  it('enforces weapon cooldown and spawns projectiles', () => {
    const rng = createRng(555);
    let state = createInitialState(555, 'manual');

    // First fire
    state = step(
      state,
      { up: false, down: false, left: false, right: false, fire: true },
      FIXED_DT,
      rng
    );
    expect(state.projectiles.length).toBe(1);
    expect(state.projectiles[0].vx).toBe(PROJECTILE_SPEED);
    expect(state.ship.cooldown).toBeGreaterThan(0);

    // Immediate next frame with fire still held should NOT spawn duplicate projectile (cooldown active)
    state = step(
      state,
      { up: false, down: false, left: false, right: false, fire: true },
      FIXED_DT,
      rng
    );
    expect(state.projectiles.length).toBe(1);
  });

  it('resolves projectile-hazard collision: awards points and clears target in manual mode', () => {
    const rng = createRng(777);
    const initial = createInitialState(777, 'manual');

    // Position projectile and hazard so they collide
    const stateWithEntities = {
      ...initial,
      projectiles: [
        {
          id: 10,
          x: 200,
          y: 200,
          vx: PROJECTILE_SPEED,
          vy: 0,
          radius: 4,
          active: true,
        },
      ],
      hazards: [
        {
          id: 20,
          kind: 'standard' as const,
          x: 210,
          y: 200,
          vx: -100,
          vy: 0,
          radius: 18,
          hp: 1,
          maxHp: 1,
          label: 'Selection Bias',
          points: 100,
          active: true,
        },
      ],
    };

    const nextState = step(
      stateWithEntities,
      { up: false, down: false, left: false, right: false, fire: false },
      FIXED_DT,
      rng
    );

    expect(nextState.score).toBe(100);
    expect(nextState.targetsDestroyed).toBe(1);
    expect(nextState.hazards.length).toBe(0);
    expect(nextState.projectiles.length).toBe(0);
    expect(nextState.particles.length).toBeGreaterThan(0);
  });

  it('does not award score points in attract demo mode', () => {
    const rng = createRng(888);
    const initial = createInitialState(888, 'attract');

    const stateWithEntities = {
      ...initial,
      projectiles: [
        {
          id: 10,
          x: 200,
          y: 200,
          vx: PROJECTILE_SPEED,
          vy: 0,
          radius: 4,
          active: true,
        },
      ],
      hazards: [
        {
          id: 20,
          kind: 'standard' as const,
          x: 210,
          y: 200,
          vx: -100,
          vy: 0,
          radius: 18,
          hp: 1,
          maxHp: 1,
          label: 'Confounding',
          points: 100,
          active: true,
        },
      ],
    };

    // No player input, staying in attract mode
    const nextState = step(
      stateWithEntities,
      { up: false, down: false, left: false, right: false, fire: false },
      FIXED_DT,
      rng
    );

    expect(nextState.targetsDestroyed).toBe(1);
    expect(nextState.score).toBe(0); // Demo points are 0
    expect(nextState.isDemo).toBe(true);
  });

  it('freezes simulation when paused', () => {
    const rng = createRng(999);
    const state = {
      ...createInitialState(999, 'manual'),
      paused: true,
      ship: {
        ...createInitialState(999, 'manual').ship,
        x: 100,
        vx: SHIP_SPEED,
      },
    };

    const nextState = step(
      state,
      { up: false, down: false, left: false, right: true, fire: true },
      FIXED_DT,
      rng
    );

    expect(nextState).toBe(state);
    expect(nextState.ship.x).toBe(100);
  });

  it('suppresses screen shake when reducedMotion is active', () => {
    const rng = createRng(444);
    const initial = createInitialState(444, 'manual');

    const stateWithCollision = {
      ...initial,
      ship: { ...initial.ship, x: 200, y: 200 },
      hazards: [
        {
          id: 20,
          kind: 'standard' as const,
          x: 205,
          y: 200,
          vx: -100,
          vy: 0,
          radius: 18,
          hp: 1,
          maxHp: 1,
          label: 'Confounding',
          points: 100,
          active: true,
        },
      ],
    };

    // Run with reducedMotion = true
    const nextState = step(
      stateWithCollision,
      { up: false, down: false, left: false, right: false, fire: false },
      FIXED_DT,
      rng,
      true
    );

    expect(nextState.screenShake).toBe(0);
    expect(nextState.ship.shields).toBeLessThan(100);
  });

  it('detonates smart bomb, creating shockwave, expending stock, and setting cooldown', () => {
    const rng = createRng(301);
    const initial = createInitialState(301, 'manual');
    expect(initial.ship.bombs).toBe(3);

    const nextState = step(
      initial,
      { up: false, down: false, left: false, right: false, fire: false, bomb: true },
      FIXED_DT,
      rng
    );

    expect(nextState.ship.bombs).toBe(2);
    expect(nextState.ship.bombCooldown).toBeGreaterThan(0);
    expect(nextState.shockwaves.length).toBe(1);
    expect(nextState.shockwaves[0].x).toBe(initial.ship.x);
    expect(nextState.shockwaves[0].y).toBe(initial.ship.y);
  });

  it('clears hazards and awards points when smart bomb shockwave expands', () => {
    const rng = createRng(302);
    const initial = createInitialState(302, 'manual');

    const stateWithBomb = {
      ...initial,
      shockwaves: [
        {
          id: 99,
          x: 200,
          y: 200,
          radius: 100,
          maxRadius: 300,
          duration: 0.4,
          maxDuration: 0.5,
          damage: 5,
          active: true,
        },
      ],
      hazards: [
        {
          id: 50,
          kind: 'standard' as const,
          x: 230,
          y: 200,
          vx: -50,
          vy: 0,
          radius: 18,
          hp: 1,
          maxHp: 1,
          label: 'Spurious Correlation',
          points: 100,
          active: true,
        },
      ],
    };

    const nextState = step(
      stateWithBomb,
      { up: false, down: false, left: false, right: false, fire: false },
      FIXED_DT,
      rng
    );

    expect(nextState.score).toBe(100);
    expect(nextState.targetsDestroyed).toBe(1);
    expect(nextState.hazards.length).toBe(0);
  });

  it('implements dual health: shields absorb first, then hull takes remaining damage', () => {
    const rng = createRng(303);
    const initial = createInitialState(303, 'manual');
    // Set ship with low shields
    const weakened = {
      ...initial,
      ship: {
        ...initial.ship,
        x: 200,
        y: 200,
        shields: 10,
        maxShields: 100,
        hull: 100,
        maxHull: 100,
        invulnerableTimer: 0,
        shieldRechargeTimer: 2.0,
      },
      hazards: [
        {
          id: 60,
          kind: 'standard' as const,
          x: 205,
          y: 200,
          vx: -80,
          vy: 0,
          radius: 18,
          hp: 1,
          maxHp: 1,
          label: 'Selection Bias',
          points: 100,
          active: true,
        },
      ],
    };

    const nextState = step(
      weakened,
      { up: false, down: false, left: false, right: false, fire: false },
      FIXED_DT,
      rng
    );

    // Standard damage is 24: 10 absorbed by shield, 14 taken by hull
    expect(nextState.ship.shields).toBe(0);
    expect(nextState.ship.hull).toBe(86);
    expect(nextState.gameOver).toBe(false);
  });

  it('triggers game over when hull reaches zero', () => {
    const rng = createRng(304);
    const initial = createInitialState(304, 'manual');
    const criticalShip = {
      ...initial,
      ship: {
        ...initial.ship,
        x: 200,
        y: 200,
        shields: 0,
        hull: 15,
        invulnerableTimer: 0,
      },
      hazards: [
        {
          id: 70,
          kind: 'standard' as const,
          x: 205,
          y: 200,
          vx: -80,
          vy: 0,
          radius: 18,
          hp: 1,
          maxHp: 1,
          label: 'Selection Bias',
          points: 100,
          active: true,
        },
      ],
    };

    const nextState = step(
      criticalShip,
      { up: false, down: false, left: false, right: false, fire: false },
      FIXED_DT,
      rng
    );

    expect(nextState.ship.hull).toBe(0);
    expect(nextState.gameOver).toBe(true);
  });

  it('splits heavy hazards into fragments when destroyed', () => {
    const rng = createRng(305);
    const initial = createInitialState(305, 'manual');

    // Heavy hazard with 1 HP left hit by projectile
    const stateWithHeavy = {
      ...initial,
      projectiles: [
        {
          id: 11,
          x: 200,
          y: 200,
          vx: PROJECTILE_SPEED,
          vy: 0,
          radius: 4,
          active: true,
        },
      ],
      hazards: [
        {
          id: 80,
          kind: 'heavy' as const,
          x: 215,
          y: 200,
          vx: -90,
          vy: 0,
          radius: 26,
          hp: 1,
          maxHp: 3,
          label: 'Collider Bias',
          points: 250,
          active: true,
        },
      ],
    };

    const nextState = step(
      stateWithHeavy,
      { up: false, down: false, left: false, right: false, fire: false },
      FIXED_DT,
      rng
    );

    expect(nextState.score).toBe(250);
    expect(nextState.targetsDestroyed).toBe(1);
    // Should have spawned 2 fragments
    const fragments = nextState.hazards.filter((h) => h.kind === 'fragment');
    expect(fragments.length).toBe(2);
    expect(fragments[0].points).toBe(75);
    expect(fragments[0].radius).toBe(13);
  });

  it('triggers sector warp, replenishes bombs and shields, and awards bonus on sector completion', () => {
    const rng = createRng(306);
    const initial = createInitialState(306, 'manual');

    // State at threshold of sector 1 completion (target is 10)
    const atSectorThreshold = {
      ...initial,
      sector: 1,
      sectorProgress: 9,
      sectorTarget: 10,
      score: 900,
      ship: {
        ...initial.ship,
        shields: 50,
        bombs: 2,
      },
      projectiles: [
        {
          id: 15,
          x: 300,
          y: 200,
          vx: PROJECTILE_SPEED,
          vy: 0,
          radius: 4,
          active: true,
        },
      ],
      hazards: [
        {
          id: 95,
          kind: 'standard' as const,
          x: 310,
          y: 200,
          vx: -80,
          vy: 0,
          radius: 18,
          hp: 1,
          maxHp: 1,
          label: 'Confounding',
          points: 100,
          active: true,
        },
      ],
    };

    const nextState = step(
      atSectorThreshold,
      { up: false, down: false, left: false, right: false, fire: false },
      FIXED_DT,
      rng
    );

    // Hazard points (100) + Sector 1 clear bonus (1000) = 900 + 100 + 1000 = 2000
    expect(nextState.score).toBe(2000);
    expect(nextState.sector).toBe(2);
    expect(nextState.sectorProgress).toBe(0);
    expect(nextState.warpTimer).toBeGreaterThan(0);
    expect(nextState.ship.shields).toBe(100); // Fully restored
    expect(nextState.ship.bombs).toBe(3); // Replenished +1 bomb
  });

  it('gradually recharges shields when unhit for the recharge delay', () => {
    const rng = createRng(307);
    const initial = createInitialState(307, 'manual');

    let state = {
      ...initial,
      ship: {
        ...initial.ship,
        shields: 50,
        maxShields: 100,
        shieldRechargeTimer: 0, // Delay elapsed
      },
    };

    // Step several times without any collisions
    for (let i = 0; i < 30; i++) {
      state = step(
        state,
        { up: false, down: false, left: false, right: false, fire: false },
        FIXED_DT,
        rng
      );
    }

    expect(state.ship.shields).toBeGreaterThan(50);
  });
});

