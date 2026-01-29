import { Request, Response } from 'express';
import { RotationService } from '../services/RotationService';
import { RotationStrategyFactory } from '../strategies/RotationStrategy';

/**
 * Controller layer for rotation endpoints
 * Handles HTTP requests and responses
 */
export class RotationController {
    constructor(private rotationService: RotationService) { }

    /**
     * Handle GET /api/next request
     * Query params: count (default: 1)
     */
    getNext = (req: Request, res: Response) => {
        try {
            const count = req.query.count ? parseInt(req.query.count as string) : 1;

            // Validate count
            if (count < 1) {
                return res.status(400).json({
                    success: false,
                    message: 'Count must be at least 1',
                });
            }

            const result = this.rotationService.getNext(RotationStrategyFactory.getRoundRobinStrategy(), count);

            if (!result.success) {
                return res.status(400).json({
                    success: false,
                    message: result.message,
                });
            }

            return res.json({
                success: true,
                data: result.data
            });
        } catch (error) {
            return res.status(500).json({
                success: false,
                message: 'Internal server error',
            });
        }
    };
}
