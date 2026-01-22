import org.junit.jupiter.api.Test
import kotlin.test.assertEquals

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
}