package org.example

class MemberNotFoundException(memberName: String) : RuntimeException("org.example.Member not found with name: $memberName")
