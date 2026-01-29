using FluentAssertions;
using Xunit;
using System;
using System.Linq;
namespace TeamRotator.Tests
{
    public class TeamRotatorTests
    {
        private readonly TeamRotator _rotator;

        public TeamRotatorTests()
        {
            _rotator = new TeamRotator();
        }

        [Fact]
        public void GetNext_MultipleMembersActive_ReturnsMembersInRoundRobinOrder()
        {
            _rotator.AddMember(1, "Alice");
            _rotator.AddMember(2, "Bob");
            _rotator.AddMember(3, "Charlie");

            var firstCall = _rotator.GetNext(1).First().Name;
            var secondCall = _rotator.GetNext(1).First().Name;
            var thirdCall = _rotator.GetNext(1).First().Name;
            var fourthCall = _rotator.GetNext(1).First().Name;

            firstCall.Should().Be("Alice");
            secondCall.Should().Be("Bob");
            thirdCall.Should().Be("Charlie");
            fourthCall.Should().Be("Alice", "because the rotation should restart");
        }

        [Fact]
        public void GetNext_OneMemberInactive_SkipsInactiveMember()
        {
            _rotator.AddMember(1, "Alice");
            _rotator.AddMember(2, "Bob", isActive: false); 
            _rotator.AddMember(3, "Charlie");

            var firstPick = _rotator.GetNext(1).First().Name;
            var secondPick = _rotator.GetNext(1).First().Name;

            firstPick.Should().Be("Alice");
            secondPick.Should().Be("Charlie", "because Bob is inactive and should be skipped");
        }

        [Fact]
        public void GetNext_LastMemberSetManually_SkipsLastSelectedMember()
        {
            _rotator.AddMember(1, "Alice");
            _rotator.AddMember(2, "Bob");

            _rotator.SetManualLastSelected(1);

            var result = _rotator.GetNext(1).First().Name;

            result.Should().Be("Bob");
        }
    
        [Fact]
        public void GetNext_AllMembersInactive_ThrowsInvalidOperationException()
        {
            _rotator.AddMember(1, "Alice", isActive: false);
            _rotator.AddMember(2, "Bob", isActive: false);

            Action act = () => _rotator.GetNext();

            act.Should().Throw<InvalidOperationException>()
                .WithMessage("No active members available"); 
        }
    
        [Fact]
        public void GetNext_RequestingMultipleMembers_ReturnsCorrectBatchSize()
        {
            _rotator.AddMember(1, "Alice");
            _rotator.AddMember(2, "Bob");
            _rotator.AddMember(3, "Charlie");

            var results = _rotator.GetNext(2); 

            results.Should().HaveCount(2);
            results[0].Name.Should().Be("Alice");
            results[1].Name.Should().Be("Bob");
        }
    }
}
