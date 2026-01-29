import express from 'express';
import { RotationService } from '../services/RotationService';
import { RotationController } from '../controllers/RotationController';

/**
 * Create and configure the API routes
 * Focuses on the core /api/next endpoint for team rotation
 * Uses Singleton RotationService to ensure single state across requests
 */
export function createApiRoutes(): express.Router {
    const router = express.Router();

    // Get or create singleton instance of RotationService
    const rotationService = RotationService.getInstance();
    const rotationController = new RotationController(rotationService);

    /**
     * Core Rotation Endpoint
     * GET /api/next?count=1
     * Returns the next member(s) for rotation
     */
    router.get('/next', rotationController.getNext);

    return router;
}
