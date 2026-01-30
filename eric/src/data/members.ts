import { Member } from '../utils/types'

/**
 * Default team members
 * Centralized member list for easy maintenance
 */
export const DEFAULT_MEMBERS: Member[] = [
    { id: 1, name: 'Alice', isActive: true },
    { id: 2, name: 'Bob', isActive: true },
    { id: 3, name: 'Charlie', isActive: true },
    { id: 4, name: 'Diana', isActive: true },
];

export const getMemberlist = (): Member[] => {
    return DEFAULT_MEMBERS;
}