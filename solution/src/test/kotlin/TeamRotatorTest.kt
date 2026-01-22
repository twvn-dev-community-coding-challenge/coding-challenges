import org.junit.jupiter.api.Test

class TeamRotatorTest {
    @Test
    fun `team rotator can have a list of members`() {
        val membersList = listOf(Member("AnhLe"), Member("Nam"));
        val teamRotator = TeamRotator(membersList);

        teamRotator.memberList() == listOf(Member("AnhLe"), Member("Nam"))
    }
}