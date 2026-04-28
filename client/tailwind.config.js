export default {
  content: ['./index.html', './src/**/*.{vue,js}'],
  theme: {
    extend: {
      colors: {
        base: 'rgb(var(--color-bg-base-rgb) / <alpha-value>)',
        surface: 'rgb(var(--color-bg-surface-rgb) / <alpha-value>)',
        elevated: 'rgb(var(--color-bg-elevated-rgb) / <alpha-value>)',
        border: 'rgb(var(--color-border-rgb) / <alpha-value>)',
        accent: 'rgb(var(--color-accent-rgb) / <alpha-value>)',
        warm: 'rgb(var(--color-accent-warm-rgb) / <alpha-value>)',
        green: 'rgb(var(--color-accent-green-rgb) / <alpha-value>)',
        amber: 'rgb(var(--color-accent-amber-rgb) / <alpha-value>)',
        primary: 'rgb(var(--color-text-primary-rgb) / <alpha-value>)',
        secondary: 'rgb(var(--color-text-secondary-rgb) / <alpha-value>)',
        muted: 'rgb(var(--color-text-muted-rgb) / <alpha-value>)'
      },
      fontFamily: {
        display: ['Space Grotesk', 'sans-serif'],
        sans: ['DM Sans', 'sans-serif'],
        mono: ['JetBrains Mono', 'monospace']
      },
      boxShadow: {
        glow: '0 0 0 1px rgb(108 99 255 / 0.35), 0 18px 60px rgb(108 99 255 / 0.16)'
      }
    }
  },
  plugins: []
};
