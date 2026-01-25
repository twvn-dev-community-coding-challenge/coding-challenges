class TeamRotator {

    val membersList: List<Member>
    private var rotateCount: Int = 0;
    constructor(membersList: List<Member>) {
        if(membersList.isEmpty()) throw IllegalArgumentException("Cannot rotate list of no member")
        this.membersList = membersList
    }

    fun memberList(): List<Member> {
        return membersList;
    }

    fun rotate(): Member {
        val result = membersList[rotateCount]
        updateRotationIndex()
        return result
    }

    private fun updateRotationIndex() {
        rotateCount++
        if (rotateCount >= membersList.size) rotateCount = 0
    }

}
