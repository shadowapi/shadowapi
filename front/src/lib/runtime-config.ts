// Runtime config injected by the SSR server from environment variables.
// Takes precedence over build-time VITE_* values baked in by Vite.
const rc: Record<string, string> =
  typeof window !== 'undefined' ? (window as Record<string, unknown>).__RUNTIME_CONFIG__ as Record<string, string> ?? {} : {};

export function getRuntimeConfig(key: string, buildTimeValue: string): string {
  return rc[key] || buildTimeValue;
}
