namespace TeamRotator
{
    public class Member
    {
        public int Id { get; }
        public string Name { get; }

        public Member(int id, string name)
        {
            Id = id;
            Name = name;
        }
    }
}

