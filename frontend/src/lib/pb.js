import PocketBase from 'pocketbase';

// Connect to PB on 8090 during local dev, else use relative for production
const url = import.meta.env.DEV ? 'http://localhost:8090' : '/';
export const pb = new PocketBase(url);
