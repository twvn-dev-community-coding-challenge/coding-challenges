import org.junit.jupiter.api.Assertions.*
import org.junit.jupiter.api.Test

class FooTest {
    @Test
    fun `team rotator can have a list of members`() {
        val teamRotator = TeamRotator();
        val membersList = listOf(Member("AnhLe"), Member("Nam"));

        teamRotator.addMembers(membersList);

        teamRotator.memberList().equals(listOf(Member("AnhLe"), Member("Nam")))
    }
}