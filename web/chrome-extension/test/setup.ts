import { vi } from 'vitest';

// Mock microformat-node parser for testing
const mockMfParser = vi.fn();

// Set up global mfParser mock
(global as any).mfParser = mockMfParser;

export { mockMfParser };