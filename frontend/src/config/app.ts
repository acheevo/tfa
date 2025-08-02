export const config = {
  app: {
    name: 'Fullstack Template',
    version: '1.0.0',
  },
  api: {
    baseUrl: process.env.VITE_API_BASE_URL || '/api',
  },
  apiUrl: process.env.VITE_API_BASE_URL || '/api',
};