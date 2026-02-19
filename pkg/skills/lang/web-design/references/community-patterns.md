## Typography
- Use a maximum of 2-3 font families
- Set body text at 16-18px minimum for readability
- Use a modular type scale: 1.25 (minor third) or 1.333 (perfect fourth)
- Line height: 1.5 for body text, 1.2 for headings
- Max line length: 60-75 characters for comfortable reading
- Use font-display: swap for web fonts

## Color
- Define a consistent color palette: primary, secondary, neutral, semantic
- Use sufficient contrast ratios (WCAG AA: 4.5:1 for text, 3:1 for large text)
- Don't rely on color alone to convey information
- Use opacity and tints/shades for hierarchy
- Test with color blindness simulators

## Spacing
- Use a consistent spacing scale: 4px base (4, 8, 12, 16, 24, 32, 48, 64)
- Apply generous whitespace — don't crowd elements
- Use consistent padding within components
- Use consistent gaps between components

## Layout
- Use CSS Grid for 2D layouts, Flexbox for 1D
- Design mobile-first, enhance for larger screens
- Use max-width containers for readable content (1200-1400px)
- Maintain visual hierarchy with size, weight, color, and position

## Accessibility
- All images need descriptive alt text
- Use semantic HTML: nav, main, article, section, aside
- Ensure keyboard navigability for all interactive elements
- Use ARIA labels only when semantic HTML isn't sufficient
- Test with screen readers (VoiceOver, NVDA)
- Focus indicators must be visible

## Performance
- Optimize images: WebP/AVIF, responsive srcset, lazy loading
- Use system fonts or minimal web font loading
- Minimize layout shifts (CLS): set explicit dimensions on media
- Defer non-critical CSS and JavaScript
