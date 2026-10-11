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
});
