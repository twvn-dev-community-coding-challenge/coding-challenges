import org.junit.jupiter.api.Nested
import org.junit.jupiter.api.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith

class TeamRotatorTest {
    @Nested
    inner class `rotate one` {
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
                listOf(Member("AnhLe"), Member("Nam"))
            );
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

        @Test
        fun `rotate first time return first of list`() {
            val membersList = listOf(Member("AnhLe"), Member("Nam"));
            val teamRotator = TeamRotator(membersList);

            val member: Member = teamRotator.rotate()
            assertEquals(Member("AnhLe"), member)

        }

        @Test
        fun `rotate 2nd time return 2nd of list`() {
            val membersList = listOf(Member("AnhLe"), Member("Nam"));
            val teamRotator = TeamRotator(membersList);

            teamRotator.rotate()
            val member: Member = teamRotator.rotate()
            assertEquals(Member("Nam"), member)
        }

        @Test
        fun `can rotate to the end of list`() {
            val membersList = listOf(
                Member("AnhLe"),
                Member("Nam"),
                Member("Hang"),
                Member("Bob"),
                Member("Martin"),
            );
            val teamRotator = TeamRotator(membersList);

            teamRotator.rotate()
            teamRotator.rotate()
            teamRotator.rotate()
            teamRotator.rotate()
            val member: Member = teamRotator.rotate()
            assertEquals(Member("Martin"), member)
        }

        @Test
        fun `can rotate repeat the list of one`() {
            val teamRotator = TeamRotator(
                listOf(
                    Member("Bob"),
                )
            );

            teamRotator.rotate()
            teamRotator.rotate()
            val member: Member = teamRotator.rotate()
            assertEquals(Member("Bob"), member)
        }

        @Test
        fun `can rotate repeat the list of many`() {
            val teamRotator = TeamRotator(
                listOf(
                    Member("Bob"),
                    Member("Martin"),
                )
            );

            teamRotator.rotate()
            teamRotator.rotate()
            val member: Member = teamRotator.rotate()
            assertEquals(Member("Bob"), member)
        }
    }

    @Nested
    inner class `Track last selected member` {

        @Test
        fun `list of one rotate 1 remember should track last selected member`() {
            val membersList = listOf(Member("AnhLe"));
            val teamRotator = TeamRotator(membersList);
            teamRotator.rotate()
            assertEquals(
                Member("AnhLe"),
                teamRotator.lastSelectedMember()
            )
        }

        @Test
        fun `list of 2 rotate 1 twice remember should track last selected member`() {
            val teamRotator = TeamRotator(listOf(
                Member("AnhLe"),
                Member("Hang")
            ));
            teamRotator.rotate()
            assertEquals(
                Member("AnhLe"),
                teamRotator.lastSelectedMember()
            )

            teamRotator.rotate()
            assertEquals(
                Member("Hang"),
                teamRotator.lastSelectedMember()
            )
        }
        @Test
        fun `list reset can track last selected member`() {
            val teamRotator = TeamRotator(listOf(
                Member("AnhLe"),
                Member("Hang")
            ));
            teamRotator.rotate()
            teamRotator.rotate()
            teamRotator.rotate()
            assertEquals(
                Member("AnhLe"),
                teamRotator.lastSelectedMember()
            )
            teamRotator.rotate()
            assertEquals(
                Member("Hang"),
                teamRotator.lastSelectedMember()
            )
        }
    }

    @Nested
    inner class `rotate n members` {
        @Test
        fun `rotate 2 members first time return first 2 members in the list`() {
            val membersList = listOf(
                Member("AnhLe"),
                Member("Nam"),
                Member("Hang"),
                Member("Bob"),
                Member("Martin"),
            );
            val teamRotator = TeamRotator(membersList);

            val rotatedMembers: List<Member> = teamRotator.rotateNMembers(2);

            assertEquals(
                listOf(
                    Member("AnhLe"),
                    Member("Nam"),
                ),
                rotatedMembers
            )
        }
    }
}