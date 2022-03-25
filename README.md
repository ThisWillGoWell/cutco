Camper
- Name: string
- Profile Picture: string
- Badges: []Badges
- Type: CamperType
- Awards: []Award
- TotalPoints: Int
- GiftPoints: Int
- JoinedOn: Timestamp
- AtCamp: bool


CamperType: [Admin, Camper, Staff, Parent, Guest]

Badge
- Name: string
- Picture: string
- Points: Int
- Description: string
- Creator: Camper
- Owners: []Camper
- CreatedAt: Timestamp
- AwardedAt: Timestamp


Create a badge to be rewarded in the future
CreateMeritBadge()

Award a Badge that has been created to one or many campers
AwardBadge()

Award:
- From: Camper
- To: Camper
- Message: string
- Timestamp: ts
- Badge: ?Badge
- PointValue: Int

AwardPoints():
DeductPoints(): 


Chat



