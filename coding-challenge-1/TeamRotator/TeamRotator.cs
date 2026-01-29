namespace TeamRotator;

public class TeamRotator
{
    private readonly List<Member> _members = new();
    private int? _lastSelectedId;

    public void AddMember(int id, string name, bool isActive = true)
    {
        _members.Add(new Member(id, name, isActive));
    }

    public void SetManualLastSelected(int id)
    {
        _lastSelectedId = id;
    }

    public List<Member> GetNext(int count = 1)
    {
        var result = new List<Member>();

        for (var i = 0; i < count; i++)
        {
            var nextMember = FindNextValidMember();
            result.Add(nextMember);
            _lastSelectedId = nextMember.Id;
        }

        return result;
    }

    private Member FindNextValidMember()
    {
        var activeMembers = _members.Where(m => m.IsActive).ToList();

        if (!activeMembers.Any()) throw new InvalidOperationException("No active members available");

        if (activeMembers.Count == 1) return activeMembers.First();

        var lastIndex = _members.FindIndex(m => m.Id == _lastSelectedId);

        for (var i = 1; i <= _members.Count; i++)
        {
            var currentIndex = (lastIndex + i) % _members.Count;
            var candidate = _members[currentIndex];

            if (candidate.IsActive && candidate.Id != _lastSelectedId) return candidate;
        }

        return activeMembers.First();
    }
}