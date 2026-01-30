/**
 * Represents a team member
 */
export interface Member {
    id: number;
    name: string;
    isActive: boolean;
}


export type ResultSuccess<T> = {
    success: true;
    data: T;
}

export type ResultFailure = {
    success: false;
    message: string;
}

export type Result<T> = ResultSuccess<T> | ResultFailure;
