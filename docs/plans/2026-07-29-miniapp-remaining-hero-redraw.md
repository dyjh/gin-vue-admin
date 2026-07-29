# Remaining Mini Program Hero Redraw Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Redraw and integrate every remaining Mini Program hero image under the approved mascot and watercolor kitchen style.

**Architecture:** Generate one independent bitmap per distinct page meaning, with only the shopping list and its read-only share page intentionally sharing an image. Cache current and generated previews in `D:\www\aiCache`, convert approved batch outputs to versioned 1500×804 JPEG files under `miniApp/assets/images`, and update only the exact hero references while preserving all unrelated dirty-worktree changes.

**Tech Stack:** WeChat Mini Program WXML/JavaScript, built-in image generation, PowerShell `System.Drawing`, existing `miniApp/tools/validate-miniapp.cjs`.

---

### Task 1: Inventory and protect existing work

**Files:**
- Inspect: `miniApp/pages/ideas/`
- Inspect: `miniApp/pages/recipes/`
- Inspect: `miniApp/pages/profile/`
- Inspect: `miniApp/pages/checkin/`
- Inspect: `miniApp/pages/meal/`
- Inspect: `miniApp/pages/shopping/`
- Inspect: `miniApp/pages/points/`
- Inspect: `miniApp/pages/notifications/`

**Steps:**
1. Record every current hero reference and dynamic profile hero assignment.
2. Run targeted `git status --short` and `git diff` for all consuming files.
3. Preserve every pre-existing modification; reject any replacement whose exact old reference does not occur once per intended target.

### Task 2: Cache current remote heroes

**Files:**
- Create only comparison copies under: `D:\www\aiCache`

**Steps:**
1. Download each unique current remote hero from `https://cache.ljdyjh.cn/assets/images/...`.
2. Keep shared originals as one cache file where the old resource is shared.
3. If a URL fails, record and report the exact URL without guessing.

### Task 3: Generate 15 page-specific hero assets

**Files:**
- Create previews under: `D:\www\aiCache`

**Scene map:**
1. `ideas/what-to-eat`: cat naturally choosing ingredients from an open refrigerator.
2. `ideas/what-to-eat-detail`: cat revealing the suggested finished dish under a small serving cover.
3. `recipes/index`: cat arranging a small set of bound recipe notebooks on a kitchen shelf.
4. `recipes/detail`: cat planning a grouped meal with recipe pages and several dish markers, distinct from dish-detail reading.
5. `profile/index`: cat tidying its personal kitchen corner and apron hooks, leaving identity overlay space.
6. `checkin/index`: cat naturally photographing a finished meal with a small camera.
7. `meal/create`: cat setting a shared dinner table with several place settings.
8. `meal/invite`: cat presenting a simple invitation card with a non-readable code pattern.
9. `meal/vote`: cat choosing among several covered home dishes with one green selection token.
10. `meal/stats`: cat checking final serving portions using bowls and a kitchen scale.
11. `meal/history`: cat reviewing a kitchen calendar and a small album of past meals.
12. `shopping/list` and `shopping/share`: cat packing vegetables into a grocery basket while checking one list.
13. `meal/prep-guide`: cat arranging three cooking stations in a natural preparation sequence.
14. `points/index`: cat placing muted green kitchen tokens into a glass jar.
15. `notifications/index`: cat opening and reading a cream envelope at the kitchen counter.

**Steps:**
1. Before every generation call, re-emit `miniApp/assets/images/home-approved-header-v3.jpg` as the sole image reference.
2. Generate one image per scene; never use one call as a substitute for distinct assets.
3. Enforce the shared constraints: one canonical cat, matched orange paws, forest-green paw-print apron, left title safe area, upper-right capsule safe area, no embedded text, no repeated cloth/tomato/rosemary cluster, no awkward object interaction.
4. Inspect every output and regenerate any asset with identity drift, unsafe crop, malformed paws, floating objects, or scene mismatch.

### Task 4: Convert and integrate assets

**Files:**
- Create: `miniApp/assets/images/*-v1.jpg`
- Modify: the 16 consuming WXML/JavaScript files listed in Task 1.

**Steps:**
1. Convert every selected source to exactly 1500×804 JPEG at quality 90.
2. Verify each file is below 1 MB.
3. Replace only the intended image string; keep `shopping/list` and `shopping/share` on the same new image.
4. Give `meal/stats` and `meal/prep-guide` separate resources.
5. Give `recommend/dish-detail` and `ideas/what-to-eat-detail` separate resources.
6. Do not add local versioned images to `miniApp/config/remote-assets.js`.

### Task 5: Document and verify

**Files:**
- Modify: `aiDoc/memory/business/active/miniapp-hero-redraw.md`

**Steps:**
1. Record the batch scope, generated filenames, shared-image decisions, and validation result.
2. Run exact reference counts for every new image and confirm old page references are gone.
3. Run `git diff --check` only on touched consuming files; report unrelated existing warnings without changing them.
4. Run `miniApp/tools/validate-miniapp.cjs`.
5. Deliver one page-to-resource checklist for a single WeChat DevTools visual pass.
