// Augment Vitest's `expect` with @testing-library/jest-dom matchers
// (`toBeInTheDocument`, `toBeDisabled`, `toHaveAttribute`, etc.) globally so
// individual test files don't need a per-file `import type {} from
// '@testing-library/jest-dom'` to satisfy the type checker.
import '@testing-library/jest-dom/vitest';
