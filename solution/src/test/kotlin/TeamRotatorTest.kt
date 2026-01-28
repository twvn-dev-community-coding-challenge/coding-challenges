import org.junit.jupiter.api.Nested
import org.junit.jupiter.api.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class TeamRotatorTest {
    @Nested
    inner class `rotate one` {
        @Test
        fun `team rotator can have a list of 1 member`() {
            val teamRotator = TeamRotator(listOf(Member("AnhLe")));

            assertEquals(
                teamRotator.memberList(),
                listOf(Member("AnhLe"))
            )
        }

        @Test
        fun `team rotator can have a list of 2 members`() {
            val teamRotator = TeamRotator(
                listOf(
                    Member("AnhLe"),
                    Member("Nam")
                )
            );

            assertEquals(
                teamRotator.memberList(),
                listOf(Member("AnhLe"), Member("Nam"))
            );
        }

        @Test
        fun `team rotator cannot have a list of no member`() {
            val listOfNoMember = listOf<Member>();

            val error = assertFailsWith<IllegalArgumentException>(
                block = {
                    val teamRotator = TeamRotator(listOfNoMember)
                }
            )
            assertEquals("Cannot rotate list of no member", error.message)
        }

        @Test
        fun `rotate first time return first of list`() {
            val teamRotator = TeamRotator(
                listOf(
                    Member("AnhLe"),
                    Member("Nam")
                )
            );

            val member: Member = teamRotator.rotate()
            assertEquals(Member("AnhLe"), member)

        }

        @Test
        fun `rotate 2nd time return 2nd of list`() {
            val membersList = listOf(
                Member("AnhLe"),
                Member("Nam")
            );
            val teamRotator = TeamRotator(membersList);

            teamRotator.rotate()
            assertEquals(Member("Nam"), teamRotator.rotate())
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
            assertEquals(Member("Martin"), teamRotator.rotate())
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
            assertEquals(Member("Bob"), teamRotator.rotate())
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
            assertEquals(Member("Bob"), teamRotator.rotate())
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
            val teamRotator = TeamRotator(
                listOf(
                    Member("AnhLe"),
                    Member("Hang")
                )
            );
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
            val teamRotator = TeamRotator(
                listOf(
                    Member("AnhLe"),
                    Member("Hang")
                )
            );
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

        @Test
        fun `list of 1 rotate N members should track last selected member`() {
            val teamRotator = TeamRotator(
                listOf(
                    Member("AnhLe"),
                    Member("Nam"),
                    Member("Hang"),
                    Member("Bob"),
                    Member("Martin"),
                )
            );
            teamRotator.rotateNMembers(3)
            assertEquals(
                Member("Hang"),
                teamRotator.lastSelectedMember()
            )
        }

        @Test
        fun `list of many rotate N members should track last selected member`() {
            val teamRotator = TeamRotator(
                listOf(
                    Member("AnhLe"),
                    Member("Nam"),
                    Member("Hang"),
                    Member("Bob"),
                    Member("Martin"),
                )
            );
            teamRotator.rotateNMembers(3)
            teamRotator.rotateNMembers(4)
            assertEquals(
                Member("Nam"),
                teamRotator.lastSelectedMember()
            )
        }
    }

    @Nested
    inner class `rotate n members` {
        @Test
        fun `rotate 2 members first time return first 2 members in the list`() {
            val teamRotator = TeamRotator(
                listOf(
                    Member("AnhLe"),
                    Member("Nam"),
                    Member("Hang"),
                    Member("Bob"),
                    Member("Martin"),
                )
            );

            assertEquals(
                listOf(
                    Member("AnhLe"),
                    Member("Nam"),
                ),
                teamRotator.rotateNMembers(2)
            )
        }

        @Test
        fun `rotate multiple times not over the list return correct members in the list`() {
            val teamRotator = TeamRotator(
                listOf(
                    Member("AnhLe"),
                    Member("Nam"),
                    Member("Hang"),
                    Member("Bob"),
                    Member("Martin"),
                )
            );

            teamRotator.rotateNMembers(2)

            assertEquals(
                listOf(
                    Member("Hang"),
                    Member("Bob"),
                ),
                teamRotator.rotateNMembers(2)
            )
        }

        @Test
        fun `rotate many times can return the list with repetition`() {
            val teamRotator = TeamRotator(
                listOf(
                    Member("AnhLe"),
                    Member("Nam"),
                    Member("Hang"),
                    Member("Bob"),
                    Member("Martin"),
                )
            );

            teamRotator.rotateNMembers(3)
            assertEquals(
                listOf(
                    Member("Bob"),
                    Member("Martin"),
                    Member("AnhLe"),
                ),
                teamRotator.rotateNMembers(3)
            )
        }

        @Test
        fun `rotate n members where n is larger than list size return repetition`() {
            val teamRotator = TeamRotator(
                listOf(
                    Member("AnhLe"),
                    Member("Nam"),
                    Member("Hang"),
                )
            );

            assertEquals(
                listOf(
                    Member("AnhLe"),
                    Member("Nam"),
                    Member("Hang"),
                    Member("AnhLe"),
                ),
                teamRotator.rotateNMembers(4)
            )
        }
    }

    @Nested
    inner class `inactive member` {
        @Test
        fun `can mark inactive member`() {
            val teamRotator = TeamRotator(
                listOf(
                    Member("AnhLe"),
                    Member("Nam")
                )
            );

            teamRotator.markMemberInactiveByName("Nam")
            assertFalse(teamRotator.isMemberActive("Nam"))
            assertTrue(teamRotator.isMemberActive("AnhLe"))
        }

        @Test
        fun `rotation of one should skip inactive member`() {
            val teamRotator = TeamRotator(
                listOf(
                    Member("AnhLe"),
                    Member("Nam")
                )
            );
            teamRotator.markMemberInactiveByName("AnhLe")
            assertEquals(
                Member("Nam"),
                teamRotator.rotate()
            )
        }

        @Test
        fun `mark inactive member however member not found should throw error`() {
            val teamRotator = TeamRotator(
                listOf(
                    Member("AnhLe"),
                    Member("Nam")
                )
            );

            val error = assertFailsWith<MemberNotFoundException>(
                block = {
                    teamRotator.markMemberInactiveByName("unknown")
                }
            )
            assertEquals("Member not found with name: unknown", error.message)
        }

        @Test
        fun `isMemberActive however member not found should throw error`() {
            val teamRotator = TeamRotator(
                listOf(
                    Member("AnhLe"),
                    Member("Nam")
                )
            );

            val error = assertFailsWith<MemberNotFoundException>(
                block = {
                    teamRotator.isMemberActive("unknown")
                }
            )
            assertEquals("Member not found with name: unknown", error.message)
        }

    }
}