export const config = {
  app: {
    name: 'Fullstack Template',
    version: '1.0.0',
  },
  api: {
    baseUrl: import.meta.env.VITE_API_BASE_URL || '/api',
  },
  apiUrl: import.meta.env.VITE_API_BASE_URL || '/api',
};