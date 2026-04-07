import org.example.Member
import org.example.NoActiveMembersAvailable
import org.example.TeamRotator
import org.junit.jupiter.api.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith

class AcceptanceTest {
    @Test
    fun `Example 1 - basic rotation`() {
        val teamRotator = TeamRotator(
            "Alice", "Bob", "Charlie", "Diana"
        )

        assertEquals(Member("Alice"), teamRotator.rotate())
        assertEquals(Member("Bob"), teamRotator.rotate())
        assertEquals(Member("Charlie"), teamRotator.rotate())
        assertEquals(Member("Diana"), teamRotator.rotate())
        assertEquals(Member("Alice"), teamRotator.rotate())
        assertEquals(Member("Bob"), teamRotator.rotate())
    }

    @Test
    fun `Example 2 - no immediate repetition`() {
        val teamRotator = TeamRotator(
            "Alice", "Bob", "Charlie"
        )
        teamRotator.rotate()
        assertEquals(Member("Bob"), teamRotator.rotate())
        assertEquals(Member("Charlie"), teamRotator.rotate())
        assertEquals(Member("Alice"), teamRotator.rotate())
        assertEquals(Member("Bob"), teamRotator.rotate())
    }

    @Test
    fun `Example 3 - skipping inactive members`() {
        val teamRotator = TeamRotator(
            "Alice", "Bob", "Charlie", "Diana"
        )
        teamRotator.markMemberInactiveByName("Bob")
        assertEquals(Member("Alice"), teamRotator.rotate())
        assertEquals(Member("Charlie"), teamRotator.rotate())
        assertEquals(Member("Diana"), teamRotator.rotate())
        assertEquals(Member("Alice"), teamRotator.rotate())
        assertEquals(Member("Charlie"), teamRotator.rotate())
    }

    @Test
    fun `Example 4 - requesting next N member`() {
        val teamRotator = TeamRotator(
            "Alice", "Bob", "Charlie", "Diana"
        )
        assertEquals(listOf(Member("Alice"), Member("Bob")), teamRotator.rotateNMembers(2))
        assertEquals(listOf(Member("Charlie"), Member("Diana")), teamRotator.rotateNMembers(2))
        assertEquals(listOf(Member("Alice"), Member("Bob")), teamRotator.rotateNMembers(2))
    }

    @Test
    fun `Example 5 - edge case only one active member`() {
        val teamRotator = TeamRotator(
            "Alice", "Bob", "Charlie"
        )

        teamRotator.markMemberInactiveByName("Bob")
        teamRotator.markMemberInactiveByName("Charlie")
        assertEquals(Member("Alice"), teamRotator.rotate())
        assertEquals(Member("Alice"), teamRotator.rotate())
        assertEquals(Member("Alice"), teamRotator.rotate())
    }

    @Test
    fun `Example 6 - all member is inactive`() {
        val teamRotator = TeamRotator(
            "Alice", "Bob", "Charlie"
        )

        teamRotator.markMemberInactiveByName("Alice")
        teamRotator.markMemberInactiveByName("Bob")
        teamRotator.markMemberInactiveByName("Charlie")
        val error = assertFailsWith<NoActiveMembersAvailable>(
            block = {
                teamRotator.rotate()
            }
        )
        assertEquals("No active members available", error.message)
    }
}