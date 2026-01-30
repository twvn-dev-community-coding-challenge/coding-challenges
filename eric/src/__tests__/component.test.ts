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

    it('should return error when count is not valid number', () => {
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

     it('should return error when count is not a number', () => {
        mockRequest.query = { count: 'aaaa' };
        controller.getNext(mockRequest as Request, mockResponse as Response);
        expect(mockResponse.status).toHaveBeenCalledWith(400);
        expect(mockResponse.json).toHaveBeenCalledWith(
            expect.objectContaining({
                success: false,
                message: 'Count must be a valid number',
            })
        );
    });

    it('should use default count of 1 when not provided', () => {
        mockRequest.query = {};

        controller.getNext(mockRequest as Request, mockResponse as Response);

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
    });

    it('should return multiple members when count > 1', () => {
        mockRequest.query = { count: '3' };

        controller.getNext(mockRequest as Request, mockResponse as Response);

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
    });

    it('should return error if count > length of memberList', () => {
        mockRequest.query = { count: '5' };

        controller.getNext(mockRequest as Request, mockResponse as Response);
        expect(mockResponse.status).toHaveBeenCalledWith(400);
        expect(mockResponse.json).toHaveBeenCalledWith(
            expect.objectContaining({
                success: false,
                message: 'Count cannot exceed total number of members (4)',
            })
        );
    });


    it('should no immediate repetition', () => {
        mockRequest.query = { count: '2' };
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

        mockResponse.json = jest.fn().mockReturnThis();

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

        mockResponse.json = jest.fn().mockReturnThis();

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


    it('should return error when not enough active members to fulfill request', () => {

        jest.spyOn(dataModule, 'getMemberlist').mockReturnValue(partialActiveMembers);

        mockRequest.query = { count: '3' };

        controller.getNext(mockRequest as Request, mockResponse as Response);

        expect(mockResponse.status).toHaveBeenCalledWith(400);
        expect(mockResponse.json).toHaveBeenCalledWith(
            expect.objectContaining({
                success: false,
            })
        );
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

    it('should return correcct in case member index and id are out of sync', () => {
        service = new RotationService();
        service['lastSelectedMemberIndex'] = 1; // Points to Bob
        service['lastSelectedMemberId'] = 3;    // But last selected id is Charlie
        controller = new RotationController(service);

        mockRequest.query = { count: '2' };

        controller.getNext(mockRequest as Request, mockResponse as Response);

        expect(mockResponse.json).toHaveBeenCalledWith(
            expect.objectContaining({
                success: true,
                data: expect.arrayContaining([
                    expect.objectContaining({ name: 'Diana' }),
                    expect.objectContaining({ name: 'Alice' }),
                ]),
            })
        );

    });

    it('should throw error if member id not found in list', () => {
        service = new RotationService();
        service['lastSelectedMemberIndex'] = 2; // Points to Charlie
        service['lastSelectedMemberId'] = 999;  // Invalid id
        controller = new RotationController(service);

        mockRequest.query = { count: '2' };

        controller.getNext(mockRequest as Request, mockResponse as Response);
        expect(mockResponse.status).toHaveBeenCalledWith(500);
        expect(mockResponse.json).toHaveBeenCalledWith(
            expect.objectContaining({
                success: false,
                message: 'not found lastMemberId in members list'
            })
        );

    });
});
