BRANDING, marketing, LOCKED DOWN!!!!

Me
- Camper: Camper
- Login: string


CheckIn()
Check the camper into camp, Welcome!
CheckOut()
Check the camper out of camp.

Refer(email: ?string, phoneNumber: ?string)
Logout()
DeleteMe()


Camper
- ID: string
- Name: string
- Profile Picture: string
- Badges: []Badges
- Type: CamperType
- Awards: []Award
- TotalPoints: Int
- JoinedOn: Timestamp
- AtCamp: bool
- DaysAtCamp: Int
- StayHistory:
- Votes: []BadgeVotes
- ParentPoints: int128


Get all campers registered 
Campers()
Get all campers currently at camp
CampersAtCamp()

CamperType: [Admin, Camper, Staff, Parent, Guest]

Badge
- Name: string
- Type: BadgeType
- Picture: string
- Points: Int
- Description: string
- Creator: Camper
- Owners: []Camper
- CreatedAt: Timestamp
- AwardedAt: Timestamp
- Likes []Camper

BadgeType: [Merit, Community]

Create a badge to be rewarded in the future
Upload Photo
CreateBadge()

Award a Badge that has been created to one or many campers
AwardBadge()

BadgeVotes:
- Decision: DecisionType
- Voter: Camper 
- Timestamp
- Badge: CommunityBadge

DecisionType: [Yes, ChangePointsHigher, ChangePointsLower, No]

CommunityBadge
A Badge with people vote yes or no, and if its worth that amount of points
Admin decides yes or no (after voting), can adjust points
- Badge: Badge
- Votes: []BadgeVotes
- ApprovedBy: Camper


Award:
- From: Camper
- To: Camper
- Message: string
- Timestamp: ts
- Badge: ?Badge
- PointValue: Int

AwardPoints():
- no more than 500 points? 
- 
DeductPoints(): 
- only parents can deduct parentPoints

Chat?




