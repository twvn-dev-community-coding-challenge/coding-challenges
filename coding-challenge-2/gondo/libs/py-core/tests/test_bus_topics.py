from py_core.bus.topics import topic_to_subject


def test_topic_to_subject() -> None:
    assert topic_to_subject("sms.dispatch.requested") == "gondo.sms.dispatch.requested"
    assert topic_to_subject(".foo.bar") == "gondo.foo.bar"
