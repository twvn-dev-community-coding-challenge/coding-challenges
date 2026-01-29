import { Member } from '../utils/types'

/**
 * Default team members
 * Centralized member list for easy maintenance
 */
export const DEFAULT_MEMBERS: Member[] = [
    { id: 1, name: 'Alice', isActive: true },
    { id: 2, name: 'Bob', isActive: false },
    { id: 3, name: 'Charlie', isActive: false },
    { id: 4, name: 'Diana', isActive: false },
];

export const getMemberlist = (): Member[] => {
    return DEFAULT_MEMBERS;
}