import { Request, Response } from 'express';
import { RotationController } from '../controllers/RotationController';
import { RotationService } from '../services/RotationService';
import { Member } from '../utils/types';
import * as dataModule from '../data/members';

describe('RotationController', () => {
    let controller: RotationController;
    let service: RotationService;
    let mockRequest: Partial<Request>;
    let mockResponse: Partial<Response>;

    const fullActimembers: Member[] = [
        { id: 1, name: 'Alice', isActive: true },
        { id: 2, name: 'Bob', isActive: true },
        { id: 3, name: 'Charlie', isActive: true },
        { id: 4, name: 'Diana', isActive: true },
    ];

    const inactiveMembers: Member[] = [
        { id: 1, name: 'Alice', isActive: false },
        { id: 2, name: 'Bob', isActive: false },
    ];

    const partialActiveMembers: Member[] = [
        { id: 1, name: 'Alice', isActive: true },
        { id: 2, name: 'Bob', isActive: false },
        { id: 3, name: 'Charlie', isActive: true },
    ];

    beforeEach(() => {
        jest.clearAllMocks();

        service = new RotationService();
        controller = new RotationController(service);



        mockResponse = {
            json: jest.fn().mockReturnThis(),
            status: jest.fn().mockReturnThis(),
        };

        mockRequest = {
            query: {},
            params: {},
        };

        jest.spyOn(dataModule, 'getMemberlist').mockReturnValue(fullActimembers);
    });

    it('should return error when count is invalid', () => {
        mockRequest.query = { count: '0' };
        controller.getNext(mockRequest as Request, mockResponse as Response);
        expect(mockResponse.status).toHaveBeenCalledWith(400);
        expect(mockResponse.json).toHaveBeenCalledWith(
            expect.objectContaining({
                success: false,
                message: 'Count must be at least 1',
            })
        );
    });

    it('should use default count of 1 when not provided', (done) => {
        mockRequest.query = {};

        controller.getNext(mockRequest as Request, mockResponse as Response);

        setTimeout(() => {
            expect(mockResponse.json).toHaveBeenCalledWith(
                expect.objectContaining({
                    success: true,
                    data: expect.arrayContaining([
                        expect.objectContaining({
                            id: 1,
                            name: 'Alice'
                        }),
                    ]),
                })
            );
            done();
        }, 0);
    });

    it('should return multiple members when count > 1', (done) => {
        mockRequest.query = { count: '3' };

        controller.getNext(mockRequest as Request, mockResponse as Response);

        setTimeout(() => {
            expect(mockResponse.json).toHaveBeenCalledWith(
                expect.objectContaining({
                    success: true,
                    data: expect.arrayContaining([
                        expect.objectContaining({ name: 'Alice' }),
                        expect.objectContaining({ name: 'Bob' }),
                        expect.objectContaining({ name: 'Charlie' }),
                    ]),
                })
            );
            done();
        }, 0);
    });

    it('should return error if count > length of memberList', (done) => {
        mockRequest.query = { count: '5' };

        controller.getNext(mockRequest as Request, mockResponse as Response);

        setTimeout(() => {
            expect(mockResponse.status).toHaveBeenCalledWith(400);
            expect(mockResponse.json).toHaveBeenCalledWith(
                expect.objectContaining({
                    success: false,
                    message: 'Count cannot exceed total number of members (4)',
                })
            );
            done();
        }, 0);
    });


    it('should no immediate repetition', (done) => {
        mockRequest.query = { count: '2' };
        controller.getNext(mockRequest as Request, mockResponse as Response);
        setTimeout(() => {
            expect(mockResponse.json).toHaveBeenCalledWith(
                expect.objectContaining({
                    success: true,
                    data: expect.arrayContaining([
                        expect.objectContaining({ name: 'Alice' }),
                        expect.objectContaining({ name: 'Bob' }),
                    ]),
                })
            );
            done();
        }, 0);
        // second call
        controller.getNext(mockRequest as Request, mockResponse as Response);

        expect(mockResponse.json).toHaveBeenCalledWith(
            expect.objectContaining({
                success: true,
                data: expect.arrayContaining([
                    expect.objectContaining({ name: 'Charlie' }),
                    expect.objectContaining({ name: 'Diana' }),
                ]),
            })
        );

        // final call 
        controller.getNext(mockRequest as Request, mockResponse as Response);

        expect(mockResponse.json).toHaveBeenCalledWith(
            expect.objectContaining({
                success: true,
                data: expect.arrayContaining([
                    expect.objectContaining({ name: 'Alice' }),
                    expect.objectContaining({ name: 'Bob' }),
                ]),
            })
        );

    });


    it('should return error when not enough active members to fulfill request', (done) => {

        jest.spyOn(dataModule, 'getMemberlist').mockReturnValue(partialActiveMembers);

        mockRequest.query = { count: '3' };

        controller.getNext(mockRequest as Request, mockResponse as Response);

        expect(mockResponse.status).toHaveBeenCalledWith(400);
        expect(mockResponse.json).toHaveBeenCalledWith(
            expect.objectContaining({
                success: false,
            })
        );
        done();
    });


    it('should return error when no active members available', () => {

        jest.spyOn(dataModule, 'getMemberlist').mockReturnValue(inactiveMembers);
        mockRequest.query = { count: '1' };
        controller.getNext(mockRequest as Request, mockResponse as Response);

        expect(mockResponse.status).toHaveBeenCalledWith(400);
        expect(mockResponse.json).toHaveBeenCalledWith(
            expect.objectContaining({
                success: false,
                message: 'No active members available',
            })
        );

    });

});
