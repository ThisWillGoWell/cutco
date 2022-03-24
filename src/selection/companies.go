package selection

type Company struct {
	SelectInfo    bool
	SelectHistory bool
	Shares        *Share
	Users         *User
	Transaction   *Transaction
}

type Share struct {
	SelectInfo  bool
	Holder      *User
	Company     *Company
	Transaction *Transaction
}

type ValueHistory struct {
}

type Transaction struct {
	User       *User
	SelectInfo bool
}
