import type { components } from '../../api/v1';

type PostgresTable = components['schemas']['storage_postgres_table'];

const STORAGE_KEY_PREFIX = 'meshpump:storage-tables:';

/**
 * Get the sessionStorage key for a storage's tables
 */
export function getStorageKey(uuid: string | undefined): string {
  return `${STORAGE_KEY_PREFIX}${uuid || 'new'}`;
}

/**
 * Get tables from sessionStorage
 */
export function getTables(storageKey: string): PostgresTable[] {
  try {
    const data = sessionStorage.getItem(storageKey);
    if (!data) return [];
    return JSON.parse(data) as PostgresTable[];
  } catch {
    return [];
  }
}

/**
 * Save tables to sessionStorage
 */
export function setTables(storageKey: string, tables: PostgresTable[]): void {
  sessionStorage.setItem(storageKey, JSON.stringify(tables));
}

/**
 * Clear tables from sessionStorage
 */
export function clearTables(storageKey: string): void {
  sessionStorage.removeItem(storageKey);
}

/**
 * Check if tables exist in sessionStorage
 */
export function hasTables(storageKey: string): boolean {
  return sessionStorage.getItem(storageKey) !== null;
}
