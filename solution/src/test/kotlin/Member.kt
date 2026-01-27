data class Member(val fullName: String, var isActive: Boolean = true) {
    fun deactivate() {
        isActive = false;
    }

}
