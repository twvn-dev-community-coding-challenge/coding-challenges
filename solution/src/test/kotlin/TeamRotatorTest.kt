import org.junit.jupiter.api.Test

class TeamRotatorTest {
    @Test
    fun `team rotator can have a list of 1 member`() {
        val membersList = listOf(Member("AnhLe"));
        val teamRotator = TeamRotator(membersList);

        teamRotator.memberList() == listOf(Member("AnhLe"))
    }

    @Test
    fun `team rotator can have a list of 2 members`() {
        val membersList = listOf(Member("AnhLe"), Member("Nam"));
        val teamRotator = TeamRotator(membersList);

        teamRotator.memberList() == listOf(Member("AnhLe"), Member("Nam"))
    }
}