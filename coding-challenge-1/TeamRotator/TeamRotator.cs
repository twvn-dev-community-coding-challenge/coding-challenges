
namespace TeamRotator
{
    public class TeamRotator
    {
        private readonly LinkedList<Member> _members = new LinkedList<Member>();
    
        public void AddMember(int id, string name)
        {
            var member = new Member(id, name);
            _members.AddLast(member);
        }

        public IEnumerable<Member> GetNext(int count)
        {
            var result = new List<Member>();
            for (int i = 0; i < count; i++)
            {
                if (_members.Count == 0) break;

                var currentMember = _members.First.Value;

                _members.RemoveFirst();
                _members.AddLast(currentMember);

                result.Add(currentMember);
            }

            return result;
        }
    }
}
