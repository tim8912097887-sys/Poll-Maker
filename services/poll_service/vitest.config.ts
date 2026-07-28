import { configDefaults, defineConfig } from 'vitest/config';
import tsconfigPaths from 'vite-tsconfig-paths';

export default defineConfig({
  plugins: [tsconfigPaths()],
  test: {
    // Enable globals like 'describe', 'it', 'expect' (optional)
    globals: true,
    environment: 'node',
    include: ['test/**/*.{test,spec}.ts'],
    exclude: [
      ...configDefaults.exclude,
      'test/setup.ts',
      'dist',
      'coverage',
      'test/utils/*.ts',
      'test/global-setup.ts',
    ], // Exclude setup file from test files
    // Setup file for environment variables or global mocks
    setupFiles: ['./test/setup.ts'],
    globalSetup: './test/global-setup.ts',
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      exclude: ['node_modules/', 'test/'],
    },
  },
});
