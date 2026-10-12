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
  hull: number;
  maxHull: number;
  bombs: number;
  maxBombs: number;
  bombCooldown: number;
  shieldRechargeTimer: number;
  invulnerableTimer: number;
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

export type HazardKind = 'standard' | 'heavy' | 'fragment';

export interface Hazard {
  id: number;
  kind: HazardKind;
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

export interface Shockwave {
  id: number;
  x: number;
  y: number;
  radius: number;
  maxRadius: number;
  duration: number;
  maxDuration: number;
  damage: number;
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
  bomb?: boolean;
}

export interface SectorConfig {
  sector: number;
  name: string;
  targetHazards: number;
  speedMultiplier: number;
  heavySpawnChance: number;
  accentColor: string;
  nebulaColor: string;
}

export interface ArcadeState {
  mode: ArcadeMode;
  paused: boolean;
  ship: Ship;
  projectiles: Projectile[];
  hazards: Hazard[];
  shockwaves: Shockwave[];
  particles: Particle[];
  stars: Star[];
  score: number;
  elapsedTime: number;
  targetsDestroyed: number;
  sector: number;
  sectorProgress: number;
  sectorTarget: number;
  warpTimer: number;
  gameOver: boolean;
  highScore: number;
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

export interface ArcadeScoreEntry {
  id?: number;
  game_version: string;
  seed: number;
  is_demo: boolean;
  duration_seconds: number;
  sectors_cleared: number;
  score: number;
  initials: string;
  created_at?: string;
}

export const LOGICAL_WIDTH = 800;
export const LOGICAL_HEIGHT = 480;
export const FIXED_DT = 1 / 60; // 60 Hz fixed simulation step
export const MAX_PROJECTILES = 45;
export const MAX_HAZARDS = 20;
export const MAX_PARTICLES = 120;
export const FIRE_COOLDOWN = 0.16; // ~6.2 shots/sec
export const BOMB_COOLDOWN = 1.0;
export const BOMB_MAX_RADIUS = 340;
export const BOMB_DURATION = 0.55;
export const SHIP_SPEED = 280;
export const PROJECTILE_SPEED = 600;

export const SECTORS: readonly SectorConfig[] = [
  {
    sector: 1,
    name: 'Univariate Void',
    targetHazards: 10,
    speedMultiplier: 1.0,
    heavySpawnChance: 0.15,
    accentColor: '#38bdf8',
    nebulaColor: 'rgba(30, 58, 138, 0.22)',
  },
  {
    sector: 2,
    name: 'Bivariate Belt',
    targetHazards: 14,
    speedMultiplier: 1.18,
    heavySpawnChance: 0.35,
    accentColor: '#a855f7',
    nebulaColor: 'rgba(107, 33, 168, 0.22)',
  },
  {
    sector: 3,
    name: 'Sampling Nebula',
    targetHazards: 18,
    speedMultiplier: 1.35,
    heavySpawnChance: 0.5,
    accentColor: '#f59e0b',
    nebulaColor: 'rgba(180, 83, 9, 0.22)',
  },
  {
    sector: 4,
    name: 'Inference Core',
    targetHazards: 22,
    speedMultiplier: 1.5,
    heavySpawnChance: 0.65,
    accentColor: '#ef4444',
    nebulaColor: 'rgba(190, 18, 60, 0.25)',
  },
] as const;

export const HAZARD_LABELS = [
  'Selection Bias',
  'Confounding',
  'Bad Denominator',
  'Spurious Correlation',
  'Missing Covariance',
] as const;

export const HEAVY_HAZARD_LABELS = [
  'Collider Bias',
  'Simpson Paradox',
  'Omitted Variable',
  'Endogeneity Matrix',
  'Heteroskedasticity',
] as const;

