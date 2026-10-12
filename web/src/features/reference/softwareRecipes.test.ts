import { describe, it, expect } from 'vitest';
import fs from 'node:fs';
import path from 'node:path';
import {
  SOFTWARE_RECIPES,
  TEMPLATE_TO_RECIPE_MAP,
  getAllRecipes,
  getRecipeById,
  getRecipesByModule,
  getRecipesByFamily,
  getRecipeForTemplate,
} from './softwareRecipes';

describe('Software Recipe Catalog (Phase 1)', () => {
  it('registers comprehensive recipes across all modules with required properties', () => {
    const recipes = getAllRecipes();
    expect(recipes).toBe(SOFTWARE_RECIPES);
    expect(recipes.length).toBeGreaterThanOrEqual(30);

    const recipeIds = new Set<string>();

    for (const recipe of recipes) {
      // Unique ID
      expect(recipe.id).toBeTruthy();
      expect(recipeIds.has(recipe.id)).toBe(false);
      recipeIds.add(recipe.id);

      // Metadata
      expect(recipe.title).toBeTruthy();
      expect(recipe.statisticalTask).toBeTruthy();
      expect(recipe.moduleId).toMatch(/^module_[1-6]$/);

      // Excel specifications
      expect(recipe.excel).toBeDefined();
      expect(recipe.excel.functionName).toBeTruthy();
      expect(recipe.excel.genericSyntax).toBeTruthy();
      expect(recipe.excel.argumentDescriptions.length).toBeGreaterThan(0);
      for (const arg of recipe.excel.argumentDescriptions) {
        expect(arg.name).toBeTruthy();
        expect(arg.description).toBeTruthy();
      }

      // R specifications
      expect(recipe.r).toBeDefined();
      expect(recipe.r.functionName).toBeTruthy();
      expect(recipe.r.genericSyntax).toBeTruthy();
      expect(recipe.r.argumentDescriptions.length).toBeGreaterThan(0);
      for (const arg of recipe.r.argumentDescriptions) {
        expect(arg.name).toBeTruthy();
        expect(arg.description).toBeTruthy();
      }
    }
  });

  it('provides recipes for all 6 curriculum modules', () => {
    for (let i = 1; i <= 6; i++) {
      const moduleId = `module_${i}`;
      const moduleRecipes = getRecipesByModule(moduleId);
      expect(moduleRecipes.length).toBeGreaterThanOrEqual(2);
    }
  });

  it('retrieves recipes by ID, module, and family', () => {
    const binomExact = getRecipeById('m2_binom_exact_pmf');
    expect(binomExact).toBeDefined();
    expect(binomExact?.excel.functionName).toContain('BINOM.DIST');
    expect(binomExact?.r.functionName).toBe('dbinom');

    const invalidRecipe = getRecipeById('non_existent_recipe_id');
    expect(invalidRecipe).toBeUndefined();

    const descRecipes = getRecipesByFamily('descriptive_stats');
    expect(descRecipes.length).toBeGreaterThanOrEqual(4);

    const mod6Recipes = getRecipesByModule('module_6');
    expect(mod6Recipes.length).toBeGreaterThanOrEqual(5);
  });

  it('maps every single one of the 137 approved templates to a valid recipe', () => {
    // Resolve curriculum/approved path relative to web root
    const approvedDir = path.resolve(__dirname, '../../../../curriculum/approved');
    expect(fs.existsSync(approvedDir)).toBe(true);

    const files = fs.readdirSync(approvedDir).filter((f) => f.endsWith('.json'));
    expect(files.length).toBe(137);

    for (const file of files) {
      const fullPath = path.join(approvedDir, file);
      const templateData = JSON.parse(fs.readFileSync(fullPath, 'utf-8'));
      const templateId = templateData.id;

      // Ensure template is mapped in TEMPLATE_TO_RECIPE_MAP
      const recipeId = TEMPLATE_TO_RECIPE_MAP[templateId];
      expect(recipeId, `Template ${templateId} should be in TEMPLATE_TO_RECIPE_MAP`).toBeDefined();

      const recipe = getRecipeById(recipeId);
      expect(recipe, `Recipe ${recipeId} mapped to template ${templateId} should exist`).toBeDefined();

      // Test resolver function directly
      const resolved = getRecipeForTemplate({
        id: templateId,
        family_id: templateData.family_id,
        module_id: templateData.module_id,
        title: templateData.title,
      });

      expect(resolved, `Resolver should return a recipe for template ${templateId}`).toBeDefined();
      expect(resolved?.id).toBe(recipeId);
    }
  });

  it('handles query fallbacks gracefully', () => {
    // 1. Fallback by family_id and module_id
    const fallbackFamilyModule = getRecipeForTemplate({
      family_id: 'binomial_pmf',
      module_id: 'module_2',
    });
    expect(fallbackFamilyModule).toBeDefined();
    expect(fallbackFamilyModule?.id).toBe('m2_binom_exact_pmf');

    // 2. Fallback by family_id alone
    const fallbackFamily = getRecipeForTemplate({
      family_id: 'linear_regression',
    });
    expect(fallbackFamily).toBeDefined();
    expect(fallbackFamily?.id).toBe('m6_reg_bivariate_fit');

    // 3. Fallback by title keyword
    const fallbackKeyword = getRecipeForTemplate({
      title: 'Interpreting the correlation coefficient and slope direction',
    });
    expect(fallbackKeyword).toBeDefined();

    // 4. Fallback by module_id alone
    const fallbackModule = getRecipeForTemplate({
      module_id: 'module_4',
    });
    expect(fallbackModule).toBeDefined();
    expect(fallbackModule?.moduleId).toBe('module_4');

    // 5. Completely empty query
    const emptyResult = getRecipeForTemplate({});
    expect(emptyResult).toBeUndefined();
  });
});
