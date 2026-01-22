import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith

class TeamRotatorTest {
    @Test
    fun `team rotator can have a list of 1 member`() {
        val membersList = listOf(Member("AnhLe"));
        val teamRotator = TeamRotator(membersList);

        assertEquals(
            teamRotator.memberList(),
            listOf(Member("AnhLe"))
        )
    }

    @Test
    fun `team rotator can have a list of 2 members`() {
        val membersList = listOf(Member("AnhLe"), Member("Nam"));
        val teamRotator = TeamRotator(membersList);

        assertEquals(
            teamRotator.memberList(),
            listOf(Member("AnhLe"), Member("Nam")));
    }

    @Test
    fun `team rotator cannot have a list of no member`() {
        val membersList = listOf<Member>();

        val error = assertFailsWith<IllegalArgumentException>(
            block = {
                val teamRotator = TeamRotator(membersList)
            }
        )
        assertEquals("Cannot rotate list of no member", error.message)
    }
}