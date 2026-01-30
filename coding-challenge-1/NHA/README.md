### Starting the NHA Coding Challenge
To run the test, you need to follow these steps:
- `cd coding-challenge-1/NHA`
- `gradle clean test`

To run the project, run:
- `gradle clean build`

The project is built using Gradle and Java 21. Make sure you have both installed on your machine before running the commands.

### Design Decisions and Trade-offs:

1. Language Choice: Kotlin
   - Reason: We heard that Kotlin is the new coding language based on Java, so that we decided to use it for experimental purposes.
   - Trade-off: Slight learning curve for developers unfamiliar with Kotlin.

2. We built the algorithm with TDD instead of using complex data structures.
   - Reason: The code will be easier to maintain and readable
   - Trade-off: Might not be the most optimized solution for very large datasets. However, actually, this problem is applied for the small team so that the performance is acceptable.

### With more time, we would have:
   - Build a simple REST API for demo version