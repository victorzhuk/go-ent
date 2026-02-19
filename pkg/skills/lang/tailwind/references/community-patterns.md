## Core Principles
- Utility-first: compose designs from small utility classes
- Responsive: mobile-first with breakpoint prefixes (`sm:`, `md:`, `lg:`, `xl:`)
- State variants: `hover:`, `focus:`, `active:`, `disabled:`, `dark:`
- Custom values: arbitrary values with `[]` syntax: `w-[137px]`

## Common Patterns
```html
<!-- Card component -->
<div class="rounded-lg border border-gray-200 bg-white p-6 shadow-sm">
  <h3 class="text-lg font-semibold text-gray-900">Title</h3>
  <p class="mt-2 text-sm text-gray-600">Description</p>
</div>

<!-- Responsive grid -->
<div class="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
  <!-- items -->
</div>

<!-- Flexbox centering -->
<div class="flex items-center justify-between gap-4">
  <!-- content -->
</div>
```

## Theme Configuration (v4)
```css
@theme {
  --color-primary: #3b82f6;
  --color-secondary: #6b7280;
  --font-sans: 'Inter', sans-serif;
  --radius-default: 0.5rem;
}
```

## Component Extraction
- Use `@apply` sparingly — prefer component abstractions in your framework
- Extract repeated patterns into React/Vue/Svelte components
- Use CSS variables for theme tokens
- Group related utilities logically: layout → spacing → typography → colors

## Dark Mode
```html
<div class="bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100">
  <!-- Automatically switches with system preference or class toggle -->
</div>
```

## Best Practices
- Don't fight Tailwind with custom CSS — embrace utilities
- Use the `cn()` utility (clsx + tailwind-merge) for conditional classes
- Install Tailwind CSS IntelliSense for IDE support
- Use Prettier plugin for consistent class ordering
- Purge unused styles automatically (built into v4)
