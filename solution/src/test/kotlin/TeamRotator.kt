class TeamRotator {

    val membersList: List<Member>
    private var lastSelectedIndex: Int = -1;

    constructor(membersList: List<Member>) {
        if (membersList.isEmpty()) throw ListOfMemberCannotBeEmptyException()
        ensureNameNotDuplicate(membersList.map{ it.fullName }.toList())
        this.membersList = membersList
    }

    constructor(vararg memberNameList: String): this(memberNameList.map { Member(it) })

    private fun ensureNameNotDuplicate(memberNameList: List<String>) {
        val seen = mutableSetOf<String>()
        for (name in memberNameList) {
            if (!seen.add(name)) {
                throw ListContainDuplicatedNameException(name)
            }
        }
    }

    fun memberList(): List<Member> {
        return membersList;
    }

    fun rotate(): Member {
        rotateLastSelectedIndex()
        val member = membersList[lastSelectedIndex]
        if (!member.isActive())
            rotateLastSelectedIndex()
        return membersList[lastSelectedIndex]
    }

    private fun rotateLastSelectedIndex() {
        lastSelectedIndex++
        if (lastSelectedIndex >= membersList.size) lastSelectedIndex = 0
    }

    fun lastSelectedMember(): Member {
        return membersList[lastSelectedIndex];
    }

    fun rotateNMembers(n: Int): List<Member> {
        val result = mutableListOf<Member>();
        for (i in 1..n) {
            val member = rotate();
            result.add(member);
        }
        return result;
    }

    fun markMemberInactiveByName(memberName: String) {
        val member = membersList.firstOrNull { it.fullName.compareTo(memberName, true) == 0 }
        if (member == null) throw MemberNotFoundException(memberName)
        member.deactivate()
    }

    fun isMemberActive(memberName: String): Boolean {
        val member = membersList.firstOrNull {
            it.fullName.equalsIgnoreCase(memberName)
        }
        if (member == null) throw MemberNotFoundException(memberName)
        return member.isActive()
    }

    fun markMemberActiveByName(memberName: String) {
        val member = membersList.firstOrNull { it.fullName.compareTo(memberName, true) == 0 }
        if (member == null) throw MemberNotFoundException(memberName)
        member.activate()
    }
}

