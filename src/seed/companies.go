package seed

import (
	"stock-simulator-serverless/src/models"
	"time"
)

var UserList = []*models.PrivateUserStruct{
	{
		Login:    "will",
		Password: "password",
		User: &models.UserStruct{
			Name:         "thisWillGoWell",
			Description:  "hello, its me",
			InvestorType: "Hacker",
		},
	}, {
		Login:    "chunt",
		Password: "password",
		User: &models.UserStruct{
			Name:         "Scroomp",
			Description:  "Hello! Its Connor Hunt.",
			InvestorType: "Raging Bull",
		},
	}, {
		Login:    "jake",
		Password: "password",
		User: &models.UserStruct{
			Name:         "Actex",
			Description:  "yes, the one true tex. Welcome to my house.",
			InvestorType: "tactical",
		},
	}, {
		Login:    "morph",
		Password: "password",
		User: &models.UserStruct{
			Name:         "Morpheus",
			Description:  "hello, its me",
			InvestorType: "Raging Bull",
		},
	}, {
		Login:    "raven",
		Password: "password",
		User: &models.UserStruct{
			Name:         "Raven",
			Description:  "hello, its me",
			InvestorType: "😈",
		},
	}, {
		Login:    "zpanduh",
		Password: "password",
		User: &models.UserStruct{
			Name:         "Business Owner",
			Description:  "",
			InvestorType: "meme",
		},
	},
}

func randomTime() time.Time {
	return time.Now()
}
func randomOpenShares() int {
	return 10000
}

func randomValue() int {
	return 420
}

var CompanyList = []*models.CompanyStruct{
	{
		Symbol:      "CHUNT",
		Name:        "Chunt Inc.",
		Description: "American multinational oil and gas corporation.",
		Type:        "entertainment",
		CreatedAt:   randomTime(),
		OpenShares:  randomOpenShares(),
		Value:       randomValue(),
	},
	{
		Symbol:      "GUZ",
		Name:        "Guzman Talent Agency",
		Description: "our CEO Joe Estevez will get you the best gigs.",
		Type:        "entertainment",
		CreatedAt:   randomTime(),
		OpenShares:  randomOpenShares(),
		Value:       randomValue(),
	}, {
		Symbol:      "TLNX",
		Name:        "Telnyx LLC",
		Description: "Universal communication made simple though our #1 Customer service platform.",
		Type:        "communications",
		CreatedAt:   randomTime(),
		OpenShares:  randomOpenShares(),
		Value:       randomValue(),
	}, {
		Symbol:      "RPISD",
		Name:        "Randy Pissman Decals",
		Description: "Come here for all your decal needs. We have the best fake flames in town!",
		Type:        "store",
		CreatedAt:   randomTime(),
		OpenShares:  randomOpenShares(),
		Value:       randomValue(),
	}, {
		Symbol:      "OSRS",
		Name:        "Runescape",
		Description: "Where everyone is a bot",
		Type:        "game",
		CreatedAt:   randomTime(),
		OpenShares:  randomOpenShares(),
		Value:       randomValue(),
	}, {
		Symbol:      "OVRW",
		Name:        "Overwatch",
		Description: "What you really want in PvE",
		Type:        "game",
		CreatedAt:   randomTime(),
		OpenShares:  randomOpenShares(),
		Value:       randomValue(),
	}, {
		Symbol:      "GWEN",
		Name:        "Grandmas's Kitchen",
		Description: "Best Grandma's Pizza in town",
		Type:        "food",
		CreatedAt:   randomTime(),
		OpenShares:  randomOpenShares(),
		Value:       randomValue(),
	}, {
		Symbol:      "CANDLE",
		Name:        "Serenity by Jane",
		Description: "Smells so good that you can work where you create.",
		Type:        "store",
		CreatedAt:   randomTime(),
		OpenShares:  randomOpenShares(),
		Value:       randomValue(),
	}, {
		Symbol:      "PAPER",
		Name:        "Dunder Mifflin",
		Description: "The company was founded by Robert Dunder and Robert Mifflin in 1949, where they supplied metal brackets. Eventually, the company started selling paper and opened several branches across the Northeastern United States.",
		Type:        "",
		CreatedAt:   randomTime(),
		OpenShares:  randomOpenShares(),
		Value:       randomValue(),
	}, {
		Symbol:      "VANCE",
		Name:        "Vance Refrigeration",
		Description: "Vance Refrigeration has been serving the Scranton area for over 40 years. They service both residential and commercial refrigerators, and their work comes with a one-year guarantee.",
		Type:        "industry",
		CreatedAt:   randomTime(),
		OpenShares:  randomOpenShares(),
		Value:       randomValue(),
	}, {
		Symbol:      "NBC",
		Name:        "Kabletown",
		Description: "Kabletown is an American telecommunications conglomerate headquartered in Philadelphia, Pennsylvania.",
		Type:        "communications",
		CreatedAt:   randomTime(),
		OpenShares:  randomOpenShares(),
		Value:       randomValue(),
	}, {
		Symbol:      "WONK",
		Name:        "Willy Wonka's Chocolate",
		Description: "We make candy. We do not use slave labor.",
		Type:        "industry",
		CreatedAt:   randomTime(),
		OpenShares:  randomOpenShares(),
		Value:       randomValue(),
	}, {
		Symbol:      "CITY",
		Name:        "City Wok",
		Description: "City Wok is a Chinese restaurant and small commercial airline. It is run by Tuong Lu Kim.",
		Type:        "food",
		CreatedAt:   randomTime(),
		OpenShares:  randomOpenShares(),
		Value:       randomValue(),
	}, {
		Symbol:      "WIPE",
		Name:        "Hawthorne Industries",
		Description: "I got my body, got my lips, got a pocket full of Hawthornes, p-p-p-pocket full of Hawthornes.",
		Type:        "industry",
		CreatedAt:   randomTime(),
		OpenShares:  randomOpenShares(),
		Value:       randomValue(),
	}, {
		Symbol:      "SAND",
		Name:        "Shirley's Sandwiches",
		Description: "Shirley's Sandwiches can offer students more food for less money and provide the school with a higher percentage of the profits.",
		Type:        "food",
		CreatedAt:   randomTime(),
		OpenShares:  randomOpenShares(),
		Value:       randomValue(),
	}, {
		Symbol:      "SCOTT",
		Name:        "Michal Scott Paper Company",
		Description: "The people person's paper people.",
		Type:        "supply-chain",
		CreatedAt:   randomTime(),
		OpenShares:  randomOpenShares(),
		Value:       randomValue(),
	}, {
		Symbol:      "SIC",
		Name:        "Church Of Scientology",
		Description: "You dont want to know",
		Type:        "religion",
		CreatedAt:   randomTime(),
		OpenShares:  randomOpenShares(),
		Value:       randomValue(),
	}, {
		Symbol:      "GE",
		Name:        "General Electric",
		Description: "We used to make things you know.",
		Type:        "industry",
		CreatedAt:   randomTime(),
		OpenShares:  randomOpenShares(),
		Value:       randomValue(),
	}, {
		Symbol:      "FANCY",
		Name:        "Shoe La La",
		Description: "Shoes for the special occasions in a man's life.",
		Type:        "store",
		CreatedAt:   randomTime(),
		OpenShares:  randomOpenShares(),
		Value:       randomValue(),
	}, {
		Symbol:      "WOLF",
		Name:        "Franks Fluids LLC",
		Description: "Creates hit beverages like Wolf Cola and Fight Milk. Located in Philadelphia, operated out of Panama.",
		Type:        "food",
		CreatedAt:   randomTime(),
		OpenShares:  randomOpenShares(),
		Value:       randomValue(),
	},
}

func init() {
	for i, u := range UserList {
		u.User.Company = CompanyList[i]
		CompanyList[i].Owner = u.User
	}
}
