class TeamRotator {

    val membersList: List<Member>

    constructor(membersList: List<Member>) {
        if(membersList.isEmpty()) throw IllegalArgumentException("Cannot rotate list of no member")
        this.membersList = membersList
    }

    fun memberList(): List<Member> {
        return membersList;
    }

}
