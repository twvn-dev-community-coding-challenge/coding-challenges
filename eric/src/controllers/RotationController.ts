import { Request, Response } from 'express';
import { RotationService } from '../services/RotationService';
import { RotationStrategyFactory } from '../strategies/RotationStrategy';

/**
 * Controller Layer: Handles HTTP requests for rotation endpoints
 * 
 * Responsibilities:
 * - Parse and validate HTTP request parameters
 * - Delegate business logic to RotationService
 * - Format and return HTTP responses
 * - Handle errors and return appropriate status codes
 */
export class RotationController {
    /**
     * Creates a new RotationController instance
     * @param rotationService - The service instance for handling rotation logic
     */
    constructor(private rotationService: RotationService) { }

    /**
     * Handles GET /api/next request
     * Returns the next member(s) in the rotation sequence
     * 
     * Query Parameters:
     * - count: Number of members to return (optional, default: 1)
     * 
     * Response Codes:
     * - 200: Success with selected members
     * - 400: Validation error (invalid count, no active members, etc.)
     * - 500: Unexpected server error
     * 
     * @param req - Express request object
     * @param res - Express response object
     */
    getNext = (req: Request, res: Response) => {
        try {
            // Parse count from query string, default to 1 if not provided
            const countParam = req.query.count as string;
            const count = countParam ? parseInt(countParam, 10) : 1;
            
            // Check if count is a valid number
            if (isNaN(count)) {
                return res.status(400).json({
                    success: false,
                    message: 'Count must be a valid number',
                });
            }

            // Validate count at controller level for early rejection
            if (count < 1) {
                return res.status(400).json({
                    success: false,
                    message: 'Count must be at least 1',
                });
            }

            // Delegate to service layer with the Round-Robin strategy
            const result = this.rotationService.getNext(RotationStrategyFactory.getRoundRobinStrategy(), count);

            // Handle business logic failures (validation, no active members, etc.)
            if (!result.success) {
                return res.status(400).json({
                    success: false,
                    message: result.message,
                });
            }

            // Return successful response with selected members
            return res.json({
                success: true,
                data: result.data
            });
        } catch (error) {
            // Handle unexpected errors (strategy errors, data issues, etc.)
            return res.status(500).json({
                success: false,
                message: error instanceof Error ? error.message : 'Internal server error',
            });
        }
    };
}
