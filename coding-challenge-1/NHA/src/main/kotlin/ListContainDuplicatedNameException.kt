package org.example

class ListContainDuplicatedNameException(memberName: String) : RuntimeException("Name duplicated: $memberName")
