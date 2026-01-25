class TeamRotator {

    val membersList: List<Member>
    private var lastSelectedIndex: Int = -1;
    constructor(membersList: List<Member>) {
        if(membersList.isEmpty()) throw IllegalArgumentException("Cannot rotate list of no member")
        this.membersList = membersList
    }

    fun memberList(): List<Member> {
        return membersList;
    }

    fun rotate(): Member {
        rotateLastSelectedIndex()
        val result = membersList[lastSelectedIndex]
        return result
    }

    private fun rotateLastSelectedIndex() {
        lastSelectedIndex++
        if (lastSelectedIndex >= membersList.size) lastSelectedIndex = 0
    }

    fun lastSelectedMember(): Member {
        return membersList[lastSelectedIndex];
    }

}
