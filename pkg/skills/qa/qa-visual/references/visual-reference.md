# Visual & Accessibility Testing Reference

## Install

```bash
pip install shot-scraper
shot-scraper install

npm install -g @axe-core/cli
npm install -g lighthouse
npm install -g @playwright/playwright-cli

brew install imagemagick
# or: apt install imagemagick
```

## shot-scraper — Screenshot CLI

```bash
shot-scraper https://example.com
shot-scraper https://example.com --full-page
shot-scraper https://example.com --selector "main .hero-section"
shot-scraper https://example.com \
  --javascript "document.querySelector('.cookie-banner')?.remove()"
shot-scraper https://example.com --retina
shot-scraper https://example.com --wait-for ".dashboard-loaded"

cat > shots.yaml << 'EOF'
- url: https://example.com
  output: baselines/home.png
  full_page: true

- url: https://example.com/login
  output: baselines/login.png
  selector: form.login-form

- url: https://example.com/products
  output: baselines/products.png
  wait_for: ".product-grid"
  javascript: "window.scrollTo(0, 0)"
EOF

shot-scraper multi shots.yaml
shot-scraper multi shots.yaml --compare baselines/
```

## Pixel Diff with ImageMagick

```bash
compare -metric AE baseline.png current.png diff.png
echo "Different pixels: $?"

compare -highlight-color red -fuzz 5% baseline.png current.png diff.png
convert baseline.png current.png +append side-by-side.png
compare -metric PSNR baseline.png current.png /dev/null 2>&1
compare -fuzz 2% baseline.png current.png diff.png

diff_visual() {
  local baseline=$1 current=$2 output=$3
  local diff_pixels=$(compare -metric AE "$baseline" "$current" "$output" 2>&1)
  if [ "$diff_pixels" -gt 0 ]; then
    echo "FAIL: $diff_pixels pixels differ → $output"
    return 1
  fi
  echo "PASS: $baseline"
}
```

## pixelmatch (Node)

```javascript
// visual-diff.mjs
import { readFileSync, writeFileSync } from 'fs';
import { PNG } from 'pngjs';
import pixelmatch from 'pixelmatch';

const baseline = PNG.sync.read(readFileSync(process.argv[2]));
const current  = PNG.sync.read(readFileSync(process.argv[3]));
const { width, height } = baseline;
const diff = new PNG({ width, height });

const numDiffPixels = pixelmatch(
  baseline.data, current.data, diff.data,
  width, height,
  { threshold: 0.1, includeAA: false }
);

writeFileSync(process.argv[4] || 'diff.png', PNG.sync.write(diff));
const percent = (numDiffPixels / (width * height) * 100).toFixed(2);
console.log(`Diff: ${numDiffPixels} pixels (${percent}%)`);
process.exit(numDiffPixels > 0 ? 1 : 0);
```

```bash
node visual-diff.mjs baselines/home.png current/home.png diffs/home-diff.png
```

## playwright-cli Visual Regression

```bash
playwright-cli open https://app.example.com
playwright-cli screenshot --output baselines/home.png --full-page

playwright-cli screenshot --output current/home.png --full-page
compare -metric AE baselines/home.png current/home.png diffs/home.png
```

### Playwright built-in visual comparison

```typescript
// tests/visual.spec.ts
import { test, expect } from '@playwright/test';

test('home page visual regression', async ({ page }) => {
  await page.goto('https://app.example.com');
  await expect(page).toHaveScreenshot('home.png', {
    maxDiffPixels: 100,
    threshold: 0.2,
  });
});

test('product card component', async ({ page }) => {
  await page.goto('https://app.example.com/products');
  const card = page.locator('.product-card').first();
  await expect(card).toHaveScreenshot('product-card.png');
});
```

```bash
npx playwright test --update-snapshots
npx playwright test tests/visual.spec.ts
```

## Accessibility Testing — axe CLI

```bash
axe https://example.com
axe https://example.com --tags wcag2a,wcag2aa,wcag21aa
axe https://example.com --reporter json > results/a11y.json
axe https://example.com --browser chrome
axe https://example.com --include "main, nav" --exclude ".ads"
axe https://example.com --exit

for url in / /login /products /checkout; do
  echo "Auditing: $url"
  axe "https://example.com$url" --reporter json >> results/a11y-all.json
done
```

## Lighthouse CLI — Performance + A11y

```bash
lighthouse https://example.com --output=html --output-path=./reports/lighthouse.html

lighthouse https://example.com --only-categories=performance
lighthouse https://example.com --only-categories=accessibility \
  --output=json --output-path=./reports/a11y-lighthouse.json

lighthouse https://example.com \
  --chrome-flags="--headless --no-sandbox" \
  --output=json \
  --output-path=/dev/stdout | jaq '.categories.performance.score'

PERF_SCORE=$(lighthouse https://example.com \
  --only-categories=performance \
  --output=json \
  --quiet \
  --chrome-flags="--headless" | jaq '.categories.performance.score')

if (( $(echo "$PERF_SCORE < 0.9" | bc -l) )); then
  echo "FAIL: Performance score $PERF_SCORE < 0.90"
  exit 1
fi
```

## Agent Patterns

### Baseline capture flow

```bash
mkdir -p baselines/{desktop,mobile}

for page in / /login /products /checkout /profile; do
  shot-scraper "https://app.example.com${page}" \
    --output "baselines/desktop$(echo $page | tr / -).png" \
    --full-page \
    --width 1280 --height 800
done

for page in / /login /products; do
  shot-scraper "https://app.example.com${page}" \
    --output "baselines/mobile$(echo $page | tr / -).png" \
    --full-page \
    --width 390 --height 844
done
```

### Regression check flow

```bash
run_visual_regression() {
  local failures=0
  mkdir -p current/ diffs/

  for baseline in baselines/*.png; do
    name=$(basename "$baseline")
    shot-scraper "$url" --output "current/$name" --full-page

    diff_pixels=$(compare -metric AE "$baseline" "current/$name" "diffs/$name" 2>&1)
    if [ "$diff_pixels" -gt 50 ]; then
      echo "FAIL [$diff_pixels px diff]: $name"
      failures=$((failures + 1))
    else
      echo "PASS: $name"
    fi
  done

  echo "Visual regression: $failures failures"
  return $failures
}
```

### A11y audit flow

```bash
axe https://app.example.com \
  --tags wcag2aa \
  --reporter json \
  --exit 2>&1 | tee results/a11y.json

VIOLATIONS=$(jaq '.violations | length' < results/a11y.json)
CRITICAL=$(jaq '[.violations[] | select(.impact == "critical")] | length' < results/a11y.json)

echo "Total violations: $VIOLATIONS"
echo "Critical: $CRITICAL"

if [ "$CRITICAL" -gt 0 ]; then
  echo "FAIL: Critical accessibility violations found"
  jaq '.violations[] | select(.impact == "critical") | {id, description, nodes: (.nodes | length)}' \
    < results/a11y.json
  exit 1
fi
```
