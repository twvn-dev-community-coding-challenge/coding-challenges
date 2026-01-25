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
        lastSelectedIndex++
        val result = membersList[lastSelectedIndex]
        return result
    }

    fun lastSelectedMember(): Member {
        return membersList[lastSelectedIndex];
    }

}
