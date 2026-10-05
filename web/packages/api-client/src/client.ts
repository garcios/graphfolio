import { createClient as createGenqlClient, type Client } from './generated';

export interface ClientConfig {
  url?: string;
  headers?: Record<string, string>;
}

export const DEFAULT_BFF_URL = 'http://localhost:8080/query';

export function createGraphfolioClient(config: ClientConfig = {}): Client {
  return createGenqlClient({
    url: config.url || DEFAULT_BFF_URL,
    headers: config.headers,
  });
}

// Default singleton client instance for applications
export const client: Client = createGraphfolioClient();
