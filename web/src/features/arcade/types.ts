export type ArcadeMode = 'attract' | 'manual';

export interface Ship {
  x: number;
  y: number;
  vx: number;
  vy: number;
  width: number;
  height: number;
  banking: number; // -1 to 1 roll tilt
  cooldown: number; // seconds until next shot allowed
  shields: number;
  maxShields: number;
}

export interface Projectile {
  id: number;
  x: number;
  y: number;
  vx: number;
  vy: number;
  radius: number;
  active: boolean;
}

export interface Hazard {
  id: number;
  x: number;
  y: number;
  vx: number;
  vy: number;
  radius: number;
  hp: number;
  maxHp: number;
  label: string;
  points: number;
  active: boolean;
}

export interface Particle {
  id: number;
  x: number;
  y: number;
  vx: number;
  vy: number;
  life: number;
  maxLife: number;
  color: string;
  size: number;
}

export interface Star {
  x: number;
  y: number;
  speed: number;
  size: number;
  brightness: number;
}

export interface ArcadeInput {
  up: boolean;
  down: boolean;
  left: boolean;
  right: boolean;
  fire: boolean;
}

export interface ArcadeState {
  mode: ArcadeMode;
  paused: boolean;
  ship: Ship;
  projectiles: Projectile[];
  hazards: Hazard[];
  particles: Particle[];
  stars: Star[];
  score: number;
  elapsedTime: number;
  targetsDestroyed: number;
  isDemo: boolean;
  nextEntityId: number;
  spawnTimer: number;
  screenShake: number;
}

export interface ArcadeConfig {
  canvasWidth: number;
  canvasHeight: number;
  fixedDt: number;
  reducedMotion: boolean;
}

export const LOGICAL_WIDTH = 800;
export const LOGICAL_HEIGHT = 480;
export const FIXED_DT = 1 / 60; // 60 Hz fixed simulation step
export const MAX_PROJECTILES = 40;
export const MAX_HAZARDS = 15;
export const MAX_PARTICLES = 70;
export const FIRE_COOLDOWN = 0.16; // ~6 shots/sec
export const SHIP_SPEED = 260;
export const PROJECTILE_SPEED = 580;

export const HAZARD_LABELS = [
  'Selection Bias',
  'Confounding',
  'Bad Denominator',
  'Spurious Correlation',
  'Missing Covariance',
] as const;
