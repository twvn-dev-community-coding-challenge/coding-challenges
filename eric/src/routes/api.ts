import express from 'express';
import { RotationService } from '../services/RotationService';
import { RotationController } from '../controllers/RotationController';

/**
 * Creates and configures the API router
 * 
 * This function sets up the core rotation endpoint using:
 * - Singleton RotationService for consistent state across requests
 * - RotationController for HTTP request handling
 * 
 * @returns Configured Express router with rotation endpoints
 */
export function createApiRoutes(): express.Router {
    const router = express.Router();

    // Get singleton instance to ensure consistent rotation state across all requests
    const rotationService = RotationService.getInstance();
    const rotationController = new RotationController(rotationService);

    /**
     * GET /api/next
     * 
     * Core Rotation Endpoint - Returns the next member(s) in rotation sequence
     * 
     * Query Parameters:
     * - count: Number of members to return (default: 1)
     * 
     * Example: GET /api/next?count=2
     */
    router.get('/next', rotationController.getNext);

    return router;
}
