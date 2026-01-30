import express from 'express';
import { createApiRoutes } from './routes/api';

const app = express();
const PORT = process.env.PORT || 1902;

// Middleware
app.use(express.json());

// API Routes with Service and Controller layers
app.use('/api', createApiRoutes());

// Health check endpoint
app.get('/api/health', (_req, res) => {
    res.json({ status: 'healthy' });
});

// Start server
app.listen(PORT, () => {
    console.log(`🚀 Smart Team Rotator API running on http://localhost:${PORT}`);
    console.log('');
    console.log('⭐ Core Endpoint:');
    console.log('   GET  /api/next?count=1');
    console.log('');
    console.log('🏥 Health Check:');
    console.log('   GET  /api/health');
});

export default app;
