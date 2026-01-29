data class Member(val fullName: String, private var isActive: Boolean = true) {
    fun deactivate() {
        isActive = false;
    }

    fun activate() {
        isActive = true
    }

    fun isActive(): Boolean {
        return isActive
    }

}
