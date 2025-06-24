import express from 'express';
import { CORS_HEADERS } from './common.js';
import createSpcdeInfoService from './services/spcdeInfoService.js';
import createSecuritiesProductInfoService from './services/getSecuritiesProductInfoService.js';

const AUTH_API_KEY = process.env.AUTH_API_KEY;

const app = express();

app.use((req, res, next) => {
  if (req.method === 'OPTIONS') {
    return res.set(CORS_HEADERS).status(204).send();
  }

  const clientApiKey = req.header('x-api-key');
  if (!clientApiKey || clientApiKey !== AUTH_API_KEY) {
    return res.status(401).json({
      error: 'Unauthorized',
      message: 'Invalid or missing x-api-key header'
    });
  }

  next();
});

app.use(createSpcdeInfoService());
app.use(createSecuritiesProductInfoService());

app.use((req, res) => {
  res.status(404).send('Not Found');
});

const PORT = process.env.PORT || 3000;
app.listen(PORT, () => {
  console.log(`Proxy server running on port ${PORT}`);
});
