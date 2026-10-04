import { createClient } from '../generated';

export const client = createClient({
  url: 'http://localhost:8080/query',
});
