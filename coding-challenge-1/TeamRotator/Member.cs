namespace TeamRotator;

public class Member
{
    public Member(int id, string name, bool isActive = true)
    {
        Id = id;
        Name = name;
        IsActive = isActive;
    }

    public int Id { get; }
    public string Name { get; }
    public bool IsActive { get; }
}