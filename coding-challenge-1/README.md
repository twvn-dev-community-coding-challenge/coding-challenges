# HoaNguyen - Coding Challenge 1

## Team Members
- Hoa Nguyen (developer/self)

## How to Run
- Step 1: Install uv
- Step 2: cd into coding-challenge-1/
- Step 3: Run uv sync
- Step 4: Run uv build
- Step 5: pip install dist/<lib_name>.whl
- Step 6: Import and run in notebooks or python code
    - Ref: Refers to samples/samples.ipynb for concrete sample

## Approach and decisions
### Classes
<code>MemberManager</code>
- Responsible for managing the collection of members and handling member-related operations.
- This provides a centralized place to control member state and retrieval logic.

<code>Member</code>
- Represents an individual member and their associated information.
- This class provides a clear structure for identifying and working with a member.

<code>Status</code>
- Extra enum class to identify statuses

### Member storing design
- Members are stored in an in-memory dictionary:
    - Each key represents a unique, incremental member ID
    - Each value is a Member object containing the member’s data
    - Member IDs are controlled internally and always increment
- This design allows:
    - O(1) lookup and updates using dictionary access
    - Simple control over member identity, ordering, and traversal
- Example:
```
members = {
    1: Member(id=1, name="A", is_active="isActive"),
    2: Member(id=2, name="B", is_active="isNotActive")
}
```

### Cyclic algorithm
- To support cyclic member retrieval:
    - An incremental pointer is maintained in memory
    - The number of active members and the total member number is tracked separately
    - The pointer is incremented and wrapped using modulo arithmetic
- This allows iteration to “loop back” to the beginning once the end is reached.
- Example:
```
formula: incremental key % members

member_count = 5

4 % member_count = 4 -> members[4]
5 % member_count = 0 -> members[0]
6 % member_count = 1 -> members[1]
7 % member_count = 2 -> members[2]
8 % member_count = 3 -> members[3]
```

### Tradeoffs
- Pros:
    - Efficient lookup, insert and updates due to dictonary, with member id as "indexing" key, O(1) time complexity for all 3 operations
    - Keeping a seprate incremental key to modulo total members reduce the complexity of managing any state entity
- Cons:
    - The cyclic algorithm depends heavily on how we control the "index" or the member id, which always need to be incremental
    - Since we control the input member id, user experience and flexibility is sacrificed
    - Also, incremental member id means that if cyclic does occur, at worst, the get_next() time complexity would be O(n), where n is the total members
        - Which also means that even if a member is inactive, we still need to traverse sequentially through and filter them out until reaching next valid member
    - Since this a coding challenge, in-memory approach member store might be fine, but it can soon blows up if member dictionary kept increasing



## Challenges Faced
- Coming up with cyclic algorithm as it needs to be effecient in the 3 main operations: search, insert, and update
- In that regards, python dictionary (or hashmap in order language) is the most suitable, due to O(1) lookup, and insert

## What We Learned
- Using uv as project environment management
- Writting and setup unit tests
- Setup packaging

## With More Time, We Would...
- More dynamic user creation, change status
    - Current operations only account for single user approach
- get_next() could implement python generators, to reduce memory usage, because at worst, if we decide to get all members, a list of all members will be created, which doubles the memory usage O(2n) since the original dict of members existed.
- Due to the cons of get_next() mentioned above, if we could find a way to only process active members, without redudant traversal and inactive filtering would be great

## AI Tools Used (if any)
- Chatgpt:
    - Pytest syntax, and test case generations
    - Packaging setup